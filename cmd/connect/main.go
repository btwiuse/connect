package main

import (
	"log"
	"net/http"

	"github.com/btwiuse/connect"
)

func logProto(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Proto, r.Host)
		next.ServeHTTP(w, r)
	})
}

func main() {
	handler := logProto(connect.Handler)
	if true {
		println("listening on QUIC", UDP_PORT)
		go ListenAndServeQUIC(handler)
	}
	if HTTPS_PORT != nil {
		println("listening on", *HTTPS_PORT)
		go ListenAndServeTLS(handler)
	}
	println("listening on", PORT)
	ListenAndServe(handler)
}
