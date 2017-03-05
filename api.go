package holler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

const (
	registerPath   = "/register"
	registeredPath = "/registered"
)

// Backend abstracts the configuration and targets for a backend
// request. Targets are assumed to be fully qualified url.URL which
// can pass url.Parse(target).
// TargetSelector can by one of: random, roundrobin.
// If ProxyBuffer settings are nil, no buffering occurs.
type Backend struct {
	NamedRoute       string   `json:"route"`
	ProxyBufferSize  int      `json:"proxy_buffer_size,omitempty"`
	ProxyBufferAlloc int      `json:"proxy_buffer_alloc,omitempty"`
	TargetSelector   string   `json:"target_selector,omitempty"`
	Targets          []string `json:"targets,omitempty"`
	proxy            *httputil.ReverseProxy
}

type route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc func(*HollerProxy) http.HandlerFunc
}

func pingHandler(h *HollerProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ping!\n"))
	}
}

func registerBackendHandler(h *HollerProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newBackend := &Backend{}
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		if err := json.Unmarshal(body, newBackend); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		h.Log.Infof("registering new backend %+v", newBackend)

		if err := h.RegisterBackend(newBackend); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		w.Write([]byte(fmt.Sprintf("registered %+v\n", newBackend)))
	}
}

func registeredBackendsHandler(h *HollerProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Registered backends!\n"))
	}
}

// logger() wraps every API request for mapping to HollerProxy.Log entry
func logger(inner http.Handler, name string, httpLog *logrus.Entry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		httpLog.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

// loadRoutes() automatically generates gorilla mux routes from our defined route{} type
func loadRoutes(router *mux.Router, h *HollerProxy) {
	var routes = []route{
		route{
			Name:        "ping",
			Method:      "GET",
			Path:        "/",
			HandlerFunc: pingHandler,
		},

		route{
			Name:        "/register/backend",
			Method:      "POST",
			Path:        strings.Join([]string{registerPath, "backend"}, "/"),
			HandlerFunc: registerBackendHandler,
		},

		route{
			Name:        "/registered/backends",
			Method:      "GET",
			Path:        strings.Join([]string{registeredPath, "backends"}, "/"),
			HandlerFunc: registeredBackendsHandler,
		},
	}

	for _, r := range routes {
		h.Log.Infof("establishing Hollar API endpoint %s at %s", r.Name, r.Path)
		var handler http.Handler

		handler = r.HandlerFunc(h)
		handler = logger(handler, r.Name, h.Log)

		router.NewRoute().
			Methods(r.Method).
			Path(r.Path).
			Name(r.Name).
			Handler(handler)
	}
}

// NewRouter iterates over a slice of Route types and creates them
// in gorilla/mux.
func newRouter(h *HollerProxy) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	loadRoutes(router, h)
	return router
}
