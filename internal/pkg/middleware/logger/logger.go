package logger

import (
	"log"
	"net/http"
	"time"
)

// Logging middleware, for now uses global Log instance

type Logger struct {
	next     http.Handler
	reqCount int
}

func NewLogger(next http.Handler) *Logger {
	return &Logger{next, 0}
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if reason := recover(); reason != nil {
			log.Printf("[%s] [client %s] %s",
				time.Now().Format(time.RFC3339),
				r.RemoteAddr, reason)
		}
	}()

	start := time.Now()

	l.next.ServeHTTP(w, r)

	log.Printf("[%s] [client %s] \"%s\"",
		start.Format(time.RFC3339),
		r.RemoteAddr,
		r.RequestURI)

}
