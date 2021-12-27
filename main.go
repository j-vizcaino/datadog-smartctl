package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/j-vizcaino/datadog-smartctl/poller"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	cfg := MustLoadValidConfig("datadog-smartctl.yaml")

	queryFunc := getDeviceQuerier(cfg)
	appCtx, abort := context.WithCancel(context.Background())
	// queryFunc := testDeviceQuery()
	var pollers []*poller.Poller
	for _, dev := range cfg.Devices {
		p := poller.New(queryFunc, getDataTranslator(cfg, dev), dev.Path)
		log.Info().
			Str("device", dev.Path).
			Dur("interval", cfg.PollingInterval).
			Msg("Starting SMART data periodic poller")
		p.Poll(appCtx, cfg.PollingInterval)
		pollers = append(pollers, p)
	}

	waitForSignal()
	abort()

	for _, p := range pollers {
		p.Stop()
	}
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	log.Info().Str("signal", sig.String()).Msg("Caught signal, exiting")
}
