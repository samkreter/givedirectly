package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/samkreter/givedirectly/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"github.com/samkreter/givedirectly/datastore"
	"github.com/samkreter/givedirectly/apiserver/mockstore"
)

const (
	testTitle = "testTitle"
)

func TestHandlePostRequest(t *testing.T){
	t.Run("Success Case", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockLibraryStore := mockstore.NewMockLibraryStore(mockCtrl)

		s := Server{
			config: &Config{},
			store: mockLibraryStore,
		}

		testServer := httptest.NewServer(s.newRouter())

		request := &types.Request{
			Email: "test@gmail.com",
			Title: testTitle,
		}

		// mock the creatRequest
		mockLibraryStore.EXPECT().CreateRequest(gomock.Any(), request).
			Return(&types.Book{
				ID: 1,
				Title: testTitle,
				Available: true,
				TimeRequested: time.Now().Format(time.RFC3339),
		}, nil).Times(1)

		b, err := json.Marshal(request)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/request", bytes.NewBuffer(b))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should be success status code.")

		defer resp.Body.Close()
		var retBook types.Book
		err = json.NewDecoder(resp.Body).Decode(&retBook)
		require.NoError(t, err)

		assert.Equal(t, retBook.ID, 1, "Should return correct book.")
	})

	t.Run("Book Not Found", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockLibraryStore := mockstore.NewMockLibraryStore(mockCtrl)

		s := Server{
			config: &Config{},
			store: mockLibraryStore,
		}

		testServer := httptest.NewServer(s.newRouter())

		request := &types.Request{
			Email: "test@gmail.com",
			Title: testTitle,
		}

		// mock the creatRequest
		mockLibraryStore.EXPECT().CreateRequest(gomock.Any(), request).
			Return(nil, datastore.ErrNotFound).Times(1)

		b, err := json.Marshal(request)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", testServer.URL+"/request", bytes.NewBuffer(b))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should retrun not found for no book.")
	})

	t.Run("Invalid Email", func(t *testing.T) {
		s := Server{
			config: &Config{},
		}

		testServer := httptest.NewServer(s.newRouter())

		request := types.Request{
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

		request := types.Request{
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