// Copyright 2014 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcp_handler

import (
	"bytes"
	"encoding/base64"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/prometheus/pushgateway/storage"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	. "github.com/prometheus/pushgateway/tcp_server"
)


// Delete returns a handler that accepts delete requests.
//
// The returned handler is already instrumented for Prometheus.
func Delete(ms storage.MetricStore, jobBase64Encoded bool, logger log.Logger) func(*Session, *Package) {
	var mtx sync.Mutex // Protects ps.

	return InstrumentWithCounter(
		"delete", func(session *Session, pkg *Package) {
		mtx.Lock()
		action := &DeleteAction{}
		reader := bytes.NewReader(pkg.GetBody())

		_, err := pbutil.ReadDelimited(reader, action)
		if err != nil {
			if err == io.EOF {
				err = nil
			}
		}

		job := *action.Job
		if jobBase64Encoded {
			var err error
			if job, err = _decodeBase64(job); err != nil {
				level.Debug(logger).Log("msg", "invalid base64 encoding in job name", "job", job, "err", err.Error())
				return
			}
		}
		labels := action.Labels

		mtx.Unlock()

		if err != nil {
			level.Debug(logger).Log("msg", "failed to parse URL", "url", labels, "err", err.Error())
			return
		}
		if job == "" {
			level.Debug(logger).Log("msg", "job name is required")
			return
		}
		labels["job"] = job
		ms.SubmitWriteRequest(storage.WriteRequest{
			Labels:    labels,
			Timestamp: time.Now(),
		})

		pkg, err = NewSuccessResponse(pkg.GetId())
		if err != nil {
			level.Error(logger).Log(err.Error())
		}
		session.GetConn().SendPackage(pkg)
	})
}

// decodeBase64 decodes the provided string using the “Base 64 Encoding with URL
// and Filename Safe Alphabet” (RFC 4648). Padding characters (i.e. trailing
// '=') are ignored.
func _decodeBase64(s string) (string, error) {
	b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
	return string(b), err
}