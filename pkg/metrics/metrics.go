// Package metrics provides Prometheus metrics for GOSP observability.
// It defines counters, gauges, and histograms for monitoring cluster health and performance.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ActiveWorkerCount tracks the number of healthy nodes in the cluster
	ActiveWorkerCount = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "osp",
		Subsystem: "master",
		Name:      "active_workers",
		Help:      "Number of currently connected and healthy worker nodes",
	})

	// SearchLatency tracks the end-to-end P95 latency of search requests
	SearchLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "osp",
		Subsystem: "master",
		Name:      "search_latency_seconds",
		Help:      "Histogram of end-to-end search request latency",
		Buckets:   prometheus.DefBuckets,
	}, []string{"engine"})

	// ScrapeErrors tracks failures per engine to trigger circuit breaking
	ScrapeErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "osp",
		Subsystem: "worker",
		Name:      "scrape_errors_total",
		Help:      "Total number of search scraping errors",
	}, []string{"engine", "error_code"})

	// TaskQueueDepth tracks the number of pending search tasks
	TaskQueueDepth = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "osp",
		Subsystem: "master",
		Name:      "task_queue_depth",
		Help:      "Number of search tasks currently queued for workers",
	})
)
