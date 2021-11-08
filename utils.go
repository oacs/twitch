package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

// HumanReadableError represents error information
// that can be fed back to a human user.
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	HumanError() string
	HTTPCode() int
}

// HumanReadableWrapper implements HumanReadableError
type HumanReadableWrapper struct {
	ToHuman string
	Code    int
	error
}

type Handler func(http.ResponseWriter, *http.Request) error

func handleFunc(path string, handler Handler) {
	http.Handle(path, errorHandling(middleware(handler)))
}

// AnnotateError wraps an error with a message that is intended for a human end-user to read,
// plus an associated HTTP error code.
func AnnotateError(err error, annotation string, code int) error {
	if err == nil {
		return nil
	}
	return HumanReadableWrapper{ToHuman: annotation, error: err}
}

func middleware(h Handler) Handler {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// parse POST body, limit request size
		if err = r.ParseForm(); err != nil {
			return AnnotateError(err, "Something went wrong! Please try again.", http.StatusBadRequest)
		}

		return h(w, r)
	}
}

// errorHandling is a middleware that centralises error handling.
// this prevents a lot of duplication and prevents issues where a missing
// return causes an error to be printed, but functionality to otherwise continue
// see https://blog.golang.org/error-handling-and-go
func errorHandling(handler func(w http.ResponseWriter, r *http.Request) error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var errorString string = "Something went wrong! Please try again."
			var errorCode int = 500

			if v, ok := err.(HumanReadableError); ok {
				errorString, errorCode = v.HumanError(), v.HTTPCode()
			}

			log.Fatal(err)
			w.Write([]byte(errorString))
			w.WriteHeader(errorCode)
			return
		}
	})
}

func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }
func (h HumanReadableWrapper) HTTPCode() int      { return h.Code }
