package main

import (
	"net/http"
	
	r "github.com/novitoll/novitoll_daemon_bot/internal/vahter/router"
)

func main() {
	handler := r.RouteHandler{}

	http.HandleFunc("/faq", handler.FaqHandler)
	http.ListenAndServe(":8080", nil)
}
