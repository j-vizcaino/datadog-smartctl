package converter

import (
	"reflect"
	"strings"

	"github.com/scylladb/go-set/strset"

	"github.com/j-vizcaino/datadog-smartctl/metric"
	"github.com/j-vizcaino/datadog-smartctl/smartctl"
)

func init() {
	populateSupportedTags()
}

type Converter struct {
	metricPrefix string
	commonTags   *strset.Set
	extractors   []metricsExtractor
}

type Option func(converter *Converter)

func WithTags(tagNames ...string) Option {
	return func(c *Converter) {
		c.commonTags = strset.New(tagNames...)
	}
}

func WithATASmartAttributes(entries ...string) Option {
	const prefix = "ata_smart_attributes."
	return func(c *Converter) {
		c.extractors = append(c.extractors,
			&extractorATASmartAttr{
				metricPrefix: c.metricPrefix + prefix,
				entries:      entries,
			})
	}
}

func WithATADeviceStats(entries ...string) Option {
	const prefix = "ata_device_stats."
	return func(c *Converter) {
		c.extractors = append(c.extractors,
			&extractorATADeviceStats{
				metricPrefix: c.metricPrefix + prefix,
				entries:      entries,
			})
	}
}

func WithNVMeHealthInfo(entries ...string) Option {
	const prefix = "nvme_health."
	return func(c *Converter) {
		c.extractors = append(c.extractors,
			&extractorNVMeHealthInfo{
				metricPrefix: c.metricPrefix + prefix,
				entries:      entries,
			})
	}
}

func New(metricPrefix string, opts ...Option) *Converter {
	c := &Converter{
		metricPrefix: strings.Trim(metricPrefix, ".") + ".",
	}
	for _, setOption := range opts {
		setOption(c)
	}
	return c
}

func (c *Converter) Convert(data smartctl.Data) metric.DeviceMetrics {
	entries := make([]metric.Metric, 0, 64)
	for _, extractor := range c.extractors {
		entries = append(
			entries,
			extractor.Extract(data)...,
		)
	}
	return metric.DeviceMetrics{
		DeviceName: data.Device.Name,
		CommonTags: c.extractTags(data),
		Entries:    entries,
	}
}

func (c *Converter) extractTags(data smartctl.Data) []string {
	tags := make([]string, 0, c.commonTags.Size())

	taggedDevice := deviceWithTags(data.Device)
	t := reflect.TypeOf(taggedDevice)
	v := reflect.ValueOf(taggedDevice)
	for idx := 0; idx < t.NumField(); idx++ {
		field := t.Field(idx)
		tagName := field.Tag.Get("name")
		if !c.commonTags.Has(tagName) {
			continue
		}
		fieldValue := v.Field(idx).String()
		if fieldValue != "" {
			tags = append(tags, tagName+":"+fieldValue)
		}
	}
	return tags
}

type metricsExtractor interface {
	Extract(smartctl.Data) []metric.Metric
}

type extractorATASmartAttr struct {
	metricPrefix string
	entries      []string
}

func (e extractorATASmartAttr) Extract(data smartctl.Data) []metric.Metric {
	return extract(data.ATASmartAttributes, e.metricPrefix, e.entries)
}

type extractorATADeviceStats struct {
	metricPrefix string
	entries      []string
}

func (e extractorATADeviceStats) Extract(data smartctl.Data) []metric.Metric {
	return extract(data.ATADeviceStats, e.metricPrefix, e.entries)
}

type extractorNVMeHealthInfo struct {
	metricPrefix string
	entries      []string
}

func (e extractorNVMeHealthInfo) Extract(data smartctl.Data) []metric.Metric {
	return extract(data.NVMeSmartHealthInfo, e.metricPrefix, e.entries)
}

func extract(data map[string]int, metricPrefix string, entries []string) []metric.Metric {
	out := make([]metric.Metric, 0, len(entries))
	for _, entry := range entries {
		value, ok := data[entry]
		if !ok {
			continue
		}
		out = append(out, metric.Metric{
			Name:  metricPrefix + entry,
			Value: value,
		})
	}
	return out
}
