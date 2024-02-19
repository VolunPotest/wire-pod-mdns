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

const (
	defaultPort = 8084
)

var (
	srvIPFlag        string
	lookupPeriodFlag time.Duration
	debugFlag        bool
)

func main() {
	// parse flags
	flag.StringVar(&srvIPFlag, "srvIP", "", "IP server of wire pod instance")
	flag.DurationVar(&lookupPeriodFlag, "period", time.Minute, "lookup period")
	flag.BoolVar(&debugFlag, "debug", false, "Enable debug logging")
	flag.Parse()

	if len(srvIPFlag) == 0 {
		log.Fatal("Please provide srvIP arg")
	}

	// Create logger
	logLevel := &slog.LevelVar{}
	if debugFlag {
		logLevel.Set(slog.LevelDebug)
	}
	opts := &slog.HandlerOptions{Level: logLevel}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opts))

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	mDNSService, err := mdns.NewmDNSService(logger, srvIPFlag, defaultPort, lookupPeriodFlag)
	if err != nil {
		log.Fatal(err)
	}
	defer mDNSService.Stop()
	mDNSService.StartLookup()

	for range sig {
		logger.Info("Received exit signal")
		break
	}

	logger.Info("Shutting down")
}
