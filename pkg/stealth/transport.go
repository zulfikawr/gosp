package stealth

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/http2"
)

// NewStealthClient returns an http.Client that spoofs its TLS fingerprint (JA3).
// It mimics a modern Chrome browser to bypass simple TLS-based bot detection.
func NewStealthClient(timeout time.Duration) *http.Client {
	// 1. Create a custom DialTLSContext that uses uTLS for the handshake.
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

			// 2. Use uTLS to spoof the HelloChrome fingerprint.
			// This matches modern Chromium (v120+) TLS behaviors.
			uConn := utls.UClient(conn, &utls.Config{ServerName: host}, utls.HelloChrome_120)
			if err := uConn.Handshake(); err != nil {
				conn.Close()
				return nil, err
			}
			return uConn, nil
		},
		TLSClientConfig: &tls.Config{
			// The actual config is handled by uTLS inside DialTLSContext.
			// Setting InsecureSkipVerify is optional here but risky for a scraper.
			InsecureSkipVerify: false,
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 3. Enable HTTP/2 for the stealth transport.
	// Many modern anti-bot systems flag clients that only use HTTP/1.1.
	if err := http2.ConfigureTransport(transport); err != nil {
		// Log error if needed, but continue with HTTP/1.1 fallback.
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
}

// GetRandomUserAgent returns a random modern User-Agent string to help avoid detection.
func GetRandomUserAgent() string {
	// We'll move the UA list here for a more consolidated stealth package.
	// This should match the TLS fingerprint we're spoofing above (Chrome 120+).
	return userAgents[0] 
}
