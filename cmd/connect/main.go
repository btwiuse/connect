package main

import (
	"github.com/btwiuse/connect"
)

func main() {
	if true {
		println("listening on QUIC", UDP_PORT)
		go ListenAndServeQUIC(connect.Handler)
	}
	if HTTPS_PORT != nil {
		println("listening on", *HTTPS_PORT)
		go ListenAndServeTLS(connect.Handler)
	}
	println("listening on", PORT)
	ListenAndServe(connect.Handler)
}
