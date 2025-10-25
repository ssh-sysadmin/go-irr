package main

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// loggingResponseWriter wraps http.ResponseWriter to capture the status code
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *loggingResponseWriter) Write(b []byte) (int, error) {
	if lrw.status == 0 {
		lrw.status = http.StatusOK
	}
	return lrw.ResponseWriter.Write(b)
}

// getSourceIP extracts the client's IP address, preferring X-Forwarded-For/X-Real-Ip
func getSourceIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := r.Header.Get("X-Real-Ip"); xr != "" {
		return strings.TrimSpace(xr)
	}
	host := r.RemoteAddr
	if host == "" {
		return ""
	}
	ip, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return ip
}

// loggingMiddleware logs status code, source IP and request path for each request
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// capture start time, then serve the request and measure elapsed time
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w}
		next.ServeHTTP(lrw, r)
		elapsed := time.Since(start)

		status := lrw.status
		if status == 0 {
			status = http.StatusOK
		}
		src := getSourceIP(r)
		// log timestamp, status, source IP, path and request duration
		fmt.Printf("%s status=%d src=%s path=%s duration=%s\n", time.Now().Format(time.RFC3339), status, src, r.URL.Path, elapsed)
	})
}
