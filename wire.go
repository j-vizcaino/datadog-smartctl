package main

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/j-vizcaino/datadog-smartctl/converter"
	"github.com/j-vizcaino/datadog-smartctl/poller"
	"github.com/j-vizcaino/datadog-smartctl/smartctl"
)

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
