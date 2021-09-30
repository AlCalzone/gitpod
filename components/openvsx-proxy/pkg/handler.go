// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package pkg

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func Handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var (
			hitCacheRegular = false
			hitCacheBackup  = false
		)

		reqid := ""
		uuid, err := uuid.NewRandom()
		if err != nil {
			log.WithError(err).Warn("cannot generate a UUID")
			reqid = fmt.Sprintf("req%d", rand.Intn(999999))
		} else {
			reqid = uuid.String()
		}

		logFields := logrus.Fields{
			LOG_FIELD_FUNC:           "request_handler",
			LOG_FIELD_REQUEST_ID:     reqid,
			LOG_FIELD_REQUEST:        fmt.Sprintf("%s %s", r.Method, r.URL),
			"request_content_length": strconv.FormatInt(r.ContentLength, 10),
		}

		log.WithFields(logFields).Info("handling request")
		r.Header.Set(REQUEST_ID_HEADER, reqid)

		key, err := key(r)
		if err != nil {
			log.WithFields(logFields).WithError(err).Error("cannot create cache key")
			r.Host = upstreamURL.Host
			p.ServeHTTP(rw, r)
			finishLog(logFields, start, hitCacheRegular, hitCacheBackup)
			DurationRequestProcessingHistogram.Observe(time.Since(start).Seconds())
			return
		}
		r.Header.Set(REQUEST_CACHE_KEY_HEADER, key)
		logFields[LOG_FIELD_REQUEST] = key

		if cfg.CacheDurationRegular > 0 {
			cached, ok, err := ReadCache(key)
			if err != nil {
				log.WithFields(logFields).WithError(err).Error("cannot read from cache")
			} else if !ok {
				log.WithFields(logFields).Debug("cache has no entry for key")
			} else {
				hitCacheBackup = true
				dateHeader := cached.Header.Get("Date")
				log.WithFields(logFields).Debugf("there is a cached value with date: %s", dateHeader)
				t, err := time.Parse("Mon, _2 Jan 2006 15:04:05 MST", dateHeader)
				if err != nil {
					log.WithFields(logFields).WithError(err).Warn("cannot parse date header of cached value")
				} else {
					minDate := time.Now().Add(-time.Duration(cfg.CacheDurationRegular))
					if t.After(minDate) {
						hitCacheRegular = true
						log.WithFields(logFields).Infof("cached value is younger than %s - using cached value", cfg.CacheDurationRegular)
						for k, v := range cached.Header {
							for i, val := range v {
								if i == 0 {
									rw.Header().Set(k, val)
								} else {
									rw.Header().Add(k, val)
								}
							}
						}
						rw.Header().Set("X-Cache", "HIT")
						rw.WriteHeader(cached.StatusCode)
						rw.Write(cached.Body)
						finishLog(logFields, start, hitCacheRegular, hitCacheBackup)
						DurationRequestProcessingHistogram.Observe(time.Since(start).Seconds())
						return
					} else {
						log.WithFields(logFields).Infof("cached value is older than %s - ignoring cached value", cfg.CacheDurationRegular)
					}
				}
			}
		}

		duration := time.Since(start)
		log.WithFields(logFields).WithFields(DurationLogFields(duration)).Info("processing request finished")
		DurationRequestProcessingHistogram.Observe(duration.Seconds())

		r.Host = upstreamURL.Host
		p.ServeHTTP(rw, r)
		finishLog(logFields, start, hitCacheRegular, hitCacheBackup)
	}
}

func finishLog(logFields logrus.Fields, start time.Time, hitCacheRegular, hitCacheBackup bool) {
	duration := time.Since(start)
	DurationOverallHistogram.Observe(duration.Seconds())
	if hitCacheBackup {
		BackupCacheHitCounter.Inc()
	} else {
		BackupCacheMissCounter.Inc()
	}
	if hitCacheRegular {
		RegularCacheHitServeCounter.Inc()
	} else {
		RegularCacheMissCounter.Inc()
	}
	log.
		WithFields(logFields).
		WithFields(DurationLogFields(duration)).
		WithField("hit_cache_regular", hitCacheRegular).
		WithField("hit_cache_backup", hitCacheBackup).
		Info("request finished")
}
