package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/j-vizcaino/datadog-smartctl/poller"
)

func main() {
	setupPrettyLogger()
	cfgFilename := "datadog-smartctl.yaml"
	if len(os.Args) >=2 {
		cfgFilename = os.Args[1]
	}

	cfg := MustLoadValidConfig(cfgFilename)

	queryFunc := getDeviceQuerier(cfg.Smartctl)
	submitter, submitterStop := getSubmitter(cfg.Statsd)
	submitter.Run(5 * time.Second)

	appCtx, abort := context.WithCancel(context.Background())
	// queryFunc := testDeviceQuery()
	var pollers []*poller.Poller
	for _, dev := range cfg.Devices {
		p := poller.New(queryFunc, getDataTranslator(cfg, dev, submitter), dev.Path)
		log.Info().
			Str("device", dev.Path).
			Dur("interval", cfg.Smartctl.PollingInterval).
			Msg("Starting SMART data periodic poller")
		p.Poll(appCtx, cfg.Smartctl.PollingInterval)
		pollers = append(pollers, p)
	}

	waitForSignal()
	abort()

	for _, p := range pollers {
		p.Stop()
	}
	submitterStop()
}

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	sig := <-sigChan
	log.Info().Str("signal", sig.String()).Msg("Caught signal, exiting")
}

func setupPrettyLogger() {
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		return
	}
	w := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: "15:04:05",
	}
	log.Logger = log.Output(w).With().Timestamp().Logger()
}
