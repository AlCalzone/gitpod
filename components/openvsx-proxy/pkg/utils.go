// Copyright (c) 2021 Gitpod GmbH. All rights reserved.
// Licensed under the GNU Affero General Public License (AGPL).
// See License-AGPL.txt in the project root for license information.

package pkg

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gitpod-io/gitpod/common-go/log"
	"github.com/sirupsen/logrus"
	"golang.org/x/xerrors"
)

func DurationLogFields(duration time.Duration) logrus.Fields {
	return logrus.Fields{
		"duration":       duration,
		"duration_human": duration.String(),
	}
}

func key(r *http.Request) (string, error) {
	bh, err := bodyHash(r)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s %s %d %s", r.Method, r.URL, r.ContentLength, bh), nil
}

func bodyHash(r *http.Request) (string, error) {
	if r == nil || r.Body == nil {
		return "", xerrors.Errorf("request or body is nil")
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	bodyHash := hash(body)
	if log.Log.Level >= logrus.DebugLevel {
		if utf8.Valid(body) {
			bodyStr := string(body)
			truncatedSuffix := ""
			if len(bodyStr) > 500 {
				truncatedSuffix = "... [truncated]"
			}
			log.WithField("bodyHash", bodyHash).Debugf("body of bodyhash '%s': %.500s%s", bodyHash, bodyStr, truncatedSuffix)
		} else {
			log.WithField("bodyHash", bodyHash).Debugf("body of bodyhash '%s' is binary", bodyHash)
		}
	}
	return bodyHash, nil
}

func hash(v []byte) string {
	h := sha1.New()
	h.Write(v)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}
