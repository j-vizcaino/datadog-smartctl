package smartctl

import (
	"github.com/stretchr/objx"

	"github.com/j-vizcaino/datadog-smartctl/exec"
)

type DeviceInfo struct {
	ModelFamily     string
	ModelName       string
	SerialNumber    string
	FirmwareVersion string
}

type Data struct {
	DeviceInfo      DeviceInfo
	SMARTAttributes map[string]int
	DeviceStats     map[string]int
}

var smartCtlCommand = func(device string) *exec.Command {
	return exec.NewCommand("smartctl", "--json", "-x", device)
}

func extractDeviceInfo(m objx.Map) DeviceInfo {
	return DeviceInfo{
		ModelFamily:     m.Get("model_family").Str(),
		ModelName:       m.Get("model_name").Str(),
		SerialNumber:    m.Get("serial_number").Str(),
		FirmwareVersion: m.Get("firmware_version").Str(),
	}
}

func extractSmartAttributes(m objx.Map) map[string]int {
	out := make(map[string]int)
	m.Get("ata_smart_attributes.table").EachObjxMap(func(_ int, obj objx.Map) bool {
		name := obj.Get("name").String()
		out[name] = obj.Get("raw.value").Int()
		return true
	})
	return out
}

func extractDeviceStats(m objx.Map) map[string]int {
	out := make(map[string]int)
	m.Get("ata_device_statistics.pages").EachObjxMap(func(_ int, page objx.Map) bool {
		if !page.Has("table") {
			return true
		}
		page.Get("table").EachObjxMap(func(_ int, entry objx.Map) bool {
			name := entry.Get("name").String()
			out[name] = entry.Get("value").Int()
			return true
		})
		return true
	})
	return out
}

func QueryDevice(device string) (Data, error) {
	cmd := smartCtlCommand(device)
	output := cmd.Run()

	raw, err := objx.FromJSON(output)
	if err != nil {
		return Data{}, err
	}

	return Data{
		DeviceInfo:      extractDeviceInfo(raw),
		SMARTAttributes: extractSmartAttributes(raw),
		DeviceStats:     extractDeviceStats(raw),
	}, nil
}
