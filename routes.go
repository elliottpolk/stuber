package stuber

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"git.platform.manulife.io/gwam/stuber/internal/respond"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func handler(d, f string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		in, err := ioutil.ReadFile(filepath.Join(d, f))
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		stubs := gjson.Get(string(in), fmt.Sprintf(`stubs.#(request.method=="%s")#`, r.Method)).Array()
		if len(stubs) < 1 {
			respond.MethodNotAllowed(w)
			return
		}

		reqb, err := ioutil.ReadAll(r.Body)
		if err != nil {
			respond.Error(w, http.StatusBadRequest, err.Error())
			return
		}

		qparams := r.URL.Query()

		for _, stub := range stubs {

			var (
				name = stub.Get("name").String()

				reqd  = stub.Get("request.payload")
				respd = stub.Get("response.payload")
			)

			// handle empty request data
			if !reqd.Exists() {
				// this expects an empty request yet the caller send over a payload
				if len(qparams.Encode()) > 0 || len(reqb) > 0 {
					respond.Error(w, http.StatusBadRequest, "invalid request")
					return
				}

				// respond with the configured error payload and exit
				if stub.Get("response.type").String() == "error" {
					respond.Error(w, int(respd.Get("code").Int()), respd.Get("message").String())
					return
				}

				log.Debugf("responding for payload set %s", name)
				respond.Json(w, respd.Value())
				return
			}

			// if we get here, we have some request data
			// of which a GET will have query params where most others will use the request body
			switch r.Method {
			case http.MethodPost, http.MethodPut:
				// force a cleanup on the request body to match the file data
				var i interface{}
				if err := json.Unmarshal(reqb, &i); err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				reqb, err = json.Marshal(i)
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				// marshal instead of using .String() to ensure the formats are consistent
				dat, err := json.Marshal(reqd.Value())
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				if string(dat) != string(reqb) {
					continue
				}

				log.Debugf("responding for payload set %s", name)

				// this should handle something like a create request via POST or PUT
				if stub.Get("response.type").String() == "data.empty" {
					respond.With(w, int(respd.Get("code").Int()), respd.Get("content-type").String(), "")
					return
				}

				respond.Json(w, respd.Value())
				return

			case http.MethodGet:
				// convert the request payload to query params
				params := &url.Values{}
				reqd.ForEach(func(k, v gjson.Result) bool {
					params.Add(k.String(), v.String())
					return true
				})

				// this tells us it's not the same request so no need to proceed
				if params.Encode() != qparams.Encode() {
					continue
				}

				log.Debugf("responding for payload set %s", name)
				respond.Json(w, respd.Value())
				return

			default:
				// TODO:: handle other methods

				respond.Error(w, http.StatusBadRequest, "invalid request")
				return
			}
		}

		respond.Error(w, http.StatusBadRequest, "invalid request")
	}
}

func LoadRoutes(mux *http.ServeMux, dir string) (*http.ServeMux, error) {
	log.Debugf("processing files for dir %s", dir)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return mux, errors.Wrapf(err, "unable to read in %s", dir)
	}

	// pick up multiple JSON files in the configured directory
	// NOTE: not handling nested directories or non-JSON files at this time
	for _, f := range files {
		if f.IsDir() || strings.ToLower(filepath.Ext(f.Name())) != ".json" {
			continue
		}

		// convert to fully qualified path
		fqp := filepath.Join(dir, f.Name())
		log.Debugf("processing file %s", fqp)

		in, err := ioutil.ReadFile(fqp)
		if err != nil {
			return mux, errors.Wrapf(err, "unable to read in file %s", fqp)
		}

		route := gjson.Get(string(in), "route").String()
		if len(route) < 1 {
			return mux, errors.Errorf("invalid route in file %s", fqp)
			continue
		}

		log.Debugf("adding handlers for route %s", route)
		mux.HandleFunc(route, handler(dir, f.Name()))
	}

	return mux, nil
}
