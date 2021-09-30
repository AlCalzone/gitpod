// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package pkg

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	reg                                 *prometheus.Registry
	BackupCacheHitCounter               prometheus.Counter
	BackupCacheMissCounter              prometheus.Counter
	BackupCacheServeCounter             prometheus.Counter
	RegularCacheHitServeCounter         prometheus.Counter
	RegularCacheMissCounter             prometheus.Counter
	RequestsCounter                     *prometheus.CounterVec
	DurationOverallHistogram            prometheus.Histogram
	DurationRequestProcessingHistogram  prometheus.Histogram
	DurationUpstreamCallHistorgram      prometheus.Histogram
	DurationResponseProcessingHistogram prometheus.Histogram
)

func StartPromotheus(cfg *Config) {
	reg = prometheus.NewRegistry()

	if cfg.PrometheusAddr != "" {
		reg.MustRegister(
			prometheus.NewGoCollector(),
			prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		)

		handler := http.NewServeMux()
		handler.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

		go func() {
			err := http.ListenAndServe(cfg.PrometheusAddr, handler)
			if err != nil {
				log.WithError(err).Error("Prometheus metrics server failed")
			}
		}()
		log.WithField("addr", cfg.PrometheusAddr).Info("started Prometheus metrics server")
	}

	createMetrics()
	collectors := []prometheus.Collector{
		BackupCacheHitCounter,
		BackupCacheMissCounter,
		BackupCacheServeCounter,
		RegularCacheHitServeCounter,
		RegularCacheMissCounter,
		RequestsCounter,
		DurationOverallHistogram,
		DurationRequestProcessingHistogram,
		DurationUpstreamCallHistorgram,
		DurationResponseProcessingHistogram,
	}
	for _, c := range collectors {
		err := reg.Register(c)
		if err != nil {
			log.WithError(err).Error("register Prometheus metric failed")
		}
	}
}

func createMetrics() {
	namespace := "gitpod"
	subsystem := "openvsx_proxy"
	BackupCacheHitCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "backup_cache_hit_total",
		Help:      "The total amount of requests where we had a cached response that we could use as backup when the upstream server is down.",
	})
	BackupCacheMissCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "backup_cache_miss_total",
		Help:      "The total amount of requests where we haven't had a cached response that we could use as backup when the upstream server is down.",
	})
	BackupCacheServeCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "backup_cache_serve_total",
		Help:      "The total amount of requests where we actually answered with a cached response because the upstream server is down.",
	})
	RegularCacheHitServeCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regular_cache_hit_and_serve_total",
		Help:      "The total amount or requests where we answered with a cached response for performance reasons.",
	})
	RegularCacheMissCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "regular_cache_miss_total",
		Help:      "The total amount or requests we haven't had a young enough cached requests to use it for performance reasons.",
	})
	RequestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "The total amount of requests by response status.",
	}, []string{"status", "path"})
	DurationOverallHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "duration_overall",
		Help:      "The duration in seconds of the HTTP requests.",
	})
	DurationRequestProcessingHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "duration_request_processing",
		Help:      "The duration in seconds of the processing of the HTTP requests before we call the upstream.",
	})
	DurationUpstreamCallHistorgram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "duration_upstream_call",
		Help:      "The duration in seconds of the call of the upstream server.",
	})
	DurationResponseProcessingHistogram = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "duration_response_processing",
		Help:      "The duration in seconds of the processing of the HTTP responses after we have called the upstream.",
	})
}

func IncStatusCounter(r *http.Request, status string) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/vscode/asset/") {
		// remove everything after /vscode/asset/ to decrease the unique numbers of paths
		path = path[:len("/vscode/asset/")]
	}
	RequestsCounter.WithLabelValues(status, fmt.Sprintf("%s %s", r.Method, path)).Inc()
}
