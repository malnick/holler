package holler

import (
	"errors"
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
	NamedRoute          string    `json:"route"`
	ProxyBufferSize     int       `json:"proxy_buffer_size,omitempty"`
	TargetSelector      string    `json:"target_selector,omitempty"`
	Targets             []*Target `json:"targets,omitempty"`
	HealthCheckInterval int       `json:"health_check_interval,omitempty"`
	proxy               *httputil.ReverseProxy
}

// SelectHealthy chooses a healthy target via RRD selection.
func (b *Backend) SelectHealthy() (*Target, error) {
	for _, t := range b.Targets {
		if t.Healthy {
			return t, nil
		}
	}
	return nil, errors.New("no healthy targets")
}

/* HollerProxy methods specific to Backend{} manipulation */

// RegisterBackend adds a new backend to Holler
func (h *HollerProxy) RegisterBackend(b *Backend) error {
	if _, ok := h.Backends[b.NamedRoute]; ok {
		return errors.New("backend " + b.NamedRoute + " already registered, ignoring")
	}

	if b.HealthCheckInterval == 0 {
		b.HealthCheckInterval = 5
	}

	h.Lock()
	defer h.Unlock()

	director := func(req *http.Request) {
		h.Log.Debugf("calling backend director for %s", b.NamedRoute)
		if len(b.Targets) == 0 {
			h.Log.Errorf("targets for backend %s are empty, bailing out", b.NamedRoute)
			return
		}

		// Need leastConn, roundRobin, etc
		target, err := b.SelectHealthy()
		if err != nil {
			h.Log.Error(err)
			return
		}
		targetURL, err := url.Parse(target.URL)
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

// DeleteBackend removes a backend from holler.
// Note:
//   Gorilla mux doesn't have a way to delete a route in memory. At the risk of
//   re-writting that library, here we're building a new http.Handlerfrom our
//   default API routes plus the updated backends with the desired backend deleted.
func (h *HollerProxy) DeleteBackend(b *Backend) error {
	if _, ok := h.Backends[b.NamedRoute]; !ok {
		return errors.New("unable to delete backend " + b.NamedRoute + " does not exist")
	}

	delete(h.Backends, b.NamedRoute)

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
