package stuber

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/elliottpolk/stuber/internal/respond"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

func data(t, d string) (interface{}, error) {
	switch t {
	case "data.array":
		a := make([]interface{}, 0)
		if err := json.Unmarshal([]byte(d), &a); err != nil {
			return nil, err
		}

		return a, nil

	case "data.object":
		m := map[string]interface{}{}
		if err := json.Unmarshal([]byte(d), &m); err != nil {
			return nil, err
		}

		return m, nil
	}

	return nil, errors.New("unsupported type")
}

func handler(d, f string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		in, err := ioutil.ReadFile(filepath.Join(d, f))
		if err != nil {
			respond.Error(w, http.StatusInternalServerError, err.Error())
			return
		}

		stubs := gjson.Get(string(in), "stubs").Array()
		for _, stub := range stubs {
			m := stub.Get("request.method").String()
			if r.Method != m {
				continue
			}

			if stub.Get("response.type").String() == "error" {
				respond.Error(w, int(stub.Get("response.error.code").Int()), stub.Get("response.error.message").String())
				return
			}

			reqd := stub.Get("request.data")

			// handle empty request data
			if !reqd.Exists() {
				if len(r.URL.Query().Encode()) > 0 {
					respond.Error(w, http.StatusBadRequest, "invalid request")
					return
				}

				sd, err := data(stub.Get("response.type").String(), stub.Get("response.data").String())
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				log.Debugf("responding for data set %s", stub.Get("name").String())
				respond.Json(w, sd)
				return
			}

			// if we get here, we have some request data
			// POST vs GET are handled differently
			switch m {
			case http.MethodPost:
				// check the body
				rb, err := ioutil.ReadAll(r.Body)
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				// force a cleanup on the request body to match the file data
				var i interface{}
				if err := json.Unmarshal(rb, &i); err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				rb, err = json.Marshal(i)
				if err := json.Unmarshal(rb, &i); err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				reqd, err := data(stub.Get("request.type").String(), stub.Get("request.data").String())
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				out, err := json.Marshal(reqd)
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				if string(out) != string(rb) {
					continue
				}

				// TODO: handle `type == "data.empty"`

				sd, err := data(stub.Get("response.type").String(), stub.Get("response.data").String())
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				log.Debugf("responding for data set %s", stub.Get("name").String())
				respond.Json(w, sd)
				return

			case http.MethodGet:
				// check the URL params
				params := &url.Values{}
				reqd.ForEach(func(k, v gjson.Result) bool {
					params.Add(k.String(), v.String())
					return true
				})

				if params.Encode() != r.URL.Query().Encode() {
					continue
				}

				sd, err := data(stub.Get("response.type").String(), stub.Get("response.data").String())
				if err != nil {
					respond.Error(w, http.StatusInternalServerError, err.Error())
					return
				}

				log.Debugf("responding for data set %s", stub.Get("name").String())
				respond.Json(w, sd)
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
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return mux, errors.Wrapf(err, "unable to read in %s", dir)
	}

	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != "json" {
			continue
		}

		in, err := ioutil.ReadFile(filepath.Join(dir, f.Name()))
		if err != nil {
			log.Error(err)
			continue
		}

		route := gjson.Get(string(in), "route").String()
		if len(route) < 1 {
			log.Errorf("invalid route for file %s", filepath.Join(dir, f.Name()))
			continue
		}

		mux.HandleFunc(route, handler(dir, f.Name()))
	}

	return mux, nil
}
