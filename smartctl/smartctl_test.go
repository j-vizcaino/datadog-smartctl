package smartctl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func runCat(testfile string) (Data, error) {
	cmd := NewCommand(WithSmartctlBinary("cat"))
	cmd.smartctlArgs = nil

	return cmd.QueryDevice(context.Background(), testfile)
}

func TestCommand_QueryDevice(t *testing.T) {
	t.Run("should work with SATA HDD", func(t *testing.T) {
		data, err := runCat("testdata/smartctl-output-wd-red.json")
		require.NoError(t, err)

		expected := Data{
			Device: DeviceInfo{
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
		require.Equal(t, expected, data)
	})

	t.Run("should work with SATA SSD", func(t *testing.T) {
		data, err := runCat("testdata/smartctl-output-ct240bx.json")
		require.NoError(t, err)

		expected := Data{
			Device: DeviceInfo{
				Name:            "/dev/sdb",
				Type:            "sat",
				Protocol:        "ATA",
				ModelFamily:     "Silicon Motion based SSDs",
				ModelName:       "CT240BX200SSD1",
				SerialNumber:    "1603F015E628",
				FirmwareVersion: "MU02.6",
			},
			ATADeviceStats: map[string]int{
				"lifetime power-on resets":                     262,
				"logical sectors read":                         1606061987,
				"logical sectors written":                      118967617,
				"number of hardware resets":                    2025,
				"number of interface crc errors":               0,
				"number of read commands":                      51774842,
				"number of reported uncorrectable errors":      0,
				"number of write commands":                     594376077,
				"percentage used endurance indicator":          3,
				"power-on hours":                               3949,
				"resets between cmd acceptance and completion": 20,
			},
			ATASmartAttributes: map[string]int{
				"available_reservd_space": 100,
				"average_erase_count":     32,
				"average_slc_erase_ct":    2663,
				"erase_fail_count_total":  0,
				"host_reads_32mib":        90042,
				"host_writes_32mib":       329495,
				"initial_bad_block_count": 334,
				"max_erase_count":         74,
				"max_slc_erase_ct":        2672,
				"min_erase_count":         8,
				"min_slc_erase_ct":        2630,
				"power-off_retract_count": 20,
				"power_cycle_count":       262,
				"power_on_hours":          3949,
				"program_fail_cnt_total":  0,
				"raid_recoverty_ct":       0,
				"raw_read_error_rate":     0,
				"reallocated_sector_ct":   0,
				"remaining_lifetime_perc": 97,
				"slc_writes_32mib":        394222,
				"tlc_writes_32mib":        247105,
				"temperature_celsius":     26,
				"total_erase_count":       41838,
				"total_slc_erase_ct":      197111,
				"udma_crc_error_count":    0,
				"uncorrectable_error_cnt": 0,
				"valid_spare_block_cnt":   23,
			},
		}
		require.Equal(t, expected, data)
	})

	t.Run("should work with NVMe data", func(t *testing.T) {
		data, err := runCat("testdata/smartctl-output-nvme.json")
		require.NoError(t, err)

		expected := Data{
			Device: DeviceInfo{
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
		require.Equal(t, expected, data)
	})

	t.Run("should surface smartctl errors", func(t *testing.T) {
		cmd := NewCommand(
			WithTimeout(100*time.Millisecond),
			WithSmartctlBinary("cat"),
		)
		cmd.smartctlArgs = []string{"testdata/smartctl-output-error-perm.json", ";", "exit"}

		// cat ...; exit 2
		data, err := cmd.QueryDevice(context.Background(), "2")
		require.Error(t, err)
		require.Contains(t, err.Error(), "Smartctl open device: /dev/sdb [SAT] failed: Permission denied")

		require.Equal(t, Data{}, data)
	})

	t.Run("should implement command timeout", func(t *testing.T) {
		cmd := NewCommand(
			WithTimeout(100*time.Millisecond),
			WithSmartctlBinary("sleep"),
		)
		cmd.smartctlArgs = nil

		// sleep 5
		startDate := time.Now()
		data, err := cmd.QueryDevice(context.Background(), "5")
		elapsed := time.Since(startDate)
		require.Error(t, err)
		require.Contains(t, err.Error(), "signal: killed")
		require.Less(t, elapsed, 4*time.Second)
		require.Equal(t, Data{}, data)
	})
}
