package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	frankfurterURL = "https://api.frankfurter.dev/v1/latest?base=USD"
	cacheTTL       = 5 * time.Minute
	serverAddr     = ":8080"
)

var (
	targetCurrencies = []string{"CNY", "SEK", "EUR"}
	httpClient       = &http.Client{Timeout: 8 * time.Second}
)

type frankfurterResponse struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

type rateCache struct {
	mu        sync.RWMutex
	payload   *frankfurterResponse
	fetchedAt time.Time
}

var cache = &rateCache{}

type apiResponse struct {
	Base      string             `json:"base"`
	UpdatedAt time.Time          `json:"updatedAt"`
	Rates     map[string]float64 `json:"rates"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/all-rate", handleAllRate)

	port := os.Getenv("PORT")
	if port == "" {
		port = serverAddr
	} else if port[0] != ':' {
		port = ":" + port
	}

	log.Printf("Server is listening on %s\n", port)
	if err := http.ListenAndServe(port, withCORS(mux)); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func handleAllRate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rates, fetchedAt, err := getRates(r.Context())
	if err != nil {
		log.Printf("failed to get rates: %v", err)
		http.Error(w, "failed to fetch rates", http.StatusBadGateway)
		return
	}

	response := apiResponse{
		Base:      "USD",
		UpdatedAt: fetchedAt.UTC(),
		Rates:     rates,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func getRates(ctx context.Context) (map[string]float64, time.Time, error) {
	cache.mu.RLock()
	if cache.payload != nil && time.Since(cache.fetchedAt) < cacheTTL {
		defer cache.mu.RUnlock()
		return filterRates(cache.payload), cache.fetchedAt, nil
	}
	cache.mu.RUnlock()

	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.payload != nil && time.Since(cache.fetchedAt) < cacheTTL {
		return filterRates(cache.payload), cache.fetchedAt, nil
	}

	payload, fetchedAt, err := fetchFrankfurter(ctx)
	if err != nil {
		return nil, time.Time{}, err
	}

	cache.payload = payload
	cache.fetchedAt = fetchedAt

	return filterRates(payload), cache.fetchedAt, nil
}

func fetchFrankfurter(ctx context.Context) (*frankfurterResponse, time.Time, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, frankfurterURL, nil)
	if err != nil {
		return nil, time.Time{}, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, time.Time{}, errors.New("frankfurter responded with status " + resp.Status)
	}

	var payload frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, time.Time{}, err
	}
	if payload.Base != "USD" {
		return nil, time.Time{}, errors.New("unexpected base currency in response")
	}
	return &payload, time.Now(), nil
}

func filterRates(payload *frankfurterResponse) map[string]float64 {
	result := make(map[string]float64, len(targetCurrencies))
	for _, code := range targetCurrencies {
		if value, ok := payload.Rates[code]; ok {
			result[code] = value
		}
	}
	return result
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
