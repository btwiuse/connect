package main

import (
	"github.com/btwiuse/connect"
)

func main() {
	if HTTPS_PORT != nil {
		ListenAndServeTLS(connect.Handler)
	}
	ListenAndServe(connect.Handler)
}
