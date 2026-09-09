package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xenptr/go-projects/caching-proxy/internal/cache"
)

const cacheTTL = 5 * time.Minute

// cachedResponse is what we store in Redis for each upstream response.
type cachedResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       []byte              `json:"body"`
}

// Proxy is a caching reverse proxy.
type Proxy struct {
	origin *url.URL
	cache  *cache.Client
	client *http.Client

	inFlight map[string]*call
	mu       sync.Mutex
}

type call struct {
	done chan struct{}
	resp *cachedResponse
	err  error
}

// New creates a new Proxy that forwards requests to origin and caches
// responses in the provided Redis client.
func New(origin string, c *cache.Client) (*Proxy, error) {
	u, err := url.Parse(origin)
	if err != nil {
		return nil, fmt.Errorf("parsing origin url: %w", err)
	}

	return &Proxy{
		origin:   u,
		cache:    c,
		client:   &http.Client{Timeout: 30 * time.Second},
		inFlight: make(map[string]*call),
	}, nil
}

// ServeHTTP implements http.Handler.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cacheable := isCacheableMethod(r.Method)

	// Non-cacheable requests go directly upstream.
	if !cacheable {
		p.serveDirect(w, r)
		return
	}

	key := cacheKey(r)

	if cached, hit := p.fromCache(r.Context(), key); hit {
		writeResponse(w, cached, "HIT")
		log.Printf("CACHE HIT  %s %s", r.Method, r.URL.Path)
		return
	}

	p.mu.Lock()

	// Check whether another goroutine is already fetching this key
	if existed, exist := p.inFlight[key]; exist {
		p.mu.Unlock()

		// Wait for the leader to finish
		select {
		case <-existed.done:
			if existed.err != nil {
				http.Error(w, "bad gateway", http.StatusBadGateway)
				return
			}

			writeResponse(w, existed.resp, "SHARED")
			log.Printf("CACHE SHARED  %s %s", r.Method, r.URL.Path)
			return

		// or if client's request is cancelled.
		case <-r.Context().Done():
			log.Printf("REQUEST CANCELLED %s %s", r.Method, r.URL.Path)
			return
		}

	}

	// No one is fetching this key.
	// This request becomes the leader.
	c := &call{
		done: make(chan struct{}),
	}

	p.inFlight[key] = c

	p.mu.Unlock()

	// Use an independent context so the shared upstream request
	// continues even if the leader's client disconnects.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	upstream, err := p.fetchUpstream(ctx, r)
	if err != nil {
		p.finishCall(key, c, nil, err)

		http.Error(w, "bad gateway", http.StatusBadGateway)
		return

	}
	defer upstream.Body.Close()

	body, err := io.ReadAll(upstream.Body)
	if err != nil {
		p.finishCall(key, c, nil, err)

		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}

	resp := &cachedResponse{
		StatusCode: upstream.StatusCode,
		Headers:    upstream.Header,
		Body:       body,
	}

	// Only cache successful responses.
	if upstream.StatusCode >= 200 && upstream.StatusCode < 300 {
		p.storeCache(r.Context(), key, resp)
	}

	p.finishCall(key, c, resp, nil)

	writeResponse(w, resp, "MISS")
	log.Printf("CACHE MISS %s %s -> %d", r.Method, r.URL.Path, upstream.StatusCode)
}

func (p *Proxy) serveDirect(w http.ResponseWriter, r *http.Request) {
	upstream, err := p.fetchUpstream(r.Context(), r)
	if err != nil {
		log.Printf("upstream error: %v", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return

	}
	defer upstream.Body.Close()

	// Forward upstream headers
	for k, vv := range upstream.Header {
		if isHopByHop(k) {
			continue
		}

		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.Header().Set("X-Cache", "BYPASS")
	w.WriteHeader(upstream.StatusCode)

	_, _ = io.Copy(w, upstream.Body)

	log.Printf("CACHE BYPASS %s %s -> %d", r.Method, r.URL.Path, upstream.StatusCode)
}

func (p *Proxy) finishCall(key string, c *call, resp *cachedResponse, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.inFlight, key)

	c.resp = resp
	c.err = err

	// Wake everyone waiting
	close(c.done)
}

// ClearAll deletes every key in the Redis DB.
// This is intentionally destructive — call only from the --clear-cache path.
func (p *Proxy) ClearAll(ctx context.Context) error {
	return p.cache.Client.FlushDB(ctx).Err()
}

// fetchUpstream forwards the incoming request to the origin server.
func (p *Proxy) fetchUpstream(ctx context.Context, r *http.Request) (*http.Response, error) {
	target := *p.origin
	target.Path = strings.TrimRight(target.Path, "/") + r.URL.Path
	target.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(ctx, r.Method, target.String(), r.Body)
	if err != nil {
		return nil, fmt.Errorf("building upstream request: %w", err)
	}

	// Forward request headers, skipping hop-by-hop headers.
	for k, vv := range r.Header {
		if isHopByHop(k) {
			continue
		}
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("X-Forwarded-For", r.RemoteAddr)

	return p.client.Do(req)
}

func (p *Proxy) fromCache(ctx context.Context, key string) (*cachedResponse, bool) {
	data, err := p.cache.Client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var resp cachedResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, false
	}

	return &resp, true
}

func (p *Proxy) storeCache(ctx context.Context, key string, resp *cachedResponse) {
	data, err := json.Marshal(resp)
	if err != nil {
		log.Printf("marshalling cache entry: %v", err)
		return
	}

	if err := p.cache.Client.Set(ctx, key, data, cacheTTL).Err(); err != nil {
		log.Printf("storing cache entry: %v", err)
	}
}

func writeResponse(w http.ResponseWriter, resp *cachedResponse, cacheStatus string) {
	for k, vv := range resp.Headers {
		if isHopByHop(k) {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	w.Header().Set("X-Cache", cacheStatus)

	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(resp.Body)
}

// cacheKey builds a Redis key from the request method + URL.
func cacheKey(r *http.Request) string {
	return fmt.Sprintf("proxy:%s:%s", r.Method, r.URL.RequestURI())
}

// isHopByHop reports whether the header is a hop-by-hop header that should
// not be forwarded or cached.
func isHopByHop(h string) bool {
	switch strings.ToLower(h) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization",
		"te", "trailers", "transfer-encoding", "upgrade":
		return true
	}
	return false
}

func isCacheableMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead
}
