package main

import (
	"crypto/tls"
	"net/http"
	"os"

	"github.com/quic-go/quic-go/http3"
	"github.com/webteleport/utils"
)

var (
	HOST       = utils.EnvHost("localhost")
	CERT       = utils.EnvCert("localhost.pem")
	KEY        = utils.EnvKey("localhost-key.pem")
	PORT       = utils.EnvPort(":3000")
	UDP_PORT   = utils.EnvUDPPort(PORT)
	HTTPS_PORT = utils.LookupEnvPort("HTTPS_PORT")
)

func LocalTLSConfig(certFile, keyFile string) *tls.Config {
	GetCertificate := func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		// Always get latest localhost.crt and localhost.key
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		return &cert, nil
	}
	nextProtos := []string{
		"http/1.1",
	}
	// Prefer h2 if H2 env is set
	if os.Getenv("H2") != "" {
		nextProtos = []string{
			"h2",
			"http/1.1",
		}
	}

	return &tls.Config{
		GetCertificate: GetCertificate,
		NextProtos:     nextProtos,
		MinVersion:     tls.VersionTLS12,
	}
}

func ListenAndServeTLS(router http.Handler) error {
	tlsConfig := LocalTLSConfig(CERT, KEY)
	ln, err := tls.Listen("tcp", *HTTPS_PORT, tlsConfig)
	if err != nil {
		println(err.Error())
		return err
	}
	err = http.Serve(ln, router)
	if err != nil {
		println(err.Error())
		return err
	}
	return nil
}

func ListenAndServeQUIC(router http.Handler) error {
	return http3.ListenAndServeQUIC(UDP_PORT, CERT, KEY, router)
}

func ListenAndServe(router http.Handler) error {
	if HTTPS_PORT != nil {
		go ListenAndServeTLS(router)
	}
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		println(err.Error())
		return err
	}
	return nil
}
