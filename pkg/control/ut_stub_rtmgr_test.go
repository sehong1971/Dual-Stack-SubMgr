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
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_models"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type HttpEventWaiter struct {
	teststub.TestWrapper
	resultChan   chan bool
	nextActionOk bool
	sleep        int
}

func (msg *HttpEventWaiter) SetResult(res bool) {
	msg.resultChan <- res
}

func (msg *HttpEventWaiter) WaitResult(t *testing.T) bool {
	select {
	case result := <-msg.resultChan:
		return result
	case <-time.After(15 * time.Second):
		msg.TestError(t, "Waiter not received result status from case within 15 secs")
		return false
	}
	msg.TestError(t, "Waiter error in default branch")
	return false
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type testingHttpRtmgrStub struct {
	sync.Mutex
	teststub.TestWrapper
	port        string
	eventWaiter *HttpEventWaiter
}

func (tc *testingHttpRtmgrStub) NextEvent(eventWaiter *HttpEventWaiter) {
	tc.Lock()
	defer tc.Unlock()
	tc.eventWaiter = eventWaiter
}

func (tc *testingHttpRtmgrStub) AllocNextEvent(nextAction bool) *HttpEventWaiter {
	eventWaiter := &HttpEventWaiter{
		resultChan:   make(chan bool),
		nextActionOk: nextAction,
	}
	eventWaiter.TestWrapper.Init("localhost:" + tc.port)
	tc.NextEvent(eventWaiter)
	return eventWaiter
}

func (tc *testingHttpRtmgrStub) AllocNextSleep(sleep int, nextAction bool) *HttpEventWaiter {
	eventWaiter := &HttpEventWaiter{
		resultChan:   make(chan bool),
		nextActionOk: nextAction,
		sleep:        sleep,
	}
	eventWaiter.TestWrapper.Init("localhost:" + tc.port)
	tc.NextEvent(eventWaiter)
	return eventWaiter
}

func (tc *testingHttpRtmgrStub) http_handler(w http.ResponseWriter, r *http.Request) {

	tc.Lock()
	defer tc.Unlock()
	var id int32 = -1

	if r.Method == http.MethodPost || r.Method == http.MethodDelete {
		var req rtmgr_models.XappSubscriptionData
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			tc.Error("%s", err.Error())
		}
		tc.Debug("handling SubscriptionID=%d Address=%s Port=%d", *req.SubscriptionID, *req.Address, *req.Port)
		id = *req.SubscriptionID
	}
	if r.Method == http.MethodPut {
		var req rtmgr_models.XappList
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			tc.Error("%s", err.Error())
		}
		tc.Debug("handling put")
	}

	var code int = 0
	switch r.Method {
	case http.MethodPost:
		code = 201
		if tc.eventWaiter != nil {
			if tc.eventWaiter.nextActionOk == false {
				code = 400
			}
			if tc.eventWaiter.sleep != 0 {
				<-time.After(time.Duration(tc.eventWaiter.sleep) * time.Millisecond)
				tc.Debug("sleeping done, %v", id)
			}
		}
	case http.MethodDelete:
		code = 200
		if tc.eventWaiter != nil {
			if tc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	case http.MethodPut:
		code = 201
		if tc.eventWaiter != nil {
			if tc.eventWaiter.nextActionOk == false {
				code = 400
			}
		}
	default:
		code = 200
	}

	waiter := tc.eventWaiter
	tc.eventWaiter = nil
	if waiter != nil {
		waiter.SetResult(true)
	}
	tc.Debug("Method=%s Reply with code %d", r.Method, code)
	w.WriteHeader(code)

}

func (tc *testingHttpRtmgrStub) run() {
	http.HandleFunc("/", tc.http_handler)
	http.ListenAndServe("localhost:"+tc.port, nil)
}

func createNewHttpRtmgrStub(desc string, port string) *testingHttpRtmgrStub {
	tc := &testingHttpRtmgrStub{}
	tc.port = port
	tc.TestWrapper.Init(desc)
	return tc
}
