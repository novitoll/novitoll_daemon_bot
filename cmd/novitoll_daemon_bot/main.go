package main

import (
	"net/http"
	
	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/router"
)

func main() {
	handler := router.RouteHandler{}

	http.HandleFunc("/faq", handler.FaqHandler)
	http.ListenAndServe(":8080", nil)
}
