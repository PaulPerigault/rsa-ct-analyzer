// Package ct is a minimal RFC 6962 Certificate Transparency client: it reads
// the Signed Tree Head and log entries, and extracts the certificates.
package ct

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// maxResponseBytes bounds one HTTP response body (32 MiB).
const maxResponseBytes = 32 << 20

// ClientConfig configures a Client.
type ClientConfig struct {
	BaseURL   string        // e.g. https://oak.ct.letsencrypt.org/2026h2
	Timeout   time.Duration // per HTTP request
	RateLimit float64       // requests per second, 0 = unlimited
	Retries   int           // retries on 429 / 5xx / network errors
	Backoff   time.Duration // first retry delay, doubled each time
	UserAgent string
	HTTP      *http.Client // optional, for tests
}

// Client talks to one CT log.
type Client struct {
	cfg  ClientConfig
	base string
	http *http.Client

	mu   sync.Mutex
	last time.Time
}

// NewClient validates cfg and returns a ready client.
func NewClient(cfg ClientConfig) (*Client, error) {
	u, err := url.Parse(cfg.BaseURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, fmt.Errorf("invalid CT log URL %q: want http(s)://host/path", cfg.BaseURL)
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Backoff <= 0 {
		cfg.Backoff = time.Second
	}
	hc := cfg.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: cfg.Timeout}
	}
	return &Client{cfg: cfg, base: strings.TrimRight(cfg.BaseURL, "/"), http: hc}, nil
}

// STH is the Signed Tree Head (signature not verified: the tool audits keys,
// not the log).
type STH struct {
	TreeSize  int64 `json:"tree_size"`
	Timestamp int64 `json:"timestamp"`
}

// RawEntry is one get-entries item, still base64-encoded.
type RawEntry struct {
	LeafInput string `json:"leaf_input"`
	ExtraData string `json:"extra_data"`
}

// GetSTH returns the current tree head.
func (c *Client) GetSTH(ctx context.Context) (*STH, error) {
	var sth STH
	if err := c.getJSON(ctx, "/ct/v1/get-sth", &sth); err != nil {
		return nil, fmt.Errorf("get-sth: %w", err)
	}
	return &sth, nil
}

// GetEntries returns entries start..end inclusive. A log may return fewer
// than requested: callers must advance by len(result).
func (c *Client) GetEntries(ctx context.Context, start, end int64) ([]RawEntry, error) {
	var resp struct {
		Entries []RawEntry `json:"entries"`
	}
	path := fmt.Sprintf("/ct/v1/get-entries?start=%d&end=%d", start, end)
	if err := c.getJSON(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("get-entries %d-%d: %w", start, end, err)
	}
	return resp.Entries, nil
}

// errRetryable marks failures worth another attempt.
type errRetryable struct {
	err   error
	after time.Duration
}

func (e *errRetryable) Error() string { return e.err.Error() }
func (e *errRetryable) Unwrap() error { return e.err }

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	delay := c.cfg.Backoff
	var err error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if err = c.throttle(ctx); err != nil {
			return err
		}
		err = c.getOnce(ctx, path, out)
		var retry *errRetryable
		if err == nil || !errors.As(err, &retry) || attempt == c.cfg.Retries {
			return err
		}
		wait := max(delay, retry.after)
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
		delay *= 2
	}
	return err
}

// throttle spaces requests to honour RateLimit.
func (c *Client) throttle(ctx context.Context) error {
	if c.cfg.RateLimit <= 0 {
		return ctx.Err()
	}
	interval := time.Duration(float64(time.Second) / c.cfg.RateLimit)
	c.mu.Lock()
	next := c.last.Add(interval)
	if now := time.Now(); next.Before(now) {
		next = now
	}
	c.last = next
	c.mu.Unlock()
	select {
	case <-time.After(time.Until(next)):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) getOnce(ctx context.Context, path string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &errRetryable{err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		secs, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return &errRetryable{
			err:   fmt.Errorf("log answered HTTP %d", resp.StatusCode),
			after: time.Duration(secs) * time.Second,
		}
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("log answered HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return &errRetryable{err: err}
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}
