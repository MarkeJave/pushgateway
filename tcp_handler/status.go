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
	"encoding/base64"
	"github.com/golang/protobuf/proto"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/prometheus/common/version"
	"github.com/prometheus/pushgateway/storage"
	. "github.com/prometheus/pushgateway/tcp_server"
)

type data struct {
	MetricGroups storage.GroupingKeyToMetricGroup
	Flags        map[string]string
	BuildInfo    map[string]string
	Birth        time.Time
	PathPrefix   string
	counter      int
}

func (d *data) Count() int {
	d.counter++
	return d.counter
}

func (data) FormatTimestamp(ts int64) string {
	return time.Unix(ts/1000, ts%1000*1000000).String()
}

// Status serves the status page.
//
// The returned handler is already instrumented for Prometheus.
func Status(
	ms storage.MetricStore,
	root http.FileSystem,
	flags map[string]string,
	pathPrefix string,
	logger log.Logger,
) func(*Session, *Package)  {
	birth := time.Now()

	return InstrumentWithCounter(
		"status", func(session *Session, pkg *Package) {
		t := template.New("status")
		t.Funcs(template.FuncMap{
			"value": func(f float64) string {
				return strconv.FormatFloat(f, 'f', -1, 64)
			},
			"timeFormat": func(t time.Time) string {
				return t.Format(time.RFC3339)
			},
			"base64": func(s string) string {
				return base64.RawURLEncoding.EncodeToString([]byte(s))
			},
		})

		f, err := root.Open("template.html")
		if err != nil {
			level.Error(logger).Log("msg", "error loading template.html", "err", err.Error())
			return
		}
		defer f.Close()
		tpl, err := ioutil.ReadAll(f)
		if err != nil {
			level.Error(logger).Log("msg", "error reading template.html", "err", err.Error())
			return
		}
		_, err = t.Parse(string(tpl))
		if err != nil {
			level.Error(logger).Log("msg", "error parsing template", "err", err.Error())
			return
		}

		buildInfo := map[string]string{
			"version":    version.Version,
			"revision":   version.Revision,
			"branch":     version.Branch,
			"buildUser":  version.BuildUser,
			"buildDate":  version.BuildDate,
			"goVersion":  version.GoVersion,
			"pathPrefix": pathPrefix,
			"birth":      birth.String(),
		}
		response := &MapResponse{
			Map: buildInfo,
		}

		//d := &data{
		//	MetricGroups: ms.GetMetricFamiliesMap(),
		//	BuildInfo:    buildInfo,
		//	Birth:        birth,
		//	PathPrefix:   pathPrefix,
		//	Flags:        flags,
		//}

		body, err := proto.Marshal(response)
		if err != nil {
			level.Error(logger).Log("msg", "Failed marshal response", "err", err)
			return
		}

		pkg = NewResponse(pkg.GetId(), KindResponse, body)
		if err != nil {
			level.Error(logger).Log(err.Error())
		}

		err = session.GetConn().SendPackage(pkg)
		if err != nil {
			level.Error(logger).Log("msg", "Failed send response", "err", err)
		}
	})
}
