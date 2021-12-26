package main

import (
	"time"
)

type Config struct {
	PollingInterval time.Duration
	SmartctlBinary  string
	UseSudo         bool
	MetricPrefix    string
	DeviceTags      []string
	Devices         []DeviceConfig
}

type DeviceConfig struct {
	Path                      string
	ATASmartAttributesMetrics []string
	ATADeviceStatsMetrics     []string
	NVMeHealthInfoMetrics     []string
}
