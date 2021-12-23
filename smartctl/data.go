package smartctl

import (
	"errors"
	"fmt"
	"strings"

	"github.com/stretchr/objx"
)

type DeviceInfo struct {
	Name            string // /dev/xxx
	Type            string // nvme, sat
	Protocol        string // NVMe, ATA
	ModelFamily     string
	ModelName       string
	SerialNumber    string
	FirmwareVersion string
}

type Data struct {
	Device              DeviceInfo
	NVMeSmartHealthInfo map[string]int
	ATASmartAttributes  map[string]int
	ATADeviceStats      map[string]int
}

func NewData(raw objx.Map) (Data, error) {
	res := Data{Device: extractDeviceInfo(raw)}

	switch res.Device.Protocol {
	case "NVMe":
		res.NVMeSmartHealthInfo = extractNVMeHealthInformation(raw)
	case "ATA":
		res.ATASmartAttributes = extractATASmartAttributes(raw)
		res.ATADeviceStats = extractATADeviceStats(raw)
	case "":
		return Data{}, errors.New("undetected device protocol, empty or missing device.protocol JSON field")
	default:
		return Data{}, fmt.Errorf("unsupported device protocol %s (expected ATA or NMVe)", res.Device.Protocol)
	}

	return res, nil
}

func extractDeviceInfo(m objx.Map) DeviceInfo {
	return DeviceInfo{
		Name:            m.Get("device.name").String(),
		Type:            m.Get("device.type").String(),
		Protocol:        m.Get("device.protocol").String(),
		ModelFamily:     m.Get("model_family").String(),
		ModelName:       m.Get("model_name").String(),
		SerialNumber:    m.Get("serial_number").String(),
		FirmwareVersion: m.Get("firmware_version").String(),
	}
}

func extractATASmartAttributes(m objx.Map) map[string]int {
	out := make(map[string]int)
	m.Get("ata_smart_attributes.table").EachObjxMap(func(_ int, obj objx.Map) bool {
		name := obj.Get("name").String()
		name = strings.ToLower(name)

		value := obj.Get("raw.value").Int()

		// NOTE: temperature raw value is borked, we need to parse the string representation.
		// Example string: "26 (35 33 36 35 0)", with 26 being the current temperature
		if name == "temperature_celsius" {
			var temp int
			valueStr := obj.Get("raw.string").String()
			if _, err := fmt.Sscan(valueStr, &temp); err == nil {
				value = temp
			}
		}

		out[name] = value
		return true
	})
	return out
}

func extractATADeviceStats(m objx.Map) map[string]int {
	out := make(map[string]int)
	m.Get("ata_device_statistics.pages").EachObjxMap(func(_ int, page objx.Map) bool {
		if !page.Has("table") {
			return true
		}
		page.Get("table").EachObjxMap(func(_ int, entry objx.Map) bool {
			name := entry.Get("name").String()
			name = strings.ToLower(name)

			out[name] = entry.Get("value").Int()
			return true
		})
		return true
	})
	return out
}

func extractNVMeHealthInformation(m objx.Map) map[string]int {
	out := make(map[string]int)
	healthInfo := m.Get("nvme_smart_health_information_log").ObjxMap()
	for key := range healthInfo {
		name := strings.ToLower(key)
		out[name] = healthInfo.Get(key).Int()
	}
	return out
}
