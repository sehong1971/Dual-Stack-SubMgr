/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package control

import (
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RequestId struct {
	e2ap.RequestId
}

func (rid *RequestId) String() string {
	return "reqid(" + rid.RequestId.String() + ")"
}

type Sdlnterface interface {
	Set(ns string, pairs ...interface{}) error
	Get(ns string, keys []string) (map[string]interface{}, error)
	GetAll(ns string) ([]string, error)
	Remove(ns string, keys []string) error
	RemoveAll(ns string) error
}

type E2SubscriptionDirectives struct {
	// How many times E2 subscription request is retried
	// Required: true
	// Maximum: 10
	// Minimum: 0
	E2MaxTryCount int64

	// How long time response is waited from E2 node
	// Maximum: 10s
	// Minimum: 1s
	E2TimeoutTimerValue time.Duration

	// Subscription needs RMR route from E2Term to xApp
	CreateRMRRoute bool
}

type ErrorInfo struct {
	ErrorCause  string
	ErrorSource string
	TimeoutType string
}

func (e *ErrorInfo) SetInfo(errorCause string, errorSource string, timeoutType string) {
	e.ErrorCause = errorCause
	e.ErrorSource = errorSource
	e.TimeoutType = timeoutType
}

type XappRnibInterface interface {
	XappRnibSubscribe(cb func(string, ...string), channel string) error
	XappRnibGetListGnbIds() ([]*xapp.RNIBNbIdentity, xapp.RNIBIRNibError)
	XappRnibStoreAndPublish(channel string, event string, pairs ...interface{}) error
	XappRnibGetNodeb(inventoryName string) (*xapp.RNIBNodebInfo, xapp.RNIBIRNibError)
}
