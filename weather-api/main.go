package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	apiKey = os.Getenv("VISUAL_CROSSING_API_KEY")
	opt    *redis.Options
)

func init() {
	if apiKey == "" {
		log.Fatal("VISUAL_CROSSING_API_KEY is not set")
	}

	host, port := os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")
	if host == "" || port == "" {
		log.Fatal("REDIS_HOST and REDIS_PORT must be set")
	}

	opt = &redis.Options{
		Addr:     host + ":" + port,
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}

const (
	apiBase          = "https://weather.visualcrossing.com/VisualCrossingWebServices/rest/services"
	timelineEndpoint = "/timeline/%s"
)

var client = http.Client{
	Timeout: 10 * time.Second,
}

func fetch(city string, out any) (int, error) {
	url := fmt.Sprintf(apiBase+timelineEndpoint, city)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	q := req.URL.Query()
	q.Set("key", apiKey)
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		// continue
	case http.StatusBadRequest:
		return http.StatusBadRequest, fmt.Errorf("invalid request")
	case http.StatusUnauthorized:
		return http.StatusUnauthorized, fmt.Errorf("unauthorized request")
	case http.StatusNotFound:
		return http.StatusNotFound, fmt.Errorf("city not found")
	case http.StatusTooManyRequests:
		return http.StatusTooManyRequests, fmt.Errorf("too many requests")
	default:
		return http.StatusBadGateway, fmt.Errorf("weather provider returned %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to Weather API")
}

func cityHandler(w http.ResponseWriter, r *http.Request) {
	allowed, err := rateLimit(r)
	if err != nil {
		http.Error(w, "rate limiter unavailable", http.StatusInternalServerError)
		return
	}
	if !allowed {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}

	city := r.PathValue("city")

	var weather any

	if !checkRedisCache(r.Context(), "weather:"+city, &weather) {
		status, err := fetch(city, &weather)
		if err != nil {
			http.Error(w, err.Error(), status)
			return
		}

		b, err := json.Marshal(weather)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = setRedisCache(r.Context(), b, "weather:"+city); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(weather); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func checkRedisCache(ctx context.Context, key string, out any) bool {
	cached, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil || err != nil {
		return false
	}

	if err = json.Unmarshal([]byte(cached), out); err != nil {
		return false
	}

	return true
}

func setRedisCache(ctx context.Context, weather any, key string) error {
	err := rdb.Set(ctx, key, weather, 12*time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func rateLimit(r *http.Request) (bool, error) {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	key := "rate:" + host
	count, err := rdb.Incr(r.Context(), key).Result()
	if err != nil {
		return false, err
	}
	if count == 1 {
		_ = rdb.Expire(r.Context(), key, time.Minute)
	} else if count > 100 {
		return false, nil
	}
	return true, nil
}

var rdb *redis.Client

func main() {
	rdb = redis.NewClient(opt)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/{city}", cityHandler)

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
