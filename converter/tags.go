package converter

import (
	"reflect"
	"sort"

	"github.com/scylladb/go-set/strset"
)

type deviceWithTags struct {
	Name            string `name:"device_name"`
	Type            string `name:"device_type"`
	Protocol        string `name:"device_protocol"`
	ModelFamily     string `name:"model_family"`
	ModelName       string `name:"model_name"`
	SerialNumber    string `name:"serial_number"`
	FirmwareVersion string `name:"firmware_version"`
}

var supportedTags *strset.Set

func populateSupportedTags() {
	t := reflect.TypeOf(deviceWithTags{})
	supportedTags = strset.NewWithSize(t.NumField())
	for idx := 0; idx < t.NumField(); idx++ {
		field := t.Field(idx)
		supportedTags.Add(field.Tag.Get("name"))
	}
}

func UnknownTags(tags []string) []string {
	var unknown []string
	for _, tag := range tags {
		if !supportedTags.Has(tag) {
			unknown = append(unknown, tag)
		}
	}
	sort.Strings(unknown)
	return unknown
}
