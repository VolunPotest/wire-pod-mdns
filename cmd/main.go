package main

import (
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
	mdns "wire-pod-mdns"
)

var (
	srvIPFlag    string
	lookupPeriod time.Duration
)

func main() {
	flag.StringVar(&srvIPFlag, "srvIP", "", "IP server of wire pod instance")
	flag.DurationVar(&lookupPeriod, "period", time.Minute, "lookup period")
	flag.Parse()

	if len(srvIPFlag) == 0 {
		log.Fatal("Please provide srvIP arg")
	}

	// Create logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	mDNSService, err := mdns.NewmDNSService(logger, srvIPFlag, 8084, lookupPeriod)
	if err != nil {
		log.Fatal(err)
	}
	defer mDNSService.Stop()
	mDNSService.StartLookup()

readSignals:
	for {
		select {
		case <-sig:
			logger.Info("received exit signal")
			break readSignals
		}
	}
	logger.Info("Shutting down.")
}
