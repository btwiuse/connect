package main

import (
	"crypto/tls"
	"net/http"

	"github.com/webteleport/utils"
)

var (
	HOST       = utils.EnvHost("localhost")
	CERT       = utils.EnvCert("localhost.pem")
	KEY        = utils.EnvKey("localhost-key.pem")
	PORT       = utils.EnvPort(":3000")
	HTTPS_PORT = utils.LookupEnvPort("HTTPS_PORT")
)

var NextProtos = []string{"http/1.1", "h2"}

func LocalTLSConfig(certFile, keyFile string) *tls.Config {
	GetCertificate := func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		// Always get latest localhost.crt and localhost.key
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		return &cert, nil
	}
	return &tls.Config{
		GetCertificate: GetCertificate,
		NextProtos:     NextProtos,
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
