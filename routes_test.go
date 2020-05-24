package stuber

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

const (
	dir  = "./test_data"
	file = "sample.json"
)

func TestHandler(t *testing.T) {
	stubs, err := ioutil.ReadFile(filepath.Join(dir, file))
	require.NoError(t, err)

	tests := map[string]func(t *testing.T){
		"error": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, ts.URL, nil)
			require.NoError(t, err)

			// this should be method not allowed
			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			del := gjson.Get(string(stubs), `stubs.#(name=="sample-4")`)

			require.Equal(t, int(del.Get("response.error.code").Int()), res.StatusCode)
			require.Equal(t, del.Get("response.error.message").String(), strings.TrimSpace(string(b)))
		},
		"get/empty-params": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-3")`)
			require.False(t, stub.Get(`request.data`).Exists())
			require.Equal(t, "data.array", stub.Get(`response.type`).String())

			a := make([]interface{}, 0)
			require.NoError(t, json.Unmarshal([]byte(stub.Get(`response.data`).String()), &a))

			out, err := json.Marshal(a)
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},
		"post/created": func(t *testing.T) {

		},
		"post/some-params": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-1")`)

			res, err := http.Post(ts.URL, stub.Get(`request.content-type`).String(), strings.NewReader(stub.Get(`request.data`).String()))
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			require.True(t, stub.Get(`request.data`).Exists())
			require.Equal(t, "data.object", stub.Get(`response.type`).String())

			m := map[string]interface{}{}
			require.NoError(t, json.Unmarshal([]byte(stub.Get(`response.data`).String()), &m))

			out, err := json.Marshal(m)
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},

		"get/some-params": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			u := fmt.Sprintf("%s?", ts.URL)
			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-0")`)
			stub.Get(`request.data`).ForEach(func(k, v gjson.Result) bool {
				u = fmt.Sprintf("%s%s=%s&", u, k, v)
				return true
			})

			res, err := http.Get(u)
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			require.True(t, stub.Get(`request.data`).Exists())
			require.Equal(t, "data.object", stub.Get(`response.type`).String())

			m := map[string]interface{}{}
			require.NoError(t, json.Unmarshal([]byte(stub.Get(`response.data`).String()), &m))

			out, err := json.Marshal(m)
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},
	}

	for n, test := range tests {
		t.Run(n, test)
	}
}

func TestLoadRoutes(t *testing.T) {

}
