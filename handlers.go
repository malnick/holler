package holler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

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
