package utils

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricType defines the type of metric
type MetricType string

const (
	// CounterMetric is a cumulative metric that only increases
	CounterMetric MetricType = "counter"
	// GaugeMetric is a metric that can increase and decrease
	GaugeMetric MetricType = "gauge"
	// HistogramMetric is a metric that samples observations and counts them in buckets
	HistogramMetric MetricType = "histogram"
)

// Metric represents a single metric
type Metric struct {
	Name        string
	Type        MetricType
	Value       float64
	Labels      map[string]string
	LastUpdated time.Time
	mutex       sync.RWMutex
}

// MetricsCollector collects and manages metrics
type MetricsCollector struct {
	metrics map[string]*Metric
	mutex   sync.RWMutex
	logger  *zap.Logger
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		metrics: make(map[string]*Metric),
		logger:  logger,
	}
}

// Counter creates or gets a counter metric
func (mc *MetricsCollector) Counter(name string, labels map[string]string) *Metric {
	return mc.getOrCreateMetric(name, CounterMetric, labels)
}

// Gauge creates or gets a gauge metric
func (mc *MetricsCollector) Gauge(name string, labels map[string]string) *Metric {
	return mc.getOrCreateMetric(name, GaugeMetric, labels)
}

// Histogram creates or gets a histogram metric
func (mc *MetricsCollector) Histogram(name string, labels map[string]string) *Metric {
	return mc.getOrCreateMetric(name, HistogramMetric, labels)
}

// getOrCreateMetric gets an existing metric or creates a new one
func (mc *MetricsCollector) getOrCreateMetric(name string, metricType MetricType, labels map[string]string) *Metric {
	key := mc.getMetricKey(name, labels)

	mc.mutex.RLock()
	metric, exists := mc.metrics[key]
	mc.mutex.RUnlock()

	if exists {
		return metric
	}

	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	// Check again in case another goroutine created it
	metric, exists = mc.metrics[key]
	if exists {
		return metric
	}

	// Create new metric
	metric = &Metric{
		Name:        name,
		Type:        metricType,
		Value:       0,
		Labels:      labels,
		LastUpdated: time.Now(),
	}

	mc.metrics[key] = metric
	return metric
}

// getMetricKey generates a unique key for a metric based on name and labels
func (mc *MetricsCollector) getMetricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += ":" + k + "=" + v
	}
	return key
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	result := make(map[string]interface{})
	for key, metric := range mc.metrics {
		metric.mutex.RLock()
		result[key] = map[string]interface{}{
			"name":         metric.Name,
			"type":         string(metric.Type),
			"value":        metric.Value,
			"labels":       metric.Labels,
			"last_updated": metric.LastUpdated,
		}
		metric.mutex.RUnlock()
	}

	return result
}

// Inc increments a counter metric by 1
func (m *Metric) Inc() {
	m.Add(1)
}

// Add adds the given value to the metric
func (m *Metric) Add(value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Value += value
	m.LastUpdated = time.Now()
}

// Set sets the metric to the given value (for gauges)
func (m *Metric) Set(value float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.Value = value
	m.LastUpdated = time.Now()
}

// Get returns the current value of the metric
func (m *Metric) Get() float64 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.Value
}

// Observe records a value for histogram metrics
func (m *Metric) Observe(value float64) {
	// For a simple implementation, we just add the value
	// In a real implementation, this would update histogram buckets
	m.Add(value)
}

// Timer measures the execution time of a function
func (m *Metric) Timer() func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		m.Observe(float64(duration.Milliseconds()))
	}
}

// RequestMetrics tracks metrics for HTTP requests
type RequestMetrics struct {
	collector *MetricsCollector
}

// NewRequestMetrics creates a new request metrics tracker
func NewRequestMetrics(collector *MetricsCollector) *RequestMetrics {
	return &RequestMetrics{
		collector: collector,
	}
}

// TrackRequest tracks metrics for an HTTP request
func (rm *RequestMetrics) TrackRequest(path, method string, statusCode int, startTime time.Time) {
	duration := time.Since(startTime)
	
	// Track request count
	rm.collector.Counter("http_requests_total", map[string]string{
		"path":   path,
		"method": method,
		"status": string(rune(statusCode)),
	}).Inc()
	
	// Track request duration
	rm.collector.Histogram("http_request_duration_ms", map[string]string{
		"path":   path,
		"method": method,
	}).Observe(float64(duration.Milliseconds()))
}

// GlobalMetricsCollector is a singleton metrics collector
var (
	globalMetricsCollector     *MetricsCollector
	globalMetricsCollectorOnce sync.Once
)

// GetGlobalMetricsCollector returns the global metrics collector
func GetGlobalMetricsCollector(logger *zap.Logger) *MetricsCollector {
	globalMetricsCollectorOnce.Do(func() {
		globalMetricsCollector = NewMetricsCollector(logger)
	})
	return globalMetricsCollector
}