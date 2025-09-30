package model

import (
	"time"
)

// Metric represents a measured value in time.
type Metric struct {
	Value float64
	TS    time.Time
}

// MetricSeries is a group of metrics identified by an ID and a context
// information.
type MetricSeries struct {
	ID      string
	Labels  map[string]string
	Metrics []Metric
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time
	End   time.Time
}

// Range is a duration representing a time range
type Range time.Duration
