package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"
)

func TestHandlePostRequest(t *testing.T){
	t.Run("Success Case", func(t *testing.T) {
		s := Server{
			config: &Config{},
		}

		testServer := httptest.NewServer(s.newRouter())

		request := Request{
			Email:"test@gmail.com",
			Title: "testTitle",
		}

		b, err := json.Marshal(request)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/request", bytes.NewBuffer(b))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should be success status code.")
	})

	t.Run("Invalid Email", func(t *testing.T) {
		s := Server{
			config: &Config{},
		}

		testServer := httptest.NewServer(s.newRouter())

		request := Request{
			Email:"",
			Title: "testTitle",
		}

		b, err := json.Marshal(request)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/request", bytes.NewBuffer(b))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should be success status code.")
	})

	t.Run("Invalid Title", func(t *testing.T) {
		s := Server{
			config: &Config{},
		}

		testServer := httptest.NewServer(s.newRouter())

		request := Request{
			Email: "test@gmail.com",
			Title: "",
		}

		b, err := json.Marshal(request)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/request", bytes.NewBuffer(b))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should be success status code.")
	})
}