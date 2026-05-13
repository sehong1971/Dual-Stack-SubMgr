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

//
// EXAMPLE HOW TO HAVE RMR STUB
//

package teststubdummy

import (
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrDummyStub struct {
	teststub.RmrStubControl
	reqMsg  int
	respMsg int
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewRmrDummyStub(desc string, srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc, stat string, mtypeseed int) *RmrDummyStub {
	dummyStub := &RmrDummyStub{}
	dummyStub.RmrStubControl.Init(desc, srcId, rtgSvc, stat, mtypeseed)
	dummyStub.reqMsg = mtypeseed + 1
	dummyStub.respMsg = mtypeseed + 2
	return dummyStub
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (tc *RmrDummyStub) SendReq(t *testing.T, plen int) {
	tc.Debug("SendReq")
	len := plen
	if len == 0 {
		len = 100
	}
	params := &xapp.RMRParams{}
	params.Mtype = tc.reqMsg
	params.SubId = -1

	params.Payload = make([]byte, len)
	params.PayloadLen = 0
	params.Meid = &xapp.RMRMeid{RanName: "TEST"}
	params.Xid = "TEST"
	params.Mbuf = nil

	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "%s", snderr.Error())
	}
	return
}

func (tc *RmrDummyStub) SendResp(t *testing.T, plen int) {
	tc.Debug("SendReq")
	len := plen
	if len == 0 {
		len = 100
	}
	params := &xapp.RMRParams{}
	params.Mtype = tc.respMsg
	params.SubId = -1
	params.Payload = make([]byte, len)
	params.PayloadLen = 0
	params.Meid = &xapp.RMRMeid{RanName: "TEST"}
	params.Xid = "TEST"
	params.Mbuf = nil

	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "%s", snderr.Error())
	}
	return
}

func (tc *RmrDummyStub) RecvReq(t *testing.T) bool {
	tc.Debug("RecvReq")

	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != tc.reqMsg {
			tc.TestError(t, "Received wrong mtype expected %d got %d, error", tc.reqMsg, msg.Mtype)
			return false
		}
		return true
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return false
}

func (tc *RmrDummyStub) RecvResp(t *testing.T) bool {
	tc.Debug("RecvResp")

	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != tc.respMsg {
			tc.TestError(t, "Received wrong mtype expected %d got %d, error", tc.respMsg, msg.Mtype)
			return false
		}
		return true
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return false
}
