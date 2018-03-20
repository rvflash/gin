// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

const (
	noWritten     = -1
	defaultStatus = 200
)

// ResponseWriter is used by an Gin HTTP handler to
// construct an HTTP response.
type ResponseWriter interface {
	http.CloseNotifier
	http.Flusher
	http.Hijacker
	http.ResponseWriter

	// Pusher returns the http.Pusher for server push if it supported.
	Pusher() (http.Pusher, bool)

	// Size returns the number of bytes already written into the response http body.
	// See Written()
	Size() int

	// Status returns the HTTP response status code of the current request.
	Status() int

	// WriteHeaderNow forces to write the http header (status code + headers).
	WriteHeaderNow()

	// WriteString writes the string into the response body.
	WriteString(string) (int, error)

	// Written returns true if the response body was already written.
	Written() bool
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponseWriter = &responseWriter{}

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

// WriteHeader implements the ResponseWriter interface.
func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
		}
		w.status = code
	}
}

// WriteHeaderNow implements the ResponseWriter interface.
func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

// Write implements the ResponseWriter interface.
func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

// WriteString implements the ResponseWriter interface.
func (w *responseWriter) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = io.WriteString(w.ResponseWriter, s)
	w.size += n
	return
}

// Status implements the ResponseWriter interface.
func (w *responseWriter) Status() int {
	return w.status
}

// Size implements the ResponseWriter interface.
func (w *responseWriter) Size() int {
	return w.size
}

// Written implements the ResponseWriter interface.
func (w *responseWriter) Written() bool {
	return w.size != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if w.size < 0 {
		w.size = 0
	}
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (w *responseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

// Flush implements the http.Flush interface.
func (w *responseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

// Pusher returns the http.Pusher interface.
func (w *responseWriter) Pusher() (pusher http.Pusher, ok bool) {
	pusher, ok = w.ResponseWriter.(http.Pusher)
	return
}
