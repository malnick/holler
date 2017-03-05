package holler

import (
	"errors"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

type HollerProxy struct {
	Backends map[string]*Backend
	Port     string
	Log      *logrus.Entry
	Server   *http.Server
	sync.Mutex
}

func New(options ...Option) (*HollerProxy, error) {
	defaultHoller := &HollerProxy{
		Backends: make(map[string]*Backend),
		Port:     ":9000",
		Log:      logrus.WithFields(logrus.Fields{"holler": "default"}),
		Server:   &http.Server{},
	}

	for _, option := range options {
		if err := option(defaultHoller); err != nil {
			return defaultHoller, err
		}
	}

	return defaultHoller, nil
}

func (h *HollerProxy) Start() {
	h.Server.Handler = newRouter(h)
	h.Server.Addr = h.Port

	h.Log.Info("starting holler on localhost" + h.Port)
	h.Log.Error(h.Server.ListenAndServe())
}

func (h *HollerProxy) RegisterBackend(b *Backend) error {
	// Lock Holler before updating backend
	h.Lock()
	defer h.Unlock()

	director := func(req *http.Request) {
		h.Log.Infof("calling backend director for %s", b.NamedRoute)
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

		h.Log.Infof("making backend request for %s:\n    Scheme %s\n    Host %s\n    Path %s", b.NamedRoute, targetURL.Scheme, targetURL.Host, targetURL.Path)
		req.URL.Scheme = targetURL.Scheme
		req.URL.Host = targetURL.Host
		req.URL.Path = targetURL.Path
	}

	b.proxy = &httputil.ReverseProxy{
		Director: director,
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				h.Log.Infof("making backend request to %s", req.URL.Host)
				return http.ProxyFromEnvironment(req)
			},
			//			Dial: func(network, addr string) (net.Conn, error) {
			//				h.Log.Info("CALLING DIAL")
			//				conn, err := (&net.Dialer{
			//					Timeout:   30 * time.Second,
			//					KeepAlive: 30 * time.Second,
			//				}).Dial(network, addr)
			//				if err != nil {
			//					h.Log.Error("Error during DIAL:", err.Error())
			//				}
			//				return conn, err
			//			},
			//			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	if _, ok := h.Backends[b.NamedRoute]; ok {
		return errors.New("backend " + b.NamedRoute + " already registered, ignoring")
	}

	h.Backends[b.NamedRoute] = b
	h.Log.Infof("establishing backend %s\n    Targets: %+v", b.NamedRoute, b.Targets)
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

	h.Log.Infof("deleteing backend %s", b.NamedRoute)
	delete(h.Backends, b.NamedRoute)
	return nil
}

func (h *HollerProxy) ReadState() error {
	return nil
}

/* Holler proxy functional options */
type Option func(*HollerProxy) error

func HollerPort(port string) Option {
	return func(h *HollerProxy) error {
		if len(port) == 0 {
			return errors.New("Holler proxy port option can not be empty")
		}

		h.Port = port
		return nil
	}
}

func HollerLog(logger *logrus.Entry) Option {
	return func(h *HollerProxy) error {
		if logger == nil {
			return errors.New("Logger option can not be nil")
		}
		h.Log = logger
		return nil
	}
}
