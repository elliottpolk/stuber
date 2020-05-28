package respond

import (
	"fmt"
	"net/http"
)

func Error(w http.ResponseWriter, code int, msg string) {
	http.Error(w, msg, code)
}

func Errorf(w http.ResponseWriter, code int, format string, args ...interface{}) {
	http.Error(w, fmt.Sprintf(format, args...), code)
}

func MethodNotAllowed(w http.ResponseWriter) {
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func Unauthorized(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusUnauthorized)
}
