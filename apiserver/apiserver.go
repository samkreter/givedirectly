package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/samkreter/go-core/httputil"
	"github.com/samkreter/go-core/log"
	"github.com/badoux/checkmail"
)

type Server struct {
	config *Config
}

// ServerConfig configuration for the message API server
type Config struct {
	ServerAddr           string
	EnableReqCorrelation bool
	EnableReqLogging     bool
}

func NewServer(config *Config) (*Server, error) {
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	return &Server{
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

	// add logging/correlation middleware
	middlewareRouter := httputil.SetUpHandler(router, &httputil.HandlerConfig{
		CorrelationEnabled: s.config.EnableReqCorrelation,
		LoggingEnabled:     s.config.EnableReqLogging,
	})

	return middlewareRouter
}

type Request struct {
	Email string
	Title string
}

type Book struct {
	ID int
	Available bool
	Title string
	Timestamp string
}

func (s *Server) handlePostRequest(w http.ResponseWriter, req *http.Request) {
	//logger := log.G(req.Context())
	defer req.Body.Close()

	var request Request
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate title
	if len(request.Title) == 0 {
		http.Error(w, "Must supply a title", http.StatusBadRequest)
	}

	// Validate email
	if err := checkmail.ValidateFormat(request.Email); err != nil {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
	}

	// ISO-8601 formatted date/time
	//time.Now().Format(time.RFC3339)

	w.WriteHeader(http.StatusOK)
}
