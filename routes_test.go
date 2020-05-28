package stuber

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

	tests := map[string]func(*testing.T){
		"PUT/some-params/empty-response": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-5")`)

			req, err := http.NewRequest(http.MethodPut, ts.URL, strings.NewReader(stub.Get(`request.payload`).String()))
			require.NoError(t, err)

			req.Header.Set("Content-Type", stub.Get(`request.content-type`).String())

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			require.Equal(t, int(stub.Get(`response.payload.code`).Int()), res.StatusCode)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)
			require.Empty(t, strings.TrimSpace(string(b)))
		},
		"DELETE/error": func(t *testing.T) {
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

			require.Equal(t, int(del.Get("response.payload.code").Int()), res.StatusCode)
			require.Equal(t, del.Get("response.payload.message").String(), strings.TrimSpace(string(b)))
		},
		"GET/empty-params": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-3")`)
			require.False(t, stub.Get(`request.payload`).Exists())
			require.Equal(t, "data.array", stub.Get(`response.type`).String())

			out, err := json.Marshal(stub.Get(`response.payload`).Value())
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},
		"POST/some-params/empty-response": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-2")`)

			res, err := http.Post(ts.URL, stub.Get(`request.content-type`).String(), strings.NewReader(stub.Get(`request.payload`).String()))
			require.NoError(t, err)

			require.Equal(t, int(stub.Get(`response.payload.code`).Int()), res.StatusCode)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)
			require.Empty(t, strings.TrimSpace(string(b)))
		},
		"POST/some-params": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-1")`)

			res, err := http.Post(ts.URL, stub.Get(`request.content-type`).String(), strings.NewReader(stub.Get(`request.payload`).String()))
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			require.True(t, stub.Get(`request.payload`).Exists())
			require.Equal(t, "data.object", stub.Get(`response.type`).String())

			out, err := json.Marshal(stub.Get(`response.payload`).Value())
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},

		"GET/some-params": func(t *testing.T) {
			ts := httptest.NewServer(handler(dir, file))
			defer ts.Close()

			u := fmt.Sprintf("%s?", ts.URL)
			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-0")`)
			stub.Get(`request.payload`).ForEach(func(k, v gjson.Result) bool {
				u = fmt.Sprintf("%s%s=%s&", u, k, v)
				return true
			})

			res, err := http.Get(u)
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			require.True(t, stub.Get(`request.payload`).Exists())
			require.Equal(t, "data.object", stub.Get(`response.type`).String())

			out, err := json.Marshal(stub.Get(`response.payload`).Value())
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},
	}

	for n, test := range tests {
		t.Run(n, test)
	}
}

func duplicateTestData(t *testing.T) string {
	td, err := ioutil.TempDir(".", "")
	require.NoError(t, err)

	tf1, err := ioutil.TempFile(td, "sample_duplicate.*.json")
	require.NoError(t, err)
	defer tf1.Close()

	tf2, err := ioutil.TempFile(td, "sample_duplicate.*.json")
	require.NoError(t, err)
	defer tf2.Close()

	of, err := os.Open(filepath.Join(dir, file))
	require.NoError(t, err)
	defer of.Close()

	_, err = io.Copy(tf1, of)
	require.NoError(t, err)
	require.NoError(t, tf1.Sync())

	_, err = io.Copy(tf2, of)
	require.NoError(t, err)
	require.NoError(t, tf2.Sync())

	t.Cleanup(func() {
		require.NoError(t, os.RemoveAll(td))
	})

	return td
}

func TestLoadRoutes(t *testing.T) {
	stubs, err := ioutil.ReadFile(filepath.Join(dir, file))
	require.NoError(t, err)

	tests := map[string]func(*testing.T){
		"POSITIVE/basic-load": func(t *testing.T) {
			mux, err := LoadRoutes(http.NewServeMux(), dir)
			require.NoError(t, err)

			ts := httptest.NewServer(mux)

			res, err := http.Get(fmt.Sprintf("%s%s", ts.URL, gjson.Get(string(stubs), "route").String()))
			require.NoError(t, err)

			b, err := ioutil.ReadAll(res.Body)
			require.NoError(t, res.Body.Close())
			require.NoError(t, err)

			stub := gjson.Get(string(stubs), `stubs.#(name=="sample-3")`)
			require.False(t, stub.Get(`request.payload`).Exists())
			require.Equal(t, "data.array", stub.Get(`response.type`).String())

			out, err := json.Marshal(stub.Get(`response.payload`).Value())
			require.NoError(t, err)

			require.Equal(t, strings.TrimSpace(string(out)), strings.TrimSpace(string(b)))
		},
		"NEGATIVE/duplicate-routes": func(t *testing.T) {
			td := duplicateTestData(t)

			_, err := LoadRoutes(http.NewServeMux(), td)
			require.Error(t, err)
		},
	}

	for n, test := range tests {
		t.Run(n, test)
	}
}
