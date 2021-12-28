package converter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/j-vizcaino/datadog-smartctl/metric"
	"github.com/j-vizcaino/datadog-smartctl/smartctl"
)

func TestConverter(t *testing.T) {
	t.Run("should work with ATA device content", func(t *testing.T) {
		data := smartctl.Data{
			Device: smartctl.DeviceInfo{
				Name:            "/dev/sdc",
				Type:            "sat",
				Protocol:        "ATA",
				ModelFamily:     "Western Digital Red Pro",
				ModelName:       "WDC WD4003FFBX-68MU3N0",
				SerialNumber:    "VBGHW31F",
				FirmwareVersion: "83.00A83",
			},
			ATADeviceStats: map[string]int{
				"average long term temperature":                35,
				"average short term temperature":               36,
				"current temperature":                          37,
				"date and time timestamp":                      27353660700,
				"head flying hours":                            7593,
				"head load events":                             333,
				"highest average long term temperature":        40,
				"highest average short term temperature":       43,
				"highest temperature":                          45,
				"lifetime power-on resets":                     17,
				"logical sectors read":                         76849055332,
				"logical sectors written":                      14999499540,
				"lowest average long term temperature":         25,
				"lowest average short term temperature":        25,
				"lowest temperature":                           16,
				"number of asr events":                         9,
				"number of hardware resets":                    76,
				"number of interface crc errors":               0,
				"number of mechanical start failures":          0,
				"number of read commands":                      114765179,
				"number of reallocated logical sectors":        0,
				"number of reported uncorrectable errors":      0,
				"number of write commands":                     27609411,
				"power-on hours":                               7598,
				"read recovery attempts":                       0,
				"resets between cmd acceptance and completion": 0,
				"specified maximum operating temperature":      60,
				"specified minimum operating temperature":      0,
				"spindle motor power-on hours":                 7593,
				"time in over-temperature":                     0,
				"time in under-temperature":                    0,
			},
			ATASmartAttributes: map[string]int{
				"current_pending_sector":  0,
				"load_cycle_count":        333,
				"offline_uncorrectable":   0,
				"power-off_retract_count": 333,
				"power_cycle_count":       17,
				"power_on_hours":          7598,
				"raw_read_error_rate":     0,
				"reallocated_event_count": 0,
				"reallocated_sector_ct":   0,
				"seek_error_rate":         0,
				"seek_time_performance":   18,
				"spin_retry_count":        0,
				"spin_up_time":            34374156674,
				"start_stop_count":        17,
				"temperature_celsius":     37,
				"throughput_performance":  96,
				"udma_crc_error_count":    0,
			},
		}
		converter := New(
			"foo.bar",
			WithTags("device_name", "model_name", "device_protocol"),
			WithATADeviceStats("logical sectors read"),
			WithATASmartAttributes("temperature_celsius", "raw_read_error_rate"),
		)
		metrics := converter.Convert(data)
		require.Equal(t, data.Device.Name, metrics.DeviceName)
		require.ElementsMatch(t, []string{
			"model_name:" + data.Device.ModelName,
			"device_name:" + data.Device.Name,
			"device_protocol:" + data.Device.Protocol,
		}, metrics.CommonTags)
		require.ElementsMatch(t, []metric.Metric{
			{"foo.bar.ata_smart_attributes.temperature_celsius", 37},
			{"foo.bar.ata_smart_attributes.raw_read_error_rate", 0},
			{"foo.bar.ata_device_stats.logical sectors read", 76849055332},
		}, metrics.Entries)
	})

	t.Run("should work with NVMe device content", func(t *testing.T) {
		data := smartctl.Data{
			Device: smartctl.DeviceInfo{
				Name:            "/dev/nvme0n1",
				Type:            "nvme",
				Protocol:        "NVMe",
				ModelName:       "WDC WDS500G2B0C-00PXH0",
				SerialNumber:    "2044DZ473606",
				FirmwareVersion: "211070WD",
			},
			NVMeSmartHealthInfo: map[string]int{
				"critical_warning":          0,
				"temperature":               35,
				"available_spare":           100,
				"available_spare_threshold": 10,
				"percentage_used":           0,
				"data_units_read":           21604166,
				"data_units_written":        2433328,
				"host_reads":                189205682,
				"host_writes":               137800871,
				"controller_busy_time":      1524,
				"power_cycles":              13,
				"power_on_hours":            7146,
				"unsafe_shutdowns":          4,
				"media_errors":              0,
				"num_err_log_entries":       1,
				"warning_temp_time":         0,
				"critical_comp_time":        0,
			},
		}

		converter := New(
			"test",
			WithTags("model_name",
				"model_family",
				"serial_number",
				"firmware_version",
				"device_name",
				"device_type",
				"device_protocol",
			),
			WithNVMeHealthInfo("temperature", "available_spare"),
		)
		metrics := converter.Convert(data)
		require.Equal(t, data.Device.Name, metrics.DeviceName)
		require.ElementsMatch(t, []string{
			"model_name:" + data.Device.ModelName,
			"serial_number:" + data.Device.SerialNumber,
			"firmware_version:" + data.Device.FirmwareVersion,
			"device_name:" + data.Device.Name,
			"device_type:" + data.Device.Type,
			"device_protocol:" + data.Device.Protocol,
		}, metrics.CommonTags)
		require.ElementsMatch(t, []metric.Metric{
			{"test.nvme_health.temperature", 35},
			{"test.nvme_health.available_spare", 100},
		}, metrics.Entries)

	})
}
