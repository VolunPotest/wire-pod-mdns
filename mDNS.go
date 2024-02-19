package mdns

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

const (
	lookupName = "_ankivector._tcp"
	domainName = "local."

	escapepodSrvName = "escapepod"
)

type mDNSService struct {
	port         int
	ip           string
	lookupPeriod time.Duration

	logger    *slog.Logger
	done      chan struct{}
	lastState bool
	register  *zeroconf.Server
}

func NewmDNSService(parentLog *slog.Logger, ip string, port int, lookupPeriod time.Duration) (*mDNSService, error) {
	logger := parentLog.With("component", "mdns")
	return &mDNSService{
		logger:       logger,
		port:         port,
		ip:           ip,
		lookupPeriod: lookupPeriod,
		done:         make(chan struct{})}, nil
}

func (srv *mDNSService) lookup() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), srv.lookupPeriod/2)
	defer cancel()

	// resolver should be created every time instead of using one
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return false, fmt.Errorf("can't create mDNS service: %w", err)
	}

	// entries closed in the resolver
	entries := make(chan *zeroconf.ServiceEntry)
	err = resolver.Browse(ctx, lookupName, domainName, entries)
	if err != nil {
		return false, fmt.Errorf("failed to browse mDSN: %w", err)
	}

	for entry := range entries {
		if strings.Contains(entry.Service, "ankivector") {
			srv.logger.Debug("Vector discovered on network")
			return true, nil
		}
	}

	srv.logger.Debug("Vektror wasn't found")
	return false, nil
}

func (srv *mDNSService) inspect() {
	srv.logger.Debug("Trying to find vektor in a network...")
	found, err := srv.lookup()
	if err != nil {
		srv.logger.Warn("Can't lookup for vektor: %w", err)
		return
	}
	if found {
		if err = srv.StartRegisterServer(); err != nil {
			srv.logger.Error("error with mDNS register", slog.Any("error", err))
			return
		}
	} else {
		srv.StopRegisterServer()
	}

	if srv.lastState != found {
		srv.logger.Info("mDNS service status was changed", "enable", found)
		srv.lastState = found
	}
}

// Perform lookup each srv.lookupPeriod and register mDNS if necessary
func (srv *mDNSService) StartLookup() {
	srv.logger.Info("Starting lookup service", "period", srv.lookupPeriod.String(), "wire-pod-ip", srv.ip)
	ticker := time.NewTicker(srv.lookupPeriod)
	go func() {
		srv.inspect()
		for {
			select {
			case <-ticker.C:
				srv.inspect()
			case <-srv.done:
				srv.logger.Info("Done signal was sent")
				ticker.Stop()
				return
			}
		}
	}()
}

func (srv *mDNSService) StartRegisterServer() error {
	// TODO: After 1-2 m, wire-pod is not responding anymore (no speech recognition) within same proxy
	// Probably connected with vektor connection to wire-pod
	// if rerun Register each 1m - the wire-pod is working as expected (stop and run register again)
	// need to investigate this
	// if srv.register != nil {
	// 	srv.logger.Debug("server is already registered")
	// 	return nil
	// }
	srv.StopRegisterServer()
	server, err := zeroconf.RegisterProxy(escapepodSrvName, "_app-proto._tcp", domainName, srv.port, escapepodSrvName,
		[]string{srv.ip}, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		return fmt.Errorf("can't register wirepod service: %w", err)
	}

	srv.register = server
	srv.logger.Debug("mDNS service is registered")
	return nil
}

func (srv *mDNSService) StopRegisterServer() {
	if srv.register == nil {
		return
	}

	srv.register.Shutdown()
	srv.register = nil
	srv.logger.Debug("mDSN server is unregistered")
}

func (srv *mDNSService) Stop() {
	srv.logger.Info("Stop mDNS Service")
	srv.StopRegisterServer()
	close(srv.done)
}
