package respond

import (
	"fmt"
	"net/http"
)

func OK(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "")
}

func Created(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "")
}

func NoContent(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; charset=us-ascii")
	w.WriteHeader(http.StatusNoContent)
	fmt.Fprintln(w, "")
}

func With(w http.ResponseWriter, code int, contentType, format string, args ...interface{}) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(code)
	fmt.Fprintln(w, fmt.Sprintf(format, args...))
}
