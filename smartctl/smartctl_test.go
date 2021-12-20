package smartctl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/j-vizcaino/datadog-smartctl/exec"
)

func TestQueryDevice(t *testing.T) {
	t.Run("should work with WD disk JSON data", func(t *testing.T) {
		smartCtlCommand = func(d string) *exec.Command {
			require.Equal(t, "/dev/sdc", d)
			return exec.NewCommand("cat", "testdata/smartctl-output-wd-red.json")
		}

		data, err := QueryDevice("/dev/sdc")
		require.NoError(t, err)
		require.Equal(t, DeviceInfo{
			ModelFamily:     "Western Digital Red Pro",
			ModelName:       "WDC WD4003FFBX-68MU3N0",
			SerialNumber:    "VBGHW31F",
			FirmwareVersion: "83.00A83",
		}, data.DeviceInfo)

		fmt.Println(data.DeviceStats)
		fmt.Println(data.SMARTAttributes)
	})

	t.Run("should work with SSD JSON data", func(t *testing.T) {
		smartCtlCommand = func(d string) *exec.Command {
			require.Equal(t, "/dev/sdb", d)
			return exec.NewCommand("cat", "testdata/smartctl-output-ct240bx.json")
		}

		data, err := QueryDevice("/dev/sdb")
		require.NoError(t, err)
		require.Equal(t, DeviceInfo{
			ModelFamily:     "Silicon Motion based SSDs",
			ModelName:       "CT240BX200SSD1",
			SerialNumber:    "1603F015E628",
			FirmwareVersion: "MU02.6",
		}, data.DeviceInfo)

		fmt.Println(data.DeviceStats)
		fmt.Println(data.SMARTAttributes)
	})
}
