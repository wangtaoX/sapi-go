package middleware

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
)

type Middleware struct {
	Logger *logrus.Logger
	Name   string
}

func NewLog(name string) *Middleware {
	log := logrus.New()
	log.Level = logrus.InfoLevel
	log.Formatter = &logrus.TextFormatter{}
	return &Middleware{Logger: log, Name: name}
}

func (l *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	entry := l.Logger.WithFields(logrus.Fields{
		"Req":           r.RequestURI,
		"Method":        r.Method,
		"RemoteAddress": r.RemoteAddr,
	})
	if reqID := r.Header.Get("X-Request-Id"); reqID != "" {
		entry = entry.WithField("request_id", reqID)
	}
	entry.Info("Starting Handling request")
	next(rw, r)
	latency := time.Since(start)
	res := rw.(negroni.ResponseWriter)
	entry.WithFields(logrus.Fields{
		"status":      res.Status(),
		"text_status": http.StatusText(res.Status()),
		"took":        latency,
	}).Info("Completed Handling request")
}
