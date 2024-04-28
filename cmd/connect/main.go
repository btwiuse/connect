package main

import (
	"fmt"
	"net/http"

	"github.com/btwiuse/connect"
)

func main() {
	if HTTPS_PORT != nil {
		ListenAndServeTLS(connect.Handler)
	}
	ListenAndServe(connect.Handler)
}
