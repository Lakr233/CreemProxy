package src

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type IPRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
}

type rateLimitEntry struct {
	count     int
	lastReset time.Time
}

var rateLimiter = &IPRateLimiter{
	entries: make(map[string]*rateLimitEntry),
}

func (rl *IPRateLimiter) CheckRateLimit(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()

	entry, exists := rl.entries[ip]
	if !exists {
		entry = &rateLimitEntry{
			count:     1,
			lastReset: now,
		}
		rl.entries[ip] = entry
		return false
	}

	if now.Sub(entry.lastReset) > time.Minute {
		entry.count = 1
		entry.lastReset = now
		return false
	}

	entry.count++
	return entry.count > 10
}

func (rl *IPRateLimiter) CleanupExpired() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, entry := range rl.entries {
		if now.Sub(entry.lastReset) > 5*time.Minute {
			delete(rl.entries, ip)
		}
	}
}

func ProxyDirector(req *http.Request) {
	targetURL, err := url.Parse(CREEM_API_HOST)
	if err != nil {
		log.Printf("[-] unable to parse target URL: %v", err)
		return
	}
	req.Host = targetURL.Host
	req.URL.Scheme = targetURL.Scheme
	req.URL.Host = targetURL.Host
	req.Header.Set("x-api-key", CREEM_API_KEY)
}

func AddSignatureToResponse(res *http.Response) error {
	if res.Body == nil || res.Body == http.NoBody {
		return nil
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("[-] failed to read response body for signing: %v", err)
		res.Body = io.NopCloser(bytes.NewReader([]byte{}))
		return fmt.Errorf("failed to read response body: %w", err)
	}
	if closeErr := res.Body.Close(); closeErr != nil {
		log.Printf("[-] failed to close original response body: %v", closeErr)
		// continue, as we have the bytes.
	}

	if len(bodyBytes) > 0 {
		signature, signErr := Sign(bodyBytes)
		if signErr != nil {
			log.Printf("[-] failed to sign response body: %v", signErr)
		} else {
			res.Header.Set("x-api-signature", signature)
		}
	}

	res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return nil
}

func ProxyHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if fwdIP := r.Header.Get("X-Forwarded-For"); fwdIP != "" {
		clientIP = fwdIP
	}

	if rateLimiter.CheckRateLimit(clientIP) {
		http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
		log.Printf("[-] rate limit exceeded for IP: %s", clientIP)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		log.Printf("[-] invalid request method: %s", r.Method)
		return
	}
	if r.URL.Path == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		log.Printf("[-] invalid request path: %s", r.URL.Path)
		return
	}
	allowed := false
	for _, path := range ALLOWED_PATHS {
		if r.URL.Path == path {
			allowed = true
			break
		}
	}
	if !allowed {
		http.Error(w, "Not Found", http.StatusNotFound)
		log.Printf("[-] invalid request path: %s", r.URL.Path)
		return
	}

	if r.Header.Get("x-api-key") != "" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("[-] invalid API key: %s", r.Header.Get("x-api-key"))
		return
	}

	// later we rely on plain text signing
	r.Header.Set("Accept", "application/json; charset=utf-8")
	r.Header.Set("Accept-Encoding", "identity")

	log.Printf("[*] qualified request received: %s %s", r.Method, r.URL.Path)

	if r.Header.Get("Content-Type") == "application/json" {
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("[-] error reading request body: %v", err)
			http.Error(w, "Error reading request body", http.StatusInternalServerError)
			return
		} else {
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if len(bodyBytes) > 0 {
				log.Printf("[*] request body: %s", string(bodyBytes))
			} else {
				log.Printf("[*] request body is empty (Content-Type: application/json)")
			}
		}
	}

	proxy := &httputil.ReverseProxy{
		Director:       ProxyDirector,
		ModifyResponse: AddSignatureToResponse,
	}
	proxy.ServeHTTP(w, r)
}
