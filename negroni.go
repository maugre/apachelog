package accesslog

// Part of this file is derived from the example shown at
// https://gist.github.com/Tantas/1fc00c5eb7c291e2a34b
// Modifications allow this to be used as part of a Negroni middleware chain

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type NegroniAccessLog struct {
	writer *os.File
}

func NewNegroniAccessLog(path string) *NegroniAccessLog {
	// Open appending or create the access log.
	accessLogFile, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("Error opening file %s for writing: %s", path, err.Error())
	}

	// Create a buffered writer and ensure it is flushed when an interrupt occurs.
	bufferedAccessLogWriter := bufio.NewWriter(accessLogFile)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		defer accessLogFile.Close()
		<-c
		log.Println("Flushing logs and exiting...")
		bufferedAccessLogWriter.Flush()
		os.Exit(0)
	}()

	return &NegroniAccessLog{accessLogFile}
}

// Define ServeHTTP method required for Negroni middleware
func (n *NegroniAccessLog) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// NOTE: The logging handler calls the next handler in the chain so we
	// don't explicitly call next() here
	// That may cause problems in some cases but so far has worked in mine.
	// Any suggestions welcome.
	handler := NewApacheLoggingHandler(next, n.writer)
	handler.ServeHTTP(rw, r)
}
