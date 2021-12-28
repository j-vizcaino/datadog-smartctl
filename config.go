package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"

	"github.com/j-vizcaino/datadog-smartctl/converter"
)

var defaultConfig = Config{
	Smartctl: SmartCtlConfig{
		PollingInterval: time.Minute,
		Binary:          "smartctl",
		UseSudo:         false,
	},
	Statsd: StatsdConfig{
		URL:            "localhost:8125",
		MetricsPrefix:  "smartctl.",
		DeviceTags:     []string{"device_path"},
		ReportInterval: 10 * time.Second,
	},
}

type Config struct {
	Smartctl SmartCtlConfig `yaml:"smartctl"`
	Statsd   StatsdConfig   `yaml:"statsd"`
	Devices  []DeviceConfig `yaml:"devices"`
}

type StatsdConfig struct {
	URL            string        `yaml:"url"`
	DeviceTags     []string      `yaml:"device_tags"`
	MetricsPrefix  string        `yaml:"metrics_prefix"`
	ReportInterval time.Duration `yaml:"report_interval"`
}

type SmartCtlConfig struct {
	PollingInterval time.Duration `yaml:"polling_interval"`
	Binary          string        `yaml:"binary"`
	UseSudo         bool          `yaml:"use_sudo"`
}

type DeviceConfig struct {
	Path                      string   `yaml:"path"`
	ATASmartAttributesMetrics []string `yaml:"ata_smart_attributes_metrics"`
	ATADeviceStatsMetrics     []string `yaml:"ata_device_stats_metrics"`
	NVMeHealthInfoMetrics     []string `yaml:"nvme_health_info_metrics"`
}

func MustLoadValidConfig(filename string) Config {
	cfg, err := LoadConfig(filename)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}
	cfgErrors := cfg.Errors()
	for _, err := range cfgErrors {
		log.Error().
			Err(errors.New(err)).
			Str("filename", filename).
			Msg("Configuration is invalid")
	}
	if len(cfgErrors) > 0 {
		log.Fatal().Msg("Configuration is invalid, aborting application")
	}
	return cfg
}

func LoadConfig(filename string) (Config, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer fd.Close()

	decoder := yaml.NewDecoder(fd)
	cfg := defaultConfig
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("failed to decode YAML from %q, %w", filename, err)
	}
	return cfg, nil
}

func (c Config) Errors() []string {
	var errorList []string
	addErr := func(format string, args ...interface{}) {
		errorList = append(errorList, fmt.Sprintf(format, args...))
	}
	addErrIf := func(condition bool, format string, args ...interface{}) {
		if condition {
			addErr(format, args...)
		}
	}

	addErrIf(len(c.Devices) == 0, "devices are not specified")
	addErrIf(c.Smartctl.Binary == "", "smartctl binary is empty")
	addErrIf(c.Smartctl.PollingInterval < time.Second,
		"smartctl polling interval must be at least one second (got %s)",
		c.Smartctl.PollingInterval.String())

	addErrIf(c.Statsd.MetricsPrefix == "", "metric prefix is not specified")
	addErrIf(c.Statsd.URL == "", "statsd URL is empty")
	addErrIf(c.Statsd.ReportInterval < time.Second,
		"statsd report interval must be at least one second (got %s)",
		c.Statsd.ReportInterval.String())
	unknownTags := converter.UnknownTags(c.Statsd.DeviceTags)
	addErrIf(len(unknownTags) > 0, "unknown device tags %s", strings.Join(unknownTags, ", "))

	for idx, dev := range c.Devices {
		if dev.Path == "" {
			addErr("devices[%d] must specify a path", idx)
			continue
		}
		ataMetrics := len(dev.ATADeviceStatsMetrics) + len(dev.ATASmartAttributesMetrics)
		nvmeMetrics := len(dev.NVMeHealthInfoMetrics)
		addErrIf(ataMetrics+nvmeMetrics == 0,
			"device %s must specify at least one of ATA or NVMe metrics",
			dev.Path)

		addErrIf(ataMetrics != 0 && nvmeMetrics != 0,
			"device %s cannot specify both ATA and NVMe metrics",
			dev.Path)
	}
	return errorList
}
