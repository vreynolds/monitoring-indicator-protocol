package mtls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
)

func NewServer(address, serverPEM, serverKey, rootCACert string, handler http.Handler) (func() error, func() error, error) {
	caCert, err := ioutil.ReadFile(rootCACert)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read root CA certificate: %s\n", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	server := &http.Server{
		Addr:    address,
		Handler: handler,
		TLSConfig: &tls.Config{
			ClientCAs:                caCertPool,
			ClientAuth:               tls.RequireAndVerifyClientCert,
			PreferServerCipherSuites: true,
		},
	}

	start := func() error { return server.ListenAndServeTLS(serverPEM, serverKey) }
	stop := func() error { return server.Close() }

	return start, stop, nil
}

func NewClient(clientCert, clientKey, rootCACert string) (*http.Client, error) {
	cert, err := tls.LoadX509KeyPair(clientCert, clientKey)
	if err != nil {
		return nil, err
	}

	// Load CA cert
	caCert, err := ioutil.ReadFile(rootCACert)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	return client, nil
}