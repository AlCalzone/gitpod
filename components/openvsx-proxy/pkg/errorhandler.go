// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package pkg

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/sirupsen/logrus"
)

func ErrorHandler(rw http.ResponseWriter, r *http.Request, e error) {
	reqid := r.Header.Get(REQUEST_ID_HEADER)

	logFields := logrus.Fields{
		LOG_FIELD_FUNC:       "error_handler",
		LOG_FIELD_REQUEST_ID: reqid,
		LOG_FIELD_REQUEST:    fmt.Sprintf("%s %s", r.Method, r.URL),
		LOG_FIELD_STATUS:     "error",
	}

	key := r.Header.Get(REQUEST_CACHE_KEY_HEADER)
	if len(key) > 0 {
		logFields[LOG_FIELD_REQUEST] = key
	}

	start := time.Now()
	defer func(ts time.Time) {
		duration := time.Since(ts)
		DurationResponseProcessingHistogram.Observe(duration.Seconds())
		log.
			WithFields(logFields).
			WithFields(DurationLogFields(duration)).
			Info("processing error finished")
	}(start)

	log.WithFields(logFields).WithError(e).Warn("handling error")
	IncStatusCounter(r, "error")

	if key == "" {
		log.WithFields(logFields).Error("cache key header is missing")
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	cached, ok, err := ReadCache(key)
	if err != nil {
		log.WithFields(logFields).WithError(err).Error("cannot read from cache")
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	if !ok {
		log.WithFields(logFields).Debug("cache has no entry for key")
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	for k, v := range cached.Header {
		for i, val := range v {
			if i == 0 {
				rw.Header().Set(k, val)
			} else {
				rw.Header().Add(k, val)
			}
		}
	}
	rw.WriteHeader(cached.StatusCode)
	rw.Write(cached.Body)
	log.WithFields(logFields).Info("used cached response due to a proxy error")
	BackupCacheServeCounter.Inc()
}
