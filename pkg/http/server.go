package rest

import (
	"encoding/json"
	"fmt"
	moduleDto "github.com/NerdShoreDev/YEP/server/pkg/module/dto"
	registryDto "github.com/NerdShoreDev/YEP/server/pkg/registry/dto"
	"net/http"
	"time"

	sentrynegroni "github.com/getsentry/sentry-go/negroni"
	"github.com/urfave/negroni"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const CONTENT_TYPE_JSON = "application/json"
const CONTENT_TYPE_KEY = "Content-Type"

type ServiceHandler interface {
	DeleteModule(name string) error
	FindModule(name string) (*moduleDto.ServerModule, error)
	RequestUpsertModule(cm moduleDto.RequestModule) error
	GetRegistryServerConfig() (*registryDto.RegistryServerConfig, error)
	GetClientConfig() (*registryDto.RegistryClientConfig, error)
	ValidateAuthorizedUserForDeletion(token string) bool
	ValidateJWTToken(tokenString string) error
	ValidateUpsertModule(moduleProspect *moduleDto.RequestModule) error
}

type WebServer interface {
	StartWebServer(serviceHandler ServiceHandler)
}

type webServer struct {
	allowedOrigins string
}

func NewWebServer(allowedOrigins string) *webServer {
	return &webServer{allowedOrigins: allowedOrigins}
}

func (wS *webServer) StartWebServer(serviceHandler ServiceHandler) {
	middlewareManager := negroni.New()
	middlewareManager.Use(sentrynegroni.New(sentrynegroni.Options{}))
	middlewareManager.UseHandler(wS.getRouter(serviceHandler))

	srv := &http.Server{
		Handler: middlewareManager,
		Addr:    ":3000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

func (wS *webServer) getRouter(s ServiceHandler) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/auth/realm/:realm/.well-known/openid-configuration", errorBump).Methods(http.MethodGet)
	router.HandleFunc("/auth/realm/:realm/protocol/openid-connect/auth", errorBump).Methods(http.MethodGet)
	router.HandleFunc("/auth/realm/:realm/protocol/openid-connect/token", errorBump).Methods(http.MethodGet)
	router.HandleFunc("/auth/realm/:realm/protocol/openid-connect/userinfo", errorBump).Methods(http.MethodGet)
	router.HandleFunc("/auth/realm/:realm/protocol/openid-connect/logout", errorBump).Methods(http.MethodGet)
	router.HandleFunc("/auth/realm/:realm/protocol/openid-connect/certs", errorBump).Methods(http.MethodGet)
	router.Handle(INTERNAL_API_ROUTE_PREFIX+"/metrics", promhttp.Handler())
	router.HandleFunc("/api/health", healthCheck).Methods(http.MethodGet)
	return router
}

func (wS *webServer) setupResponse(w *http.ResponseWriter, req *http.Request) {
	log.Debugln("Request logged from: ", req.RequestURI)
	(*w).Header().Set("Access-Control-Allow-Origin", wS.allowedOrigins)
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "User-Agent, Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Host")
	(*w).Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
	(*w).Header().Set(CONTENT_TYPE_KEY, CONTENT_TYPE_JSON)
}

func errorBump(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	var message string
	if subject, ok := r.URL.Query()["subject"]; ok {
		message = fmt.Sprintf("%s requested with subject '%s'", r.URL.Path, subject[0])
	} else {
		message = fmt.Sprintf("%s requested without subject", r.URL.Path)
	}
	panic(message)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func upsertModule(s ServiceHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CONTENT_TYPE_KEY, CONTENT_TYPE_JSON)

		// Validate JWT access token
		if err := s.ValidateJWTToken(r.Header.Get("Authorization")); err != nil {
			log.Debugf("upsertModule: JWT validation error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": fmt.Sprint(err)})
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)

		var requestModule moduleDto.RequestModule
		if err := decoder.Decode(&requestModule); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "could not parse module meta data"})
			return
		}

		if err := s.ValidateUpsertModule(&requestModule); err != nil {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": err.Error()})
			return
		}

		if err := s.RequestUpsertModule(requestModule); err != nil {
			log.Errorf("new module could not be saved: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "new module could not be saved"})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func readModule(s ServiceHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CONTENT_TYPE_KEY, CONTENT_TYPE_JSON)

		// Validate JWT access token
		if err := s.ValidateJWTToken(r.Header.Get("Authorization")); err != nil {
			log.Debugf("readModule: JWT validation error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": fmt.Sprint(err)})
			return
		}

		name := r.URL.Query()["name"]
		if len(name) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "name required"})
			return
		}
		module, err := s.FindModule(name[0])
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "module with name: " + name[0] + " could not be found."})
				return
			}
			log.Errorf("error during reading of module %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "error during reading of module"})
			return
		}
		json.NewEncoder(w).Encode(module)
	}
}

func deleteModule(s ServiceHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CONTENT_TYPE_KEY, CONTENT_TYPE_JSON)

		// Validate JWT access token
		if err := s.ValidateJWTToken(r.Header.Get("Authorization")); err != nil {
			log.Debugf("deleteModule: JWT validation error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": fmt.Sprint(err)})
			return
		}

		// Validate user token in header for authorization
		if authorized := s.ValidateAuthorizedUserForDeletion(r.Header.Get("Authorized-User")); !authorized {
			log.Debugf("user not allowed to delete a module")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "user not allowed to delete a module"})
			return
		}

		// Get module name from url query
		name := r.URL.Query()["name"]
		if len(name) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "name required"})
			return
		}

		// Request module deletion
		err := s.DeleteModule(name[0])
		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "module with name: " + name[0] + " could not be found."})
				return
			}
			log.Errorf("error during deleting of module: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "error during deleting of module"})
			return
		}

		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	}
}

func readRegistryConfig(s ServiceHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CONTENT_TYPE_KEY, CONTENT_TYPE_JSON)

		// Validate JWT access token
		if err := s.ValidateJWTToken(r.Header.Get("Authorization")); err != nil {
			log.Debugf("readRegistryConfig: JWT validation error: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": fmt.Sprint(err)})
			return
		}

		registryConfig, err := s.GetRegistryServerConfig()

		if err != nil {
			if err.Error() == "server: unable to retrieve registry data" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "could not registry data"})
				return
			}
			log.Errorf("error during reading of registry: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "error during reading of registry"})
			return
		}

		json.NewEncoder(w).Encode(registryConfig)
	}
}

func (wS *webServer) readClientConfig(s ServiceHandler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		wS.setupResponse(&w, r)

		clientConfig, err := s.GetClientConfig()

		if err != nil {
			if err.Error() == "server: unable to retrieve client config data" {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "could not retrieve client config data"})
				return
			}
			log.Errorf("error during reading client config: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{"ok": false, "message": "error during reading of client config"})
			return
		}

		json.NewEncoder(w).Encode(clientConfig)
	}
}
