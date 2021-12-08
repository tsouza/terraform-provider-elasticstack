// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.
//
// Code generated from specification version 7.16.0: DO NOT EDIT

package esapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func newLicensePostStartBasicFunc(t Transport) LicensePostStartBasic {
	return func(o ...func(*LicensePostStartBasicRequest)) (*Response, error) {
		var r = LicensePostStartBasicRequest{}
		for _, f := range o {
			f(&r)
		}
		return r.Do(r.ctx, t)
	}
}

// ----- API Definition -------------------------------------------------------

// LicensePostStartBasic - Starts an indefinite basic license.
//
// See full documentation at https://www.elastic.co/guide/en/elasticsearch/reference/master/start-basic.html.
//
type LicensePostStartBasic func(o ...func(*LicensePostStartBasicRequest)) (*Response, error)

// LicensePostStartBasicRequest configures the License Post Start Basic API request.
//
type LicensePostStartBasicRequest struct {
	Acknowledge *bool

	Pretty     bool
	Human      bool
	ErrorTrace bool
	FilterPath []string

	Header http.Header

	ctx context.Context
}

// Do executes the request and returns response or error.
//
func (r LicensePostStartBasicRequest) Do(ctx context.Context, transport Transport) (*Response, error) {
	var (
		method string
		path   strings.Builder
		params map[string]string
	)

	method = "POST"

	path.Grow(len("/_license/start_basic"))
	path.WriteString("/_license/start_basic")

	params = make(map[string]string)

	if r.Acknowledge != nil {
		params["acknowledge"] = strconv.FormatBool(*r.Acknowledge)
	}

	if r.Pretty {
		params["pretty"] = "true"
	}

	if r.Human {
		params["human"] = "true"
	}

	if r.ErrorTrace {
		params["error_trace"] = "true"
	}

	if len(r.FilterPath) > 0 {
		params["filter_path"] = strings.Join(r.FilterPath, ",")
	}

	req, err := newRequest(method, path.String(), nil)
	if err != nil {
		return nil, err
	}

	if len(params) > 0 {
		q := req.URL.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	if len(r.Header) > 0 {
		if len(req.Header) == 0 {
			req.Header = r.Header
		} else {
			for k, vv := range r.Header {
				for _, v := range vv {
					req.Header.Add(k, v)
				}
			}
		}
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	res, err := transport.Perform(req)
	if err != nil {
		return nil, err
	}

	response := Response{
		StatusCode: res.StatusCode,
		Body:       res.Body,
		Header:     res.Header,
	}

	return &response, nil
}

// WithContext sets the request context.
//
func (f LicensePostStartBasic) WithContext(v context.Context) func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		r.ctx = v
	}
}

// WithAcknowledge - whether the user has acknowledged acknowledge messages (default: false).
//
func (f LicensePostStartBasic) WithAcknowledge(v bool) func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		r.Acknowledge = &v
	}
}

// WithPretty makes the response body pretty-printed.
//
func (f LicensePostStartBasic) WithPretty() func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		r.Pretty = true
	}
}

// WithHuman makes statistical values human-readable.
//
func (f LicensePostStartBasic) WithHuman() func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		r.Human = true
	}
}

// WithErrorTrace includes the stack trace for errors in the response body.
//
func (f LicensePostStartBasic) WithErrorTrace() func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		r.ErrorTrace = true
	}
}

// WithFilterPath filters the properties of the response body.
//
func (f LicensePostStartBasic) WithFilterPath(v ...string) func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		r.FilterPath = v
	}
}

// WithHeader adds the headers to the HTTP request.
//
func (f LicensePostStartBasic) WithHeader(h map[string]string) func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		for k, v := range h {
			r.Header.Add(k, v)
		}
	}
}

// WithOpaqueID adds the X-Opaque-Id header to the HTTP request.
//
func (f LicensePostStartBasic) WithOpaqueID(s string) func(*LicensePostStartBasicRequest) {
	return func(r *LicensePostStartBasicRequest) {
		if r.Header == nil {
			r.Header = make(http.Header)
		}
		r.Header.Set("X-Opaque-Id", s)
	}
}
