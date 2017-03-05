package holler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func getBackendFromRequest(r *http.Request) (*Backend, error) {
	newBackend := &Backend{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return newBackend, err
	}

	if err := json.Unmarshal(body, newBackend); err != nil {
		return newBackend, err
	}

	return newBackend, nil
}

func pingHandler(h *HollerProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Ping!\n"))
	}
}

func registerBackendHandler(h *HollerProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			backends := h.Backends
			json, err := json.Marshal(backends)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Write([]byte(fmt.Sprintf("%s\n", json)))
			return
		}

		backend, err := getBackendFromRequest(r)
		if err != nil {
			h.Log.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if r.Method == "DELETE" {
			h.Log.Debugf("deleting backend %s", backend.NamedRoute)
			if err := h.DeleteBackend(backend); err != nil {
				h.Log.Error(err)
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Write([]byte("deleted backend " + backend.NamedRoute + "\n"))
			return
		}

		h.Log.Debugf("registering new backend %s", backend.NamedRoute)
		if err := h.RegisterBackend(backend); err != nil {
			h.Log.Error(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write([]byte(fmt.Sprintf("registered new holler backend %s\n", backend.NamedRoute)))
	}
}

func registeredBackendsHandler(h *HollerProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.Server.Handler.(*mux.Router).Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			t, err := route.GetPathTemplate()
			if err != nil {
				return err
			}

			w.Write([]byte(t))
			return nil
		})
	}
}
