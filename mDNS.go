package mdns

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync/atomic"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	lookupName       = "_ankivector._tcp"
	domainName       = "local."
	escapepodSrvName = "escapepod"
)

type mDNSService struct {
	port         int
	ip           string
	lookupPeriod time.Duration

	logger      *slog.Logger
	vektorFound atomic.Bool
	done        chan struct{}
	register    *zeroconf.Server
}

func NewmDNSService(parentLog *slog.Logger, ip string, port int, lookupPeriod time.Duration) (*mDNSService, error) {
	logger := parentLog.With("component", "mdns")
	return &mDNSService{logger: logger, port: port, ip: ip, lookupPeriod: lookupPeriod, done: make(chan struct{})}, nil
}

func (srv *mDNSService) Lookup(ctx context.Context) (bool, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return false, fmt.Errorf("can't create mDNS service: %w", err)
	}

	entries := make(chan *zeroconf.ServiceEntry)
	err = resolver.Browse(ctx, lookupName, domainName, entries)
	if err != nil {
		return false, fmt.Errorf("failed to browse mDSN: %w", err)
	}

	for entry := range entries {
		if strings.Contains(entry.Service, "ankivector") {
			srv.logger.Debug("Vector discovered on network")
			srv.vektorFound.Store(true)
			return true, nil
		}
	}

	srv.vektorFound.Store(false)
	srv.logger.Debug("Vektror wasn't found")
	return false, nil
}

// Perform lookup each minute
func (srv *mDNSService) StartLookup() {
	srv.logger.Info("Starting lookup service", "period", srv.lookupPeriod.String())
	ticker := time.NewTicker(srv.lookupPeriod)

	go func() {
		for {
			select {

			case <-ticker.C:
				srv.logger.Debug("Looking up vektor...")
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				found, err := srv.Lookup(ctx)
				if err != nil {
					srv.logger.Warn("can't lookup for vektor: %w", err)
					break
				}

				if found {
					if err = srv.StartRegisterServer(); err != nil {
						log.Println(err)
					}
				} else {
					srv.StopRegisterServer()
				}
			case <-srv.done:
				srv.logger.Info("done signal was sent")
				ticker.Stop()
				return
			}
		}
	}()
}

func (srv *mDNSService) StartRegisterServer() error {
	if srv.register != nil {
		srv.logger.Debug("server is already registered")
		return nil
	}

	server, err := zeroconf.RegisterProxy(escapepodSrvName, "_app-proto._tcp", domainName, srv.port, escapepodSrvName,
		[]string{srv.ip}, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		return fmt.Errorf("can't register wirepod service")
	}

	srv.register = server
	srv.logger.Info("mDNS service is registered")
	return nil
}

func (srv *mDNSService) StopRegisterServer() {
	if srv.register == nil {
		srv.logger.Debug("server is not created. Return")
		return
	}

	srv.register.Shutdown()
	srv.register = nil
	srv.logger.Info("mDSN server is unregistered")
}

func (srv *mDNSService) Stop() {
	srv.logger.Info("Stop mDNS Service")
	srv.StopRegisterServer()
	close(srv.done)
}
