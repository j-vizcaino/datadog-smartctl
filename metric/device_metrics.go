package metric

type DeviceMetrics struct {
	DeviceName string
	CommonTags []string
	Entries    []Metric
}

type Metric struct {
	Name  string
	Value int
}
