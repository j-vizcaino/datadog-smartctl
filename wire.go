package main

import (
	"context"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/rs/zerolog/log"

	"github.com/j-vizcaino/datadog-smartctl/converter"
	"github.com/j-vizcaino/datadog-smartctl/poller"
	"github.com/j-vizcaino/datadog-smartctl/smartctl"
	"github.com/j-vizcaino/datadog-smartctl/submitter"
)

func getDataTranslator(cfg Config, devConfig DeviceConfig, submit *submitter.Submitter) poller.OnNewDataFunc {
	conv := converter.New(
		cfg.Statsd.MetricsPrefix,
		converter.WithTags(cfg.Statsd.DeviceTags...),
		converter.WithATASmartAttributes(devConfig.ATASmartAttributesMetrics...),
		converter.WithATADeviceStats(devConfig.ATADeviceStatsMetrics...),
		converter.WithNVMeHealthInfo(devConfig.NVMeHealthInfoMetrics...),
	)

	return func(ctx context.Context, data smartctl.Data) {
		metrics := conv.Convert(data)
		submit.Update(ctx, metrics)
	}
}

func getDeviceQuerier(cfg SmartCtlConfig) poller.QueryDeviceFunc {
	var opts []smartctl.CommandOption

	if cfg.UseSudo {
		opts = append(opts, smartctl.WithSudoEnabled())
	}
	if cfg.Binary != "" {
		opts = append(opts, smartctl.WithSmartctlBinary(cfg.Binary))
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

func submitErrorLog(err error) {
	log.Warn().Err(err).Msg("Submitter error")
}

func getSubmitter(cfg StatsdConfig) (*submitter.Submitter, func()) {
	statsdClient, err := statsd.NewBuffered(cfg.URL, 32)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize statsd client")
	}
	s := submitter.New(statsdClient, submitErrorLog)
	stop := func() {
		s.Stop()
		if err := statsdClient.Flush(); err != nil {
			log.Warn().Err(err).Msg("statsd client flush failed")
		}
		if err := statsdClient.Close(); err != nil {
			log.Warn().Err(err).Msg("statsd client close operation failed")
		}
	}
	return s, stop
}

/*
/!\ For local testing purposes where smartctl is not available

func testDeviceQuery() poller.QueryDeviceFunc {
	out, _ := ioutil.ReadFile("smartctl/testdata/smartctl-output-wd-red.json")
	obj := objx.MustFromJSON(string(out))
	d, err := smartctl.NewData(obj)
	return func(ctx context.Context, dev string) (smartctl.Data, error) {
		return d, err
	}
}
*/
