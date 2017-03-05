package holler

import (
	"errors"
	"net/http"

	"github.com/Sirupsen/logrus"
)

/* Holler proxy functional options */
type Option func(*HollerProxy) error

// HollerPort allows you to override the default Holler port (:9000)
func HollerPort(port string) Option {
	return func(h *HollerProxy) error {
		if len(port) == 0 {
			return errors.New("Holler proxy port option can not be empty")
		}

		h.Port = port
		return nil
	}
}

// HollerServer allows you to set your own instance of *http.Server for Holler to run on
func HollerServer(server *http.Server) Option {
	return func(h *HollerProxy) error {
		if server == nil {
			return errors.New("server option can not be nil")
		}
		h.Server = server
		return nil
	}
}

// HollerLog allows you to override the default logrus entry (useful if you have
// many instances of Holler proxy running
func HollerLog(logger *logrus.Entry) Option {
	return func(h *HollerProxy) error {
		if logger == nil {
			return errors.New("logger option can not be nil")
		}
		h.Log = logger
		return nil
	}
}

// HollerBackends allows you to set backends at start. Useful if you
// are reading backends fromm a baseline config file at start
func HollerBackends(backends map[string]*Backend) Option {
	return func(h *HollerProxy) error {
		if len(backends) == 0 {
			return errors.New("backends option can not be empty")
		}
		h.Backends = backends
		return nil
	}
}
