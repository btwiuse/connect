package main

import (
	"github.com/btwiuse/connect"
)

func main() {
	if HTTPS_PORT != nil {
		println("listening on", *HTTPS_PORT)
		ListenAndServeTLS(connect.Handler)
	}
	println("listening on", PORT)
	ListenAndServe(connect.Handler)
}
