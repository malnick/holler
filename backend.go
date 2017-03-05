package holler

import (
	"errors"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/oxtoacart/bpool"
)

// Backend abstracts the configuration and targets for a backend
// request. Targets are assumed to be fully qualified url.URL which
// can pass url.Parse(target).
// TargetSelector can by one of: random, roundrobin.
// If ProxyBuffer settings are nil, no buffering occurs.
type Backend struct {
	NamedRoute      string   `json:"route"`
	ProxyBufferSize int      `json:"proxy_buffer_size,omitempty"`
	TargetSelector  string   `json:"target_selector,omitempty"`
	Targets         []string `json:"targets,omitempty"`
	proxy           *httputil.ReverseProxy
}

func (h *HollerProxy) RegisterBackend(b *Backend) error {
	if _, ok := h.Backends[b.NamedRoute]; ok {
		return errors.New("backend " + b.NamedRoute + " already registered, ignoring")
	}

	// Lock Holler before updating backend
	h.Lock()
	defer h.Unlock()

	director := func(req *http.Request) {
		h.Log.Debugf("calling backend director for %s", b.NamedRoute)
		if len(b.Targets) == 0 {
			h.Log.Errorf("targets for backend %s are empty, bailing out", b.NamedRoute)
			return
		}

		// Need leastConn, roundRobin, etc
		target := b.Targets[rand.Int()%len(b.Targets)]
		targetURL, err := url.Parse(target)
		if err != nil {
			h.Log.Error(err)
			return
		}

		h.Log.Debugf("making backend request for %s:\n    Scheme %s\n    Host %s\n    Path %s", b.NamedRoute, targetURL.Scheme, targetURL.Host, targetURL.Path)
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = targetURL.Path
	}

	b.proxy = &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				h.Log.Debugf("making backend request to %s", req.URL.Host)
				return http.ProxyFromEnvironment(req)
			},
		},
	}

	// If ProxyBufferSize is set, allocate a new pool with max size and new size set to the
	// configured size
	if b.ProxyBufferSize != 0 {
		b.proxy.BufferPool = bpool.NewBytePool(b.ProxyBufferSize, b.ProxyBufferSize)
	}

	h.Backends[b.NamedRoute] = b
	h.Log.Debugf("establishing backend %s\n    Targets: %+v", b.NamedRoute, b.Targets)
	h.Server.Handler.(*mux.Router).NewRoute().
		Name(b.NamedRoute).
		Path(b.NamedRoute).
		Handler(b.proxy)

	return nil
}

func (h *HollerProxy) DeleteBackend(b *Backend) error {
	if _, ok := h.Backends[b.NamedRoute]; !ok {
		return errors.New("unable to delete backend " + b.NamedRoute + " does not exist")
	}

	delete(h.Backends, b.NamedRoute)

	// Gorilla mux doesn't have a way to delete a route in memory. At the risk of
	// re-writting that library, here we're building a new http.Handler
	// from our default API routes plus the updated backends with the
	// desired backend deleted.
	h.Server.Handler = newRouter(h)

	for n, b := range h.Backends {
		h.Log.Debugf("re-registering backend %s", n)
		h.Server.Handler.(*mux.Router).NewRoute().
			Name(b.NamedRoute).
			Path(b.NamedRoute).
			Handler(b.proxy)
	}

	return nil
}

type HollerHandler struct {
	Routes []*mux.Route
}
