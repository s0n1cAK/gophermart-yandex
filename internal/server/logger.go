package server

import (
	"net/http"
)

type responseData struct {
	status int
	size   int
}
type loggerRW struct {
	http.ResponseWriter
	responseData *responseData
	wroteHeader  bool
	Err          error
}

func (w *loggerRW) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.responseData.status = code
	w.ResponseWriter.WriteHeader(code)
	w.wroteHeader = true
}

func (w *loggerRW) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	size, err := w.ResponseWriter.Write(b)
	w.responseData.size += size
	return size, err
}
