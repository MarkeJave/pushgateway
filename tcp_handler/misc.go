// Copyright 2017 The Prometheus Authors
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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/pushgateway/storage"
	. "github.com/prometheus/pushgateway/tcp_server"
)

// Healthy is used to report the health of the Pushgateway. It currently only
// uses the Healthy method of the MetricScore to detect healthy state.
//
// The returned handler is already instrumented for Prometheus.
func Healthy(ms storage.MetricStore, logger log.Logger) func(*Session, *Package) {
	return InstrumentWithCounter(
		"healthy",
		func(session *Session, pkg *Package) {
			err := ms.Healthy()

			var response *Package
			if err == nil{
				response, _ = NewStateResponse(pkg.GetId(), CodeSuccess)
			} else {
				response, _ = NewStateResponse(pkg.GetId(), CodeFailed)
			}
			if err != nil {
				level.Error(logger).Log(err.Error())
			}
			session.GetConn().SendPackage(response)

		},
	)
}

// Ready is used to report if the Pushgateway is ready to process requests. It
// currently only uses the Ready method of the MetricScore to detect ready
// state.
//
// The returned handler is already instrumented for Prometheus.
func Ready(ms storage.MetricStore, logger log.Logger) func(*Session, *Package) {
	return InstrumentWithCounter(
		"ready",
		func(session *Session, pkg *Package) {
			err := ms.Ready()

			var response *Package
			if err == nil{
				response, _ = NewStateResponse(pkg.GetId(), CodeSuccess)
			} else {
				response, _ = NewStateResponse(pkg.GetId(), CodeFailed)
			}
			if err != nil {
				level.Error(logger).Log(err.Error())
			}
			session.GetConn().SendPackage(response)
		},
	)
}
