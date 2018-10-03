package router

import (
	"encoding/json"
	"github.com/novitoll/novitoll_daemon_bot/internal/vahter/bot"
	"net/http"
)

type requestHandler func(w http.ResponseWriter, r *http.Request) error

func handlerWrapper(fn requestHandler) http.Handler {
	return handleRequest(fn)
}

func handleRequest(fn requestHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			if err == ErrAccess {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if _, dontlog := err.(ErrDontLog); !dontlog {
				log.Errorf(c, "%v", err)
			}
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}