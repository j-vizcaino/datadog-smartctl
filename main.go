package main

import (
	"context"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/objx"

	"github.com/j-vizcaino/datadog-smartctl/converter"
	"github.com/j-vizcaino/datadog-smartctl/poller"
	"github.com/j-vizcaino/datadog-smartctl/smartctl"
)

func getDeviceQuerier(cfg Config) poller.QueryDeviceFunc {
	var opts []smartctl.CommandOption

	if cfg.UseSudo {
		opts = append(opts, smartctl.WithSudoEnabled())
	}
	if cfg.SmartctlBinary != "" {
		opts = append(opts, smartctl.WithSmartctlBinary(cfg.SmartctlBinary))
	}
	smartCmd := smartctl.NewCommand(opts...)

	return func(ctx context.Context, device string) (smartctl.Data, error) {
		logger := log.With().Str("device", device).Logger()
		logger.Info().Msg("Querying SMART information")
		data, err := smartCmd.QueryDevice(ctx, device)
		if err != nil {
			logger.Warn().Err(err).Msg("Querying SMART information failed")
		}
		return data, err
	}
}

func testDeviceQuery() poller.QueryDeviceFunc {
	out, _ := ioutil.ReadFile("smartctl/testdata/smartctl-output-wd-red.json")
	obj := objx.MustFromJSON(string(out))
	d, err := smartctl.NewData(obj)
	return func(ctx context.Context, dev string) (smartctl.Data, error) {
		return d, err
	}
}

func getDataTranslator(cfg Config, devConfig DeviceConfig) poller.OnNewDataFunc {
	conv := converter.New(
		cfg.MetricPrefix,
		converter.WithTags(cfg.DeviceTags...),
		converter.WithATASmartAttributes(devConfig.ATASmartAttributesMetrics...),
		converter.WithATADeviceStats(devConfig.ATADeviceStatsMetrics...),
		converter.WithNVMeHealthInfo(devConfig.NVMeHealthInfoMetrics...),
	)

	return func(ctx context.Context, data smartctl.Data) {
		metrics := conv.Convert(data)
		log.Info().Interface("metrics", metrics).Send()
	}
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

	cfg := Config{
		PollingInterval: 5 * time.Second,
		MetricPrefix:    "smartctl.",
		DeviceTags: []string{
			"device_name",
			"model_name",
		},
		Devices: []DeviceConfig{
			{
				Path: "/dev/sdc",
				ATASmartAttributesMetrics: []string{
					"raw_read_error_rate",
					"seek_error_rate",
					"udma_crc_error_count",
					"temperature_celsius",
				},
				ATADeviceStatsMetrics: []string{
					"number of reported uncorrectable errors",
					"read recovery attempts",
					"number of reallocated logical sectors",
					"number of interface crc errors",
				},
			},
		},
	}

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
