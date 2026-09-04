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
}

// New creates a new Proxy that forwards requests to origin and caches
// responses in the provided Redis client.
func New(origin string, c *cache.Client) (*Proxy, error) {
	u, err := url.Parse(origin)
	if err != nil {
		return nil, fmt.Errorf("parsing origin url: %w", err)
	}

	return &Proxy{
		origin: u,
		cache:  c,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ServeHTTP implements http.Handler.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := cacheKey(r)

	if cached, hit := p.fromCache(r.Context(), key); hit {
		writeResponse(w, cached, true)
		log.Printf("CACHE HIT  %s %s", r.Method, r.URL.Path)
		return
	}

	upstream, err := p.fetchUpstream(r)
	if err != nil {
		log.Printf("upstream error: %v", err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	defer upstream.Body.Close()

	body, err := io.ReadAll(upstream.Body)
	if err != nil {
		log.Printf("reading upstream body: %v", err)
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

	writeResponse(w, resp, false)
	log.Printf("CACHE MISS %s %s -> %d", r.Method, r.URL.Path, upstream.StatusCode)
}

// ClearAll deletes every key in the Redis DB.
// This is intentionally destructive — call only from the --clear-cache path.
func (p *Proxy) ClearAll(ctx context.Context) error {
	return p.cache.Client.FlushDB(ctx).Err()
}

// fetchUpstream forwards the incoming request to the origin server.
func (p *Proxy) fetchUpstream(r *http.Request) (*http.Response, error) {
	target := *p.origin
	target.Path = strings.TrimRight(target.Path, "/") + r.URL.Path
	target.RawQuery = r.URL.RawQuery

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target.String(), r.Body)
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

func writeResponse(w http.ResponseWriter, resp *cachedResponse, fromCache bool) {
	for k, vv := range resp.Headers {
		if isHopByHop(k) {
			continue
		}
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}

	if fromCache {
		w.Header().Set("X-Cache", "HIT")
	} else {
		w.Header().Set("X-Cache", "MISS")
	}

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
