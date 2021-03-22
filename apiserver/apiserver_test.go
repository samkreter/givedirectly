package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"fmt"

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


func TestHandleGetRequest(t *testing.T){
	t.Run("Success Case", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockLibraryStore := mockstore.NewMockLibraryStore(mockCtrl)

		s := Server{
			config: &Config{},
			store: mockLibraryStore,
		}

		testServer := httptest.NewServer(s.newRouter())

		testRequestID := 123

		// mock the creatRequest
		mockLibraryStore.EXPECT().GetRequest(gomock.Any(), testRequestID).
			Return(&types.Request{
				Email: "test@gmail.com",
				Title: testTitle,
				ID: testRequestID,
			}, nil).Times(1)

		url := fmt.Sprintf("%s/%d", testServer.URL+"/request", testRequestID)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should be success status code.")

		defer resp.Body.Close()
		var retRequest types.Request
		err = json.NewDecoder(resp.Body).Decode(&retRequest)
		require.NoError(t, err)

		assert.Equal(t, retRequest.ID, testRequestID, "Should return correct request.")
	})

	t.Run("Request Not Found", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockLibraryStore := mockstore.NewMockLibraryStore(mockCtrl)

		s := Server{
			config: &Config{},
			store: mockLibraryStore,
		}

		testServer := httptest.NewServer(s.newRouter())

		testRequestID := 123

		mockLibraryStore.EXPECT().GetRequest(gomock.Any(), testRequestID).
			Return(nil, datastore.ErrNotFound).Times(1)

		url := fmt.Sprintf("%s/%d", testServer.URL+"/request", testRequestID)
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return not found for no request.")
	})

	t.Run("Invalid Request ID", func(t *testing.T) {
		s := Server{
			config: &Config{},
		}

		testServer := httptest.NewServer(s.newRouter())


		url := fmt.Sprintf("%s/%s", testServer.URL+"/request", "invalidReqeustID")
		req, err := http.NewRequest("GET", url, nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should be badrequest status code.")
	})
}

func TestHandleListRequests(t *testing.T) {
	t.Run("Success Case", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockLibraryStore := mockstore.NewMockLibraryStore(mockCtrl)

		s := Server{
			config: &Config{},
			store:  mockLibraryStore,
		}

		testServer := httptest.NewServer(s.newRouter())

		retRequests := []*types.Request{}
		for i:=0;i<10;i++ {
			retRequests = append(retRequests, &types.Request{
				Title: fmt.Sprintf("test%d", i),
				Email: "testemail",
			})
		}

		// mock the creatRequest
		mockLibraryStore.EXPECT().ListRequest(gomock.Any()).
			Return(retRequests, nil).Times(1)

		req, err := http.NewRequest("GET", testServer.URL+"/request", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusOK, resp.StatusCode, "Should be success status code.")

		defer resp.Body.Close()
		var requests []*types.Request
		err = json.NewDecoder(resp.Body).Decode(&requests)
		require.NoError(t, err)

		assert.Equal(t, len(requests), 10, "Should return correct num of requests.")
	})

	t.Run("Datastore error", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		mockLibraryStore := mockstore.NewMockLibraryStore(mockCtrl)

		s := Server{
			config: &Config{},
			store:  mockLibraryStore,
		}

		testServer := httptest.NewServer(s.newRouter())

		// mock the creatRequest
		mockLibraryStore.EXPECT().ListRequest(gomock.Any()).
			Return(nil, fmt.Errorf("random error")).Times(1)

		req, err := http.NewRequest("GET", testServer.URL+"/request", nil)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "Should be internal error status code.")
	})
}