package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var defaultConfig = Config{
	PollingInterval: time.Minute,
	SmartctlBinary:  "smartctl",
	UseSudo:         false,
}

type Config struct {
	PollingInterval time.Duration  `yaml:"polling_interval"`
	SmartctlBinary  string         `yaml:"smartctl_binary"`
	UseSudo         bool           `yaml:"use_sudo"`
	MetricPrefix    string         `yaml:"metric_prefix"`
	DeviceTags      []string       `yaml:"device_tags"`
	Devices         []DeviceConfig `yaml:"devices"`
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
	err = cfg.Validate()
	if err != nil {
		log.Fatal().
			Err(err).
			Str("filename", filename).
			Msg("Configuration is invalid")
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

func (c Config) Validate() error {
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
	addErrIf(c.SmartctlBinary == "", "smartctl binary is empty")
	addErrIf(c.MetricPrefix == "", "metric prefix is not specified")
	addErrIf(c.PollingInterval < time.Second,
		"polling interval must be at least one second (got %s)",
		c.PollingInterval.String())

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

	if len(errorList) == 0 {
		return nil
	}
	return errors.New(strings.Join(errorList, ", "))
}
