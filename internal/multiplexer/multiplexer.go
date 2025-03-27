package multiplexer

import (
	"context"
	"encoding/json"
	"fmt"
	"http_server/internal/multiplexer/cache"
	"http_server/internal/multiplexer/retryer"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type Multiplexer struct {
	retryer retryer.Retryer
	cache   cache.Cache

	fetchTimeOut int
	// sem is a channel for limiting concurrent requests
	sem         chan struct{}
	maxURLs     int
	workerLimit int
}

type result struct {
	url  string
	body string
}

type Options struct {
	MaxUrls       int
	RequestsLimit int
	WorkerLimit   int
	FetchTimeout  int
	Cache         cache.Cache

	Retry      bool
	NumRetries int
	Delay      int
	FillRatio  int
}

const (
	maxURLsDefault      = 20
	requestLimitDefault = 100
	workerLimitDefault  = 4
	fetchTimeoutDefault = 1
)

func NewMultiplexer(opts Options) *Multiplexer {
	var m Multiplexer

	m.maxURLs = maxURLsDefault

	if opts.MaxUrls != 0 {
		m.maxURLs = opts.MaxUrls
	}

	var requestLimit = requestLimitDefault

	if opts.RequestsLimit != 0 {
		requestLimit = opts.RequestsLimit
	}

	m.sem = make(chan struct{}, requestLimit)

	m.workerLimit = workerLimitDefault

	if opts.WorkerLimit != 0 {
		m.workerLimit = opts.WorkerLimit
	}

	m.fetchTimeOut = fetchTimeoutDefault

	if opts.FetchTimeout != 0 {
		m.fetchTimeOut = opts.FetchTimeout
	}

	if opts.Cache != nil {
		m.cache = opts.Cache
	} else {
		m.cache = &cache.DummyCache{}
	}

	if opts.Retry {
		m.retryer = retryer.NewRetryer(opts.NumRetries, opts.Delay, opts.FillRatio, m.getRatio)
	} else {
		m.retryer = &retryer.DummyRetryer{}
	}

	return &m
}

type fetchRequest struct {
	URLs []string `json:"urls"`
}

func (m *Multiplexer) FetchHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case m.sem <- struct{}{}:
		defer func() { <-m.sem }()
	default:
		http.Error(w, "too many requests. try again later", http.StatusTooManyRequests)
		return
	}

	var req fetchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(req.URLs) == 0 || len(req.URLs) > m.maxURLs {
		http.Error(w, "too many urls", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	results, err := m.fetchURLs(ctx, req.URLs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(results)
}

// fetchURLs concurrently fetches urls by spawning worker goroutines
func (m *Multiplexer) fetchURLs(ctx context.Context, urls []string) (map[string]string, error) {
	var (
		wg        sync.WaitGroup
		resultCh  = make(chan result)
		urlsCh    = make(chan string, len(urls))
		responses = make(map[string]string, len(urls))
	)

	for _, url := range urls {
		urlsCh <- url
	}

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	wg.Add(m.workerLimit)

	for i := 0; i < m.workerLimit; i++ {
		go m.worker(ctx, &wg, resultCh, urlsCh, cancel)
	}

	go func() {
		wg.Wait()
		close(urlsCh)
	}()

loop:
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case res := <-resultCh:
			responses[res.url] = res.body[:20]

			if len(responses) >= len(urls) {
				close(resultCh)
				break loop
			}
		}
	}

	return responses, nil
}

func (m *Multiplexer) worker(ctx context.Context, wg *sync.WaitGroup, results chan<- result, urls <-chan string, cancel func(error)) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case url, ok := <-urls:
			if !ok {
				return
			}

			resp, err := m.fetchURL(ctx, url)
			if err != nil {
				cancel(err)

				return
			}

			results <- result{url: url, body: resp}

		}
	}
}

func (m *Multiplexer) fetchURL(ctx context.Context, url string) (string, error) {
	cachedRes, err := m.cache.Get(url)
	if err == nil {
		return cachedRes, nil
	}

	log.Printf("Fetching %s from cache error: %s", url, err)
	log.Printf("Fetching %s from server", url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: time.Duration(m.fetchTimeOut) * time.Second}

	var (
		attempts = 0
		resp     *http.Response
	)

	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		attempts++

		resp, err = client.Do(req)
		if err != nil {
			log.Printf("Error fetching %s: %v, %d attempts made", url, err, attempts)

			if !m.retryer.Retry(ctx, attempts) {
				break
			}

			continue
		}

		defer resp.Body.Close()

		break
	}

	if err != nil {
		return "", err
	}

	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got %d status code for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	res := string(body)

	err = m.cache.Set(url, res)
	if err != nil {
		log.Printf("Error caching %s: %v", url, err)
	}

	return res, nil
}

func (m *Multiplexer) getRatio() float64 {
	return float64(len(m.sem)) / float64(cap(m.sem))
}
