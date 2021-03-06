package holler

import (
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
)

// HollerProxy abstracts the Holler application
type HollerProxy struct {
	Backends  map[string]*Backend
	Port      string
	Log       *logrus.Entry
	LogLevel  logrus.Level
	LogOutput io.Writer
	Server    *http.Server
	sync.Mutex
}

// New returns a new initialized instance of HollerProxy. See options.go for
// functional options to override default configuration.
func New(options ...Option) (*HollerProxy, error) {
	defaultHoller := &HollerProxy{
		Backends:  make(map[string]*Backend),
		Port:      ":9000",
		Log:       logrus.WithFields(logrus.Fields{"holler": "default"}),
		LogLevel:  logrus.DebugLevel,
		LogOutput: os.Stdout,
		Server:    &http.Server{},
	}

	for _, option := range options {
		if err := option(defaultHoller); err != nil {
			return defaultHoller, err
		}
	}

	return defaultHoller, nil
}

// Start assumes that New() was called and HollerProxy has an initialized
// *http.Server, and port setting.
func (h *HollerProxy) Start() {
	logrus.SetLevel(h.LogLevel)
	logrus.SetOutput(h.LogOutput)

	h.Server.Handler = newRouter(h)
	h.Server.Addr = h.Port

	h.Log.Info("starting holler on localhost" + h.Port)
	go func() { h.HealthSupervisor() }()

	h.Log.Error(h.Server.ListenAndServe())
}

func (h *HollerProxy) ReadState() error {
	return nil
}
