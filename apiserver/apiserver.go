package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/samkreter/givedirectly/datastore"
	"net/http"
	"strconv"

	"github.com/badoux/checkmail"
	"github.com/gorilla/mux"
	"github.com/samkreter/go-core/httputil"
	"github.com/samkreter/go-core/log"

	"github.com/samkreter/givedirectly/types"
)

//go:generate sh -c "mockgen -package=mockstore github.com/samkreter/givedirectly/apiserver LibraryStore >./mockstore/mock_librarystore.go"

type LibraryStore interface {
	CreateRequest(ctx context.Context, request *types.Request) (*types.Book, error)
	GetRequest(ctx context.Context, requestID int)  (*types.Request, error)
	ListRequest(ctx context.Context)  ([]*types.Request, error)
	DeleteRequest(ctx context.Context, requestID int) error
}

type Server struct {
	config *Config
	store LibraryStore
}

// ServerConfig configuration for the message API server
type Config struct {
	ServerAddr           string
	EnableReqCorrelation bool
	EnableReqLogging     bool
}

func NewServer(store LibraryStore, config *Config) (*Server, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &Server{
		store: store,
		config:  config,
	}, nil
}

func (s *Server) Run() error {
	router := s.newRouter()

	log.G(context.TODO()).WithField("address: ", s.config.ServerAddr).Info("Starting Request API Server:")
	if err := http.ListenAndServe(s.config.ServerAddr, router); err != nil {
		return err
	}

	return nil
}

func validateConfig(config *Config) error {
	if config == nil {
		return errors.New("missing server configuration")
	}

	if config.ServerAddr == "" {
		return errors.New("must supply API servering address")
	}

	return nil
}

func (s *Server) newRouter() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/request", s.handlePostRequest).Methods("POST")
	router.HandleFunc("/request", s.handleListRequest).Methods("GET")
	router.HandleFunc("/request/{id}", s.handleGetRequest).Methods("GET")
	router.HandleFunc("/request/{id}", s.handleDeleteRequest).Methods("DELETE")

	// add logging/correlation middleware
	middlewareRouter := httputil.SetUpHandler(router, &httputil.HandlerConfig{
		CorrelationEnabled: s.config.EnableReqCorrelation,
		LoggingEnabled:     s.config.EnableReqLogging,
	})

	return middlewareRouter
}

func (s *Server) handlePostRequest(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	ctx := req.Context()
	logger := log.G(req.Context())

	var request *types.Request
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Validate title
	if len(request.Title) == 0 {
		http.Error(w, "Must supply a title", http.StatusBadRequest)
		return
	}

	// Validate email
	if err := checkmail.ValidateFormat(request.Email); err != nil {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	book, err := s.store.CreateRequest(ctx, request)
	if err != nil {
		switch {
		case err == datastore.ErrNotFound:
			http.Error(w, "Requested book not found", http.StatusNotFound)
			return
		default:
			logger.Errorf("failed to create request with error: %v", err)
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}
	}

	if err := json.NewEncoder(w).Encode(book); err != nil{
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Errorf("handlePostRequest: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleListRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.G(ctx)


	requests, err := s.store.ListRequest(ctx)
	if err != nil {
		logger.Errorf("failed to list requests with error: %v", err)
		http.Error(w, "failed with internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(requests)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Errorf("handleGetMessage: %v", err)
		return
	}
}

func (s *Server) handleGetRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.G(ctx)
	vars := mux.Vars(req)

	requestIDStr := vars["id"]
	if requestIDStr == "" {
		http.Error(w, "missing request id", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}

	request, err := s.store.GetRequest(ctx, requestID)
	if err != nil {
		switch {
		case err == datastore.ErrNotFound:
			http.Error(w, "request not found", http.StatusNotFound)
			return
		default:
			logger.Errorf("failed to get request with error: %v", err)
			http.Error(w, "failed with internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(request)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Errorf("handleGetMessage: %v", err)
		return
	}
}

func (s *Server) handleDeleteRequest(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	logger := log.G(ctx)
	vars := mux.Vars(req)

	requestIDStr := vars["id"]
	if requestIDStr == "" {
		http.Error(w, "missing request id", http.StatusBadRequest)
		return
	}

	requestID, err := strconv.Atoi(requestIDStr)
	if err != nil {
		http.Error(w, "invalid request id", http.StatusBadRequest)
		return
	}

	err = s.store.DeleteRequest(ctx, requestID)
	if err != nil {
		switch {
		case err == datastore.ErrNotFound:
			http.Error(w, "request not found", http.StatusNotFound)
			return
		default:
			logger.Errorf("failed to delete request with error: %v", err)
			http.Error(w, "failed with internal server error", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}
