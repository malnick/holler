package holler

import (
	"net/http"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

const (
	registerPath   = "/register"
	registeredPath = "/registered"
)

type route struct {
	Name        string
	Method      string
	Path        string
	HandlerFunc func(*HollerProxy) http.HandlerFunc
}

// newRouter iterates over a slice of Route types and creates them
// in gorilla/mux.
func newRouter(h *HollerProxy) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	loadRoutes(router, h)
	return router
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
