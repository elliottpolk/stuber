package respond

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

func Json(w http.ResponseWriter, d interface{}) {
	if d == nil {
		Created(w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	out, err := json.Marshal(d)
	if err != nil {
		Error(w, http.StatusInternalServerError, errors.Wrap(err, "unable to convert results to JSON").Error())
		return
	}

	fmt.Fprint(w, string(out))
}
