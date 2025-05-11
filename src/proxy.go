package src

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

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
	if r.Header.Get("Accept") == "" {
		r.Header.Set("Accept", "application/json; charset=utf-8")
	} else {
		log.Printf("[-] invalid Accept header: %s", r.Header.Get("Accept"))
		http.Error(w, "Bad Request: Header `Accept` must not be set", http.StatusBadRequest)
		return
	}
	if r.Header.Get("Accept-Encoding") == "" {
		r.Header.Set("Accept-Encoding", "identity")
	} else {
		log.Printf("[-] invalid Accept-Encoding header: %s", r.Header.Get("Accept-Encoding"))
		http.Error(w, "Bad Request: Header `Accept-Encoding` must not be set", http.StatusBadRequest)
		return
	}

	log.Printf("[*] qualified request received: %s %s", r.Method, r.URL.Path)

	proxy := &httputil.ReverseProxy{
		Director:       ProxyDirector,
		ModifyResponse: AddSignatureToResponse,
	}
	proxy.ServeHTTP(w, r)
}
