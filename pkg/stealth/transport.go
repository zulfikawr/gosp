// Package stealth provides anti-detection HTTP transport for web scraping.
// It uses TLS fingerprint spoofing to evade bot detection mechanisms.
package stealth

import (
	"context"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
)

// NewStealthClient returns an http.Client that spoofs its TLS fingerprint (JA3).
// It defaults to HTTP/1.1 to maximize stability against strict H2 server checks.
func NewStealthClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		DialContext: dialer.DialContext,
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			conn, err := dialer.DialContext(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			host, _, _ := net.SplitHostPort(addr)

			// Use uTLS to spoof the HelloChrome fingerprint.
			// We force NextProtos to only include http/1.1 to avoid H2 handshake failures
			// which are common when scraping from datacenter IPs.
			uConn := utls.UClient(conn, &utls.Config{
				ServerName: host,
				NextProtos: []string{"http/1.1"},
			}, utls.HelloChrome_120)

			if err := uConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			return uConn, nil
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Explicitly disable H2 to prevent transport mismatches
		ForceAttemptHTTP2: false,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
}

// GetRandomUserAgent returns a random user agent string for request headers.
func GetRandomUserAgent() string {
	return userAgents[0]
}
