// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package pkg

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/sirupsen/logrus"
)

const (
	REQUEST_CACHE_KEY_HEADER = "X-Gitpod-Cache-Key"
	REQUEST_ID_HEADER        = "X-Gitpod-Request-ID"
	LOG_FIELD_REQUEST_ID     = "request_id"
	LOG_FIELD_REQUEST        = "request"
	LOG_FIELD_FUNC           = "func"
	LOG_FIELD_STATUS         = "status"
)

var (
	cfg         *Config
	upstreamURL *url.URL
)

func Run() {
	log.Init("openvsx-proxy", "", true, true)

	if len(os.Args) != 2 {
		log.Panicf("Usage: %s </path/to/config.json>", os.Args[0])
	}

	var err error
	cfg, err = ReadConfig(os.Args[1])
	if err != nil {
		log.WithError(err).Panic("error reading config")
	}

	log.WithField("config", string(cfg.ToJson())).Info("starting OpenVSX proxy ...")

	StartPromotheus(cfg)

	err = SetupCache()
	if err != nil {
		log.WithError(err).Panic("error setting up cache")
	}

	upstreamURL, err = url.Parse(cfg.URLUpstream)
	if err != nil {
		log.WithError(err).Panic("error parsing upstream URL")
	}

	http.DefaultTransport.(*http.Transport).MaxIdleConns = cfg.MaxIdleConns
	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost

	proxy := httputil.NewSingleHostReverseProxy(upstreamURL)
	proxy.ErrorHandler = ErrorHandler
	proxy.ModifyResponse = ModifyResponse
	proxy.Transport = &DurationTrackingTransport{}

	http.HandleFunc("/", Handler(proxy))
	log.Info("listening on port 8080 ...")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Panic(err)
	}
}

type DurationTrackingTransport struct {
}

func (t *DurationTrackingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	reqid := r.Header.Get(REQUEST_ID_HEADER)
	key := r.Header.Get(REQUEST_CACHE_KEY_HEADER)
	if len(key) == 0 {
		key = fmt.Sprintf("%s %s", r.Method, r.URL)
	}

	logFields := logrus.Fields{
		LOG_FIELD_FUNC:       "transport_roundtrip",
		LOG_FIELD_REQUEST_ID: reqid,
		LOG_FIELD_REQUEST:    key,
	}

	start := time.Now()
	defer func(ts time.Time) {
		duration := time.Since(ts)
		DurationUpstreamCallHistorgram.Observe(duration.Seconds())
		log.
			WithFields(logFields).
			WithFields(DurationLogFields(duration)).
			Info("upstream call finished")
	}(start)
	return http.DefaultTransport.RoundTrip(r)
}
