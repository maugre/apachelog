package accesslog

// This file is forked from https://gist.github.com/Tantas/1fc00c5eb7c291e2a34b
// Original license applies

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// https://httpd.apache.org/docs/2.2/logs.html#combined + execution time.
const (
	apacheFormatPattern = "%s - - [%s] \"%s %s %s\" %d %d \"%s\" \"%s\" %.6f\n"
	defaultTimeFormat   = "02/Jan/2006 15:04:05"
)

var formatPattern string = apacheFormatPattern
var timeFormat string = defaultTimeFormat

type ApacheLogRecord struct {
	http.ResponseWriter

	ip                    string
	time                  time.Time
	method, uri, protocol string
	status                int
	responseBytes         int64
	referer               string
	userAgent             string
	elapsedTime           time.Duration
}

func (r *ApacheLogRecord) Log(out io.Writer) {
	timeFormatted := r.time.Format(timeFormat)
	fmt.Fprintf(out, apacheFormatPattern, r.ip, timeFormatted, r.method,
		r.uri, r.protocol, r.status, r.responseBytes, r.referer, r.userAgent,
		r.elapsedTime.Seconds())
}

func (r *ApacheLogRecord) Write(p []byte) (int, error) {
	written, err := r.ResponseWriter.Write(p)
	r.responseBytes += int64(written)
	return written, err
}

func (r *ApacheLogRecord) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

type ApacheLoggingHandler struct {
	handler http.Handler
	out     io.Writer
}

func NewApacheLoggingHandler(handler http.Handler, out io.Writer) http.Handler {
	return &ApacheLoggingHandler{
		handler: handler,
		out:     out,
	}
}

func (h *ApacheLoggingHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	clientIP := r.RemoteAddr
	if colon := strings.LastIndex(clientIP, ":"); colon != -1 {
		clientIP = clientIP[:colon]
	}

	referer := r.Referer()
	if referer == "" {
		referer = "-"
	}

	userAgent := r.UserAgent()
	if userAgent == "" {
		userAgent = "-"
	}

	// This section is modified to avoid setting startTime and finishTime variables.
	// Instead we record the time directly to the ApacheLogRecord struct
	record := &ApacheLogRecord{
		ResponseWriter: rw,
		ip:             clientIP,
		time:           time.Now(),
		method:         r.Method,
		uri:            r.RequestURI,
		protocol:       r.Proto,
		status:         http.StatusOK,
		referer:        referer,
		userAgent:      userAgent,
		elapsedTime:    time.Duration(0),
	}

	h.handler.ServeHTTP(record, r)
	record.elapsedTime = time.Since(record.time)

	record.Log(h.out)
}
