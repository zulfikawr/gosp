// Package protocol provides gRPC protocol definitions and TLS utilities for GOSP.
// It includes mTLS configuration for secure master-worker communication.
package protocol

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc/credentials"
)

// LoadServerCredentials loads the mTLS configuration for the Master server.
func LoadServerCredentials(caCert []byte, serverCert []byte, serverKey []byte) (credentials.TransportCredentials, error) {
	// 1. Load the OSP Cluster CA to verify Worker certificates
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append OSP Cluster CA to pool")
	}

	// 2. Load the Master's certificate and private key
	cert, err := tls.X509KeyPair(serverCert, serverKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load master keypair: %w", err)
	}

	// 3. Configure mTLS (Require client auth)
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(config), nil
}

// LoadClientCredentials loads the mTLS configuration for the Worker node.
func LoadClientCredentials(caCert []byte, clientCert []byte, clientKey []byte, serverName string) (credentials.TransportCredentials, error) {
	// 1. Load the OSP Cluster CA to verify the Master's certificate
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append OSP Cluster CA to pool")
	}

	// 2. Load the Worker's certificate and private key
	cert, err := tls.X509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, fmt.Errorf("failed to load worker keypair: %w", err)
	}

	// 3. Configure mTLS
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   serverName, // Must match the Master's certificate CommonName
		MinVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(config), nil
}
