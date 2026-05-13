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
	"fmt"
	"strings"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/nodeb-rnib.git/entities"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststube2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/assert"
)

// In below test cases there is done only one retry for E2 messages
// In Helm chart retry count is currently 2 By default. Retry count
// used in test cases is defined in submgr-config.yaml file.

func TestSuiteSetup(t *testing.T) {
	// The effect of this call shall endure though the UT suite!
	// If this causes any issues, the previous interface can be restored
	// like this:git log
	// SetPackerIf(e2ap_wrapper.NewAsn1E2APPacker())

	mainCtrl.InitAllCounterMap()
	SetPackerIf(e2ap_wrapper.NewUtAsn1E2APPacker())
	mainCtrl.c.restDuplicateCtrl.Init()

}
func TestRanStatusChangeViaSDLNotification(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cE2StateChangedToUp, 3},
	})

	// Current UT test cases use these ran names
	xappRnibMock.CreateGnb("RAN_NAME_1", entities.ConnectionStatus_DISCONNECTED)
	xappRnibMock.CreateGnb("RAN_NAME_11", entities.ConnectionStatus_DISCONNECTED)
	xappRnibMock.CreateGnb("RAN_NAME_2", entities.ConnectionStatus_DISCONNECTED)

	mainCtrl.c.e2IfState.ReadE2ConfigurationFromRnib()
	mainCtrl.c.e2IfState.SubscribeChannels()

	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTED")
	mainCtrl.SetE2State(t, "RAN_NAME_2_CONNECTED")
	mainCtrl.SetE2State(t, "RAN_NAME_11_CONNECTED")

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAfterE2ConnBreak
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |         [E2 Conn. DOWN]        |
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |     RESTSubFail |              |
//     |<----------------|              |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAfterE2ConnBreak(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestReqRejDueE2Down, 1},
		Counter{cE2StateChangedToDown, 1},
		Counter{cE2StateChangedToUp, 1},
	})

	// E2 disconnect after E2term has received response
	mainCtrl.SetE2State(t, "RAN_NAME_1_DISCONNECTED")
	// Req
	const subReqCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	xappConn1.SendRESTSubsReq(t, params)

	// Restore E2 connection for following test cases
	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTED")

	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqE2ConnBreak
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |         [E2 Conn. DOWN]        |
//     |        [Int. SUBS DELETE]      |
//     |                 |              |
//     |      RESTNotif(unsuccessful)   |
//     |<----------------|              |
//     |                 |              |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqE2ConnBreak(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cE2StateChangedToDown, 1},
		Counter{cE2StateChangedToUp, 1},
	})

	// Req
	const subReqCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)

	// E2 disconnect after E2term has received response
	mainCtrl.SetE2State(t, "RAN_NAME_1_DISCONNECTED")

	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	<-time.After(time.Second * 1)
	assert.Equal(t, 0, len(mainCtrl.c.registry.register))
	assert.Equal(t, 0, len(mainCtrl.c.registry.restSubscriptions))

	subIds, register, err := mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		xapp.Logger.Error("%v", err)
	} else {
		assert.Equal(t, 65534, len(subIds)) // range 1-65535 , FFFF = 65535
		assert.Equal(t, 0, len(register))
	}

	restSubscriptions, err := mainCtrl.c.ReadAllRESTSubscriptionsFromSdl()
	if err != nil {
		xapp.Logger.Error("%v", err)
	} else {
		assert.Equal(t, 0, len(restSubscriptions))
	}

	// Restore E2 connection for following test cases
	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTED")

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubscriptionDeleteAfterE2ConnectionBreak
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |           [E2 Conn. DOWN]      |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |  [No valid subscription found] |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubscriptionDeleteAfterE2ConnectionBreak(t *testing.T) {
	xapp.Logger.Debug("TEST: TestRESTSubscriptionDeleteAfterE2ConnectionBreak")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cE2StateChangedToDown, 1},
		Counter{cE2StateChangedToUp, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// E2 disconnect after E2term has received response
	mainCtrl.SetE2State(t, "RAN_NAME_1_DISCONNECTED")

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	<-time.After(time.Second * 1)
	assert.Equal(t, 0, len(mainCtrl.c.registry.register))
	assert.Equal(t, 0, len(mainCtrl.c.registry.restSubscriptions))

	subIds, register, err := mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		xapp.Logger.Error("%v", err)
	} else {
		assert.Equal(t, 65534, len(subIds)) // range 1-65535 , FFFF = 65535
		assert.Equal(t, 0, len(register))
	}

	restSubscriptions, err := mainCtrl.c.ReadAllRESTSubscriptionsFromSdl()
	if err != nil {
		xapp.Logger.Error("%v", err)
	} else {
		assert.Equal(t, 0, len(restSubscriptions))
	}

	// Restore E2 connection for following test cases
	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTED")

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTOtherE2ConnectionChanges
//

//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |  [E2 CONNECTED_SETUP_FAILED]   |
//     |         [E2 CONNECTING]        |
//     |        [E2 SHUTTING_DOWN]      |
//     |          [E2 SHUT_DOWN]        |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTOtherE2ConnectionChanges(t *testing.T) {
	xapp.Logger.Debug("TEST: TestRESTOtherE2ConnectionChanges")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cE2StateChangedToUp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	// Submgr should not react to any other connection state changes than CONNECTED and DISCONNECTED
	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTED_SETUP_FAILED")
	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTING")
	mainCtrl.SetE2State(t, "RAN_NAME_1_SHUTTING_DOWN")
	mainCtrl.SetE2State(t, "RAN_NAME_1_SHUT_DOWN")

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Restore E2 connection for following test cases
	mainCtrl.SetE2State(t, "RAN_NAME_1_CONNECTED")

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndDeleteOkWithE2apUtWrapper
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|  // The order of these events may vary
//     |                 |              |             |
//     |     RESTSubResp |              |             |  // The order of these events may vary
//     |<----------------|              |             |
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 |              |             |
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |              |             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndDeleteOkWithE2apUtWrapper(t *testing.T) {

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, nil)

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndE1apDeleteReqPackingError
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|  // The order of these events may vary
//     |                 |              |             |
//     |     RESTSubResp |              |             |  // The order of these events may vary
//     |<----------------|              |             |
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 |              |             |
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |              |             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 |              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndE1apDeleteReqPackingError(t *testing.T) {

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, nil)

	e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_REQ, false)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	defer e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_REQ, true)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndE2APDeleteRespUnpackingError
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|  // The order of these events may vary
//     |                 |              |             |
//     |     RESTSubResp |              |             |  // The order of these events may vary
//     |<----------------|              |             |
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 |              |             |
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |              |             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             | // The order of these events may vary
//     |                 |              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             | // 1.st NOK
//     |                 |              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             | // 2.nd NOK
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndE2APDeleteRespUnpackingError(t *testing.T) {

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, nil)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_RESP, false)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	defer e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_RESP, true)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndRouteNok
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    |  rtmgr  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | RouteCreate  |
//     |              |------------->|
//     |              |              |
//     |              | RouteCreate  |
//     |              |  status:400  |
//     |              |<-------------|
//     |              |              |
//     |       [SUBS INT DELETE]     |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndRouteNok(t *testing.T) {
	CaseBegin("TestSubReqAndRouteNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cRouteCreateFail, 1},
	})

	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)
	xappConn1.SendSubsReq(t, nil, nil)
	waiter.WaitResult(t)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	<-time.After(1 * time.Second)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndRouteUpdateNok

//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    |  rtmgr  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |        [SUBS CREATE]       |              |
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     | SubReq (mergeable)         |              |
//     |--------------------------->|              |              |
//     |             |              |              |
//     |             |              | RouteUpdate  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | RouteUpdate  |
//     |             |              |  status:400  |
//     |             |              |<-------------|
//     |             |              |              |
//     |       [SUBS INT DELETE]    |              |
//     |             |              |              |
//     |             |              |              |
//     |        [SUBS DELETE]       |              |
//     |             |              |              |

func TestSubReqAndRouteUpdateNok(t *testing.T) {
	CaseBegin("TestSubReqAndRouteUpdateNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cRouteCreateUpdateFail, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)
	xappConn2.SendSubsReq(t, nil, nil)
	waiter.WaitResult(t)

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubDelReqAndRouteDeleteNok
//
//   stub                          stub
// +-------+     +---------+    +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |    |  rtmgr  |
// +-------+     +---------+    +---------+    +---------+
//     |              |              |              |
//     |         [SUBS CREATE]       |              |
//     |              |              |              |
//     |              |              |              |
//     |              |              |              |
//     | SubDelReq    |              |              |
//     |------------->|              |              |
//     |              |  SubDelReq   |              |
//     |              |------------->|              |
//     |              |  SubDelRsp   |              |
//     |              |<-------------|              |
//     |  SubDelRsp   |              |              |
//     |<-------------|              |              |
//     |              | RouteDelete  |              |
//     |              |---------------------------->|
//     |              |              |              |
//     |              | RouteDelete  |              |
//     |              |  status:400  |              |
//     |              |<----------------------------|
//     |              |              |              |
func TestSubDelReqAndRouteDeleteNok(t *testing.T) {
	CaseBegin("TestSubDelReqAndRouteDeleteNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cRouteDeleteFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	waiter := rtmgrHttp.AllocNextEvent(false)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	waiter.WaitResult(t)

	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubMergeDelAndRouteUpdateNok
//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              | RouteUpdate  |
//     |             |              |-----> rtmgr  |
//     |             |              |              |
//     |             |              | RouteUpdate  |
//     |             |              |  status:400  |
//     |             |              |<----- rtmgr  |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubMergeDelAndRouteUpdateNok(t *testing.T) {
	CaseBegin("TestSubMergeDelAndRouteUpdateNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cRouteDeleteUpdateFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
		Counter{cUnmergedSubscriptions, 1},
	})

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	//Del1
	waiter := rtmgrHttp.AllocNextEvent(false)
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	waiter.WaitResult(t)

	xappConn1.RecvSubsDelResp(t, deltrans1)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOk(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOk")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkOutofOrderIEs
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp | (Out of Order IEs)
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkOutofOrderIEs(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkOutofOrderIEs")

	mainCtrl.c.e2ap.SetE2IEOrderCheck(0)
	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	if cretrans == nil {
		t.Logf("Could not send SubsReq")
		t.FailNow()
	}
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	if crereq == nil || cremsg == nil {
		t.Logf("Could not recieve SubsReq")
		t.FailNow()
	}

	e2termConn1.SendSubsResp(t, crereq, cremsg)

	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)
	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.c.e2ap.SetE2IEOrderCheck(1)
}

//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// TestSubReqRetransmission
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |  SubReq      |              |
//     | (retrans)    |              |
//     |------------->|              |
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |         [SUBS DELETE]       |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqRetransmission(t *testing.T) {
	CaseBegin("TestSubReqRetransmission")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.SendSubsReq(t, nil, cretrans) //Retransmitted SubReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// hack as there is no real way to see has message be handled.
	// Previuos counter check just tells that is has been received by submgr
	// --> artificial delay
	<-time.After(1 * time.Second)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqRetransmission
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     | SubDelReq    |              |
//     | (same sub)   |              |
//     | (same xid)   |              |
//     |------------->|              |
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubDelReqRetransmission(t *testing.T) {
	CaseBegin("TestSubDelReqRetransmission")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.SendSubsDelReq(t, deltrans, e2SubsId) //Retransmitted SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// hack as there is no real way to see has message be handled.
	// Previuos counter check just tells that is has been received by submgr
	// --> artificial delay
	<-time.After(1 * time.Second)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqCollision
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq 1  |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq 1  |
//     |              |------------->|
//     |              |              |
//     | SubDelReq 2  |              |
//     | (same sub)   |              |
//     | (diff xid)   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelResp 1 |
//     |              |<-------------|
//     |              |              |
//     | SubDelResp 1 |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelResp 2 |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqCollision(t *testing.T) {
	CaseBegin("TestSubDelReqCollision")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)

	// Subs Delete colliding
	seqBef := mainCtrl.get_msgcounter(t)
	deltranscol2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsDelReq(t, deltranscol2, e2SubsId) //Colliding SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// hack as there is no real way to see has message be handled.
	// Previuos counter check just tells that is has been received by submgr
	// --> artificial delay
	<-time.After(1 * time.Second)

	// Del resp for first and second
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// don't care in which order responses are received
	xappConn1.RecvSubsDelResp(t, nil)
	xappConn1.RecvSubsDelResp(t, nil)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkTwoParallel
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp  |  | xapp  |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          |              |              |
//     |          |              |              |
//     |          | SubReq1      |              |
//     |          |------------->|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       SubReq2           |              |
//     |------------------------>|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          |    SubResp1  |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       SubResp2          |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkTwoParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkTwoParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()

	rparams2.Req.EventTriggerDefinition.Data.Length = 1
	rparams2.Req.EventTriggerDefinition.Data.Data = make([]uint8, rparams2.Req.EventTriggerDefinition.Data.Length)
	rparams2.Req.EventTriggerDefinition.Data.Data[0] = 2

	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Resp2
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSameSubsDiffRan
// Same subscription to different RANs
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq(r1)   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq(r1)   |
//     |              |------------->|
//     |              |              |
//     |              | SubResp(r1)  |
//     |              |<-------------|
//     |              |              |
//     | SubResp(r1)  |              |
//     |<-------------|              |
//     |              |              |
//     | SubReq(r2)   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq(r2)   |
//     |              |------------->|
//     |              |              |
//     |              | SubResp(r2)  |
//     |              |<-------------|
//     |              |              |
//     | SubResp(r2)  |              |
//     |<-------------|              |
//     |              |              |
//     |       [SUBS r1 DELETE]      |
//     |              |              |
//     |       [SUBS r2 DELETE]      |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSameSubsDiffRan(t *testing.T) {
	CaseBegin("TestSameSubsDiffRan")

	//Req1
	cretrans1 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsReq(t, nil, cretrans1)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	cretrans2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_2")
	xappConn1.SendSubsReq(t, nil, cretrans2)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsDelReq(t, deltrans1, e2SubsId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_2")
	xappConn1.SendSubsDelReq(t, deltrans2, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn1.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqRetryInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |         [SUBS DELETE]       |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqRetryInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqRetryInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq and send SubsResp
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	// Xapp: Receive SubsResp
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqTwoRetriesNoRespSubDelRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqRetryNoRespSubDelRespInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqTwoRetriesNoRespSubDelRespInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	// Xapp: Send SubsReq
	xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Send receive SubsDelReq and send SubsResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqTwoRetriesNoRespAtAllInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqTwoRetriesNoRespAtAllInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqTwoRetriesNoRespAtAllInSubmgr start")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
	})

	// Xapp: Send SubsReq
	xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq
	delreq, _ := e2termConn1.RecvSubsDelReq(t)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.InstanceId, 15)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqSubFailRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubFail |
//     |              |<-------------|
//     |              |              |
//     |      SubFail |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqSubFailRespInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqSubFailRespInSubmgr start")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cSubFailToXapp, 1},
	})

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// Xapp: Receive SubsFail
	e2SubsId := xappConn1.RecvSubsFail(t, cretrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqSubFailRespInSubmgrOutofOrderIEs
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubFail | (Out of Order IEs)
//     |              |<-------------|
//     |              |              |
//     |      SubFail |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqSubFailRespInSubmgrOutofOrderIEs(t *testing.T) {
	CaseBegin("TestSubReqSubFailRespInSubmgrOutofOrderIEs start")

	mainCtrl.c.e2ap.SetE2IEOrderCheck(0)
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cSubFailToXapp, 1},
	})

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// Xapp: Receive SubsFail
	e2SubsId := xappConn1.RecvSubsFail(t, cretrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.c.e2ap.SetE2IEOrderCheck(1)
}

//-----------------------------------------------------------------------------
// TestSubDelReqRetryInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqRetryInSubmgr(t *testing.T) {

	CaseBegin("TestSubDelReqRetryInSubmgr start")

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Subs Delete
	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqTwoRetriesNoRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqTwoRetriesNoRespInSubmgr(t *testing.T) {

	CaseBegin("TestSubDelReTwoRetriesNoRespInSubmgr start")

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Subs Delete
	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqSubDelFailRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     |  SubDelReq   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelFail |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqSubDelFailRespInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqSubDelFailRespInSubmgr start")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubDelReqSubDelFailRespInSubmgrOutofOrderIEs
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     |  SubDelReq   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelFail | (Out of Order IEs)
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqSubDelFailRespInSubmgrOutofOrderIEs(t *testing.T) {
	CaseBegin("TestSubReqSubDelFailRespInSubmgr start")

	mainCtrl.c.e2ap.SetE2IEOrderCheck(0)
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.c.e2ap.SetE2IEOrderCheck(1)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameAction
//
//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameAction(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameAction")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	xappConn1.RecvSubsDelResp(t, deltrans1)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameActionParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Resp2
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	xappConn1.RecvSubsDelResp(t, deltrans1)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelNokSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |    SubFail1  |
//     |             |              |<-------------|
//     |             |              |              |
//     |             |    SubFail1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubFail2          |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNokSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelNokSameActionParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)

	// E2t: Receive SubsReq (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	// E2t: send SubsFail (first)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	//Fail1
	e2SubsId1 := xappConn1.RecvSubsFail(t, cretrans1)
	//Fail2
	xappConn2.RecvSubsFail(t, cretrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 15)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelNoAnswerSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             | SubReq2      |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              |              |
//     |             |              | SubDelReq    |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              |   SubDelResp |
//     |             |              |<-------------|
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNoAnswerSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelNoAnswerSameActionParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	xappConn1.SendSubsReq(t, rparams1, nil)

	crereq1, _ := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	xappConn2.SendSubsReq(t, rparams2, nil)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	//Req1 (retransmitted)
	e2termConn1.RecvSubsReq(t)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq1.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 15)
}

//-----------------------------  Policy cases ---------------------------------
//-----------------------------------------------------------------------------
// TestSubReqPolicyAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqPolicyAndSubDelOk(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOk")

	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypePolicy
	cretrans := xappConn1.SendSubsReq(t, rparams1, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqPolicyChangeAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqPolicyChangeAndSubDelOk(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOk")

	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypePolicy
	cretrans := xappConn1.SendSubsReq(t, rparams1, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Policy change
	rparams1.Req.RequestId.InstanceId = e2SubsId
	rparams1.Req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitW200ms
	xappConn1.SendSubsReq(t, rparams1, cretrans)

	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId = xappConn1.RecvSubsResp(t, cretrans)
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkTwoE2termParallel
//
//   stub                          stub           stub
// +-------+     +---------+    +---------+    +---------+
// | xapp  |     | submgr  |    | e2term1 |    | e2term2 |
// +-------+     +---------+    +---------+    +---------+
//     |              |              |              |
//     |              |              |              |
//     |              |              |              |
//     | SubReq1      |              |              |
//     |------------->|              |              |
//     |              |              |              |
//     |              | SubReq1      |              |
//     |              |------------->|              |
//     |              |              |              |
//     | SubReq2      |              |              |
//     |------------->|              |              |
//     |              |              |              |
//     |              | SubReq2      |              |
//     |              |---------------------------->|
//     |              |              |              |
//     |              |    SubResp1  |              |
//     |              |<-------------|              |
//     |    SubResp1  |              |              |
//     |<-------------|              |              |
//     |              |    SubResp2  |              |
//     |              |<----------------------------|
//     |    SubResp2  |              |              |
//     |<-------------|              |              |
//     |              |              |              |
//     |        [SUBS 1 DELETE]      |              |
//     |              |              |              |
//     |        [SUBS 2 DELETE]      |              |
//     |              |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkTwoE2termParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkTwoE2termParallel")

	//Req1
	cretrans1 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsReq(t, nil, cretrans1)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	cretrans2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_11")
	xappConn1.SendSubsReq(t, nil, cretrans2)
	crereq2, cremsg2 := e2termConn2.RecvSubsReq(t)

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Resp2
	e2termConn2.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn1.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn1.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	e2termConn2.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqInsertAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqInsertAndSubDelOk(t *testing.T) {
	CaseBegin("TestInsertSubReqAndSubDelOk")

	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeInsert
	cretrans := xappConn1.SendSubsReq(t, rparams1, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqRetransmissionWithSameSubIdDiffXid
//
// This case simulates case where xApp restarts and starts sending same
// subscription requests which have already subscribed successfully

//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     | xApp restart |              |
//     |              |              |
//     |  SubReq      |              |
//     | (retrans with same xApp generated subid but diff xid)
//     |------------->|              |
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |         [SUBS DELETE]       |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqRetransmissionWithSameSubIdDiffXid(t *testing.T) {
	CaseBegin("TestSubReqRetransmissionWithSameSubIdDiffXid")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// xApp restart here
	// --> artificial delay
	<-time.After(1 * time.Second)

	//Subs Create
	cretrans = xappConn1.SendSubsReq(t, nil, nil) //Retransmitted SubReq
	e2SubsId = xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqNokAndSubDelOkWithRestartInMiddle
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |                        <----|
//     |                             |
//     |        Submgr restart       |
//     |                             |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqNokAndSubDelOkWithRestartInMiddle(t *testing.T) {
	CaseBegin("TestSubReqNokAndSubDelOkWithRestartInMiddle")

	mainCtrl.SetResetTestFlag(t, true) // subs.DoNotWaitSubResp will be set TRUE for the subscription
	xappConn1.SendSubsReq(t, nil, nil)
	e2termConn1.RecvSubsReq(t)
	mainCtrl.SetResetTestFlag(t, false)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})
	e2SubsId := uint32(resp[0].SubscriptionID)
	t.Logf("e2SubsId = %v", e2SubsId)

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// Submgr send delete for uncompleted subscription
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkWithRestartInMiddle
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |                             |
//     |        Submgr restart       |
//     |                             |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkWithRestartInMiddle(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkWithRestartInMiddle")

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Check subscription
	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// That needs to be completed before successful subscription query is possible
	<-time.After(time.Second * 1)

	// Check that subscription is restored correctly after restart
	resp, _ = xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionWithRestartsInMiddle
//
//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |                                           |
//     |              submgr restart               |
//     |                                           |
//     |             |              |              |
//     |             |              |              |
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |             |              |              |
//     |                                           |
//     |              submgr restart               |
//     |                                           |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkSameActionWithRestartsInMiddle(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameActionWithRestartsInMiddle")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	// Check subscription
	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// That needs to be completed before successful subscription query is possible
	<-time.After(time.Second * 1)

	// Check that subscription is restored correctly after restart
	resp, _ = xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	xapp.Logger.Debug("xappConn1.RecvSubsDelResp")
	xappConn1.RecvSubsDelResp(t, deltrans1)
	xapp.Logger.Debug("xappConn1.RecvSubsDelResp received")

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// Submgr need be ready before successful subscription deletion is possible
	<-time.After(time.Second * 1)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//*****************************************************************************
//  REST interface test cases
//*****************************************************************************

//-----------------------------------------------------------------------------
// Test debug GET and POST requests
//
//   curl
// +-------+     +---------+
// | user  |     | submgr  |
// +-------+     +---------+
//     |              |
//     | GET/POST Req |
//     |------------->|
//     |         Resp |
//     |<-------------|
//     |              |
func TestGetSubscriptions(t *testing.T) {

	mainCtrl.SendGetRequest(t, "localhost:8088", "/ric/v1/subscriptions")
}

func TestGetSymptomData(t *testing.T) {

	mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/symptomdata")
}

func TestPostdeleteSubId(t *testing.T) {

	mainCtrl.SendPostRequest(t, "localhost:8080", "/ric/v1/test/deletesubid=1")
}

func TestPostEmptyDb(t *testing.T) {

	mainCtrl.SendPostRequest(t, "localhost:8080", "/ric/v1/test/emptydb")
}

func TestGetRestSubscriptions(t *testing.T) {

	mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/restsubscriptions")
}

//-----------------------------------------------------------------------------
// TestDelAllE2nodeSubsViaDebugIf
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | REST get_all_e2nodes           |             |
//     |---------------->|              |             |
//     |    OK 200       |              |             |
//     |<----------------|              |             |
//     | REST delete_all_e2node_subscriptions         | ranName = RAN_NAME_1
//     |---------------->|              |             |
//     |    OK 200       |              |             |
//     |<----------------|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//     |                 | RouteDelete  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//
//-----------------------------------------------------------------------------

func TestDelAllE2nodeSubsViaDebugIf(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	e2nodesJson := mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/get_all_e2nodes")

	var e2nodesList []string
	err := json.Unmarshal(e2nodesJson, &e2nodesList)
	if err != nil {
		t.Errorf("Unmarshal error: %s", err)
	}
	assert.Equal(t, true, mainCtrl.VerifyStringExistInSlice("RAN_NAME_1", e2nodesList))

	e2RestSubsJson := mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/get_e2node_rest_subscriptions/RAN_NAME_1") // RAN_NAME_1 = ranName
	var e2RestSubsMap map[string]RESTSubscription
	err = json.Unmarshal(e2RestSubsJson, &e2RestSubsMap)
	if err != nil {
		t.Errorf("Unmarshal error: %s", err)
	}

	if len(e2RestSubsMap) != 1 {
		t.Errorf("Incorrect e2RestSubsMap length %v", len(e2RestSubsMap))
	}

	// Simulate deletion through REST test and debug interface
	mainCtrl.SendDeleteRequest(t, "localhost:8080", "/ric/v1/delete_all_e2node_subscriptions/RAN_NAME_1") // RAN_NAME_1 = ranName
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestDelAllxAppSubsViaDebugIf
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | REST get_all_xapps             |             |
//     |---------------->|              |             |
//     |    OK 200       |              |             |
//     |<----------------|              |             |
//     | REST delete_all_xapp_subscriptions           |  xappServiceName = localhost
//     |---------------->|              |             |
//     |    OK 200       |              |             |
//     |<----------------|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//     |                 | RouteDelete  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//
//-----------------------------------------------------------------------------

func TestDelAllxAppSubsViaDebugIf(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	xappsJson := mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/get_all_xapps")

	var xappList []string
	err := json.Unmarshal(xappsJson, &xappList)
	if err != nil {
		t.Errorf("Unmarshal error: %s", err)
	}
	assert.Equal(t, true, mainCtrl.VerifyStringExistInSlice("localhost", xappList))

	// Simulate deletion through REST test and debug interface
	mainCtrl.SendDeleteRequest(t, "localhost:8080", "/ric/v1/delete_all_xapp_subscriptions/localhost") // localhost = xappServiceName
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestDelViaxAppSubsIf
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | REST get_xapp_rest_restsubscriptions         |
//     |---------------->|              |             |
//     |    OK 200       |              |             |
//     |<----------------|              |             |
//     | RESTSudDel      |              |             |
//     |---------------->|              |             | Via user curl command (port 8088)
//     |     RESTSudDel  |              |             |
//     |<----------------|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//     |                 | RouteDelete  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//
//-----------------------------------------------------------------------------

func TestDelViaxAppSubsIf(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	restSubsListJson := mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/get_xapp_rest_restsubscriptions/localhost") // localhost = xappServiceName

	var restSubsMap map[string]RESTSubscription
	err := json.Unmarshal(restSubsListJson, &restSubsMap)
	if err != nil {
		t.Errorf("Unmarshal error: %s", err)
	}
	_, ok := restSubsMap[restSubId]
	if !ok {
		t.Errorf("REST subscription not found. restSubId=%s", restSubId)
	}

	var e2Subscriptions []Subscription
	e2SubscriptionsJson := mainCtrl.SendGetRequest(t, "localhost:8080", "/ric/v1/get_e2subscriptions/"+restSubId)
	err = json.Unmarshal(e2SubscriptionsJson, &e2Subscriptions)
	if err != nil {
		t.Errorf("Unmarshal error: %s", err)
	}
	if len(e2Subscriptions) != 1 {
		t.Errorf("Incorrect e2Subscriptions length %v", len(e2Subscriptions))
	}

	// Simulate deletion through xapp REST test interface
	mainCtrl.SendDeleteRequest(t, "localhost:8088", "/ric/v1/subscriptions/"+restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndRouteNok
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | rtmgr   |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | RouteCreate  |
//     |                 |------------->|
//     |                 | RouteCreate  |
//     |                 |  status:400  |
//     |                 |(Bad request) |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     |          [SUBS INT DELETE]     |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndRouteNok(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cRouteCreateFail, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	// Add delay for rtmgt HTTP handling so that HTTP response is received before notify on XAPP side
	waiter := rtmgrHttp.AllocNextSleep(50, false)
	newSubsId := mainCtrl.get_registry_next_subid(t)

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xappConn1.ExpectRESTNotificationNok(t, restSubId, "failAll")
	waiter.WaitResult(t)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndRouteUpdateNok
//
//   stub        stub                         stub           stub
// +-------+   +-------+    +---------+    +---------+    +---------+
// | xapp1 |   | xapp2 |    | submgr  |    | rtmgr   |    | e2term  |
// +-------+   +-------+    +---------+    +---------+    +---------+
//     |           |             |              |              |
//     | RESTSubReq1             |              |              |
//     |------------------------>|              |              |
//     |     RESTSubResp2        |              |              |
//     |<------------------------|              |              |
//     |           |             |              |              |
//     |           |             | RouteCreate  |              |
//     |           |             |------------->|              |
//     |           |             | CreateResp   |              |
//     |           |             |<-------------|              |
//     |           |             | SubReq       |              |
//     |           |             |---------------------------->|
//     |           |             |      SubResp |              |
//     |           |             |<----------------------------|
//     |      RESTNotif1         |              |              |
//     |<------------------------|              |              |
//     |           |             |              |              |
//     |           | RESTSubReq2 |              |              |
//     |           |------------>|              |              |
//     |           | RESTSubResp2|              |              |
//     |           |<------------|              |              |
//     |           |             | RouteUpdate  |              |
//     |           |             |------------->|              |
//     |           |             | RouteUpdate  |              |
//     |           |             |  status:400  |              |
//     |           |             |(Bad request) |              |
//     |           |             |<-------------|              |
//     |           | RESTNotif2(unsuccessful)   |              |
//     |           |<------------|              |              |
//     |           |             |              |              |
//     |          [SUBS INT DELETE]             |              |
//     |           |             |              |              |
//     | RESTSubDelReq1          |              |              |
//     |------------------------>|              |              |
//     |  RESTSubDelResp1        |              |              |
//     |<------------------------|              |              |
//     |           |             |              |              |
//     |           |             |        [SUBS DELETE]        |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndRouteUpdateNok(t *testing.T) {

	//Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRouteCreateUpdateFail, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 2},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	// Subs create for xapp1
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// xapp2 ROUTE creation shall fail with  400 from rtmgr -> submgr
	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for second subscriber : %v", restSubId2)
	xappConn2.ExpectRESTNotificationNok(t, restSubId2, "allFail")
	waiter.WaitResult(t)
	xappConn2.WaitRESTNotification(t, restSubId2)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)

	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqAndRouteDeleteNok
//
//   stub                             stub           stub
// +-------+        +---------+    +---------+    +---------+
// | xapp  |        | submgr  |    | rtmgr   |    | e2term  |
// +-------+        +---------+    +---------+    +---------+
//     |                 |              |              |
//     | RESTSubReq      |              |              |
//     |---------------->|              |              |
//     |                 |              |              |
//     |     RESTSubResp |              |              |
//     |<----------------|              |              |
//     |                 | SubReq       |              |
//     |                 |---------------------------->|
//     |                 | SubResp      |              |
//     |                 |<----------------------------|
//     |       RESTNotif |              |              |
//     |<----------------|              |              |
//     |                 |              |              |
//     |                 |              |              |
//     | RESTSubDelReq   |              |              |
//     |---------------->|              |              |
//     |  RESTSubDelResp |              |              |
//     |<----------------|              |              |
//     |                 | SubSelReq    |              |
//     |                 |---------------------------->|
//     |                 | SubSelResp   |              |
//     |                 |<----------------------------|
//     |                 | RouteDelete  |              |
//     |                 |------------->|              |
//     |                 | Routedelete  |              |
//     |                 |  status:400  |              |
//     |                 |(Bad request) |              |
//     |                 |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubDelReqAndRouteDeleteNok(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRouteDeleteFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	waiter := rtmgrHttp.AllocNextEvent(false)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	waiter.WaitResult(t)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubMergeDelAndRouteUpdateNok
//
//   stub        stub                         stub           stub
// +-------+   +-------+    +---------+    +---------+    +---------+
// | xapp1 |   | xapp2 |    | submgr  |    | rtmgr   |    | e2term  |
// +-------+   +-------+    +---------+    +---------+    +---------+
//     |           |             |              |              |
//     | RESTSubReq1             |              |              |
//     |------------------------>|              |              |
//     |     RESTSubResp2        |              |              |
//     |<------------------------|              |              |
//     |           |             |              |              |
//     |           |             | RouteCreate  |              |
//     |           |             |------------->|              |
//     |           |             | CreateResp   |              |
//     |           |             |<-------------|              |
//     |           |             | SubReq       |              |
//     |           |             |---------------------------->|
//     |           |             |      SubResp |              |
//     |           |             |<----------------------------|
//     |      RESTNotif1         |              |              |
//     |<------------------------|              |              |
//     |           |             |              |              |
//     |           | RESTSubReq2 |              |              |
//     |           |------------>|              |              |
//     |           | RESTSubResp2|              |              |
//     |           |<------------|              |              |
//     |           |             | RouteCreate  |              |
//     |           |             |------------->|              |
//     |           |             | CreateResp   |              |
//     |           |             |<-------------|              |
//     |           |             | SubReq       |              |
//     |           |             |---------------------------->|
//     |           |             |      SubResp |              |
//     |           |             |<----------------------------|
//     |           | RESTNotif2  |              |              |
//     |           |<------------|              |              |
//     |           |             |              |              |
//     |          [SUBS INT DELETE]             |              |
//     |           |             |              |              |
//     | RESTSubDelReq1          |              |              |
//     |------------------------>|              |              |
//     |  RESTSubDelResp1        |              |              |
//     |<------------------------|              |              |
//     |           |             | SubDelReq    |              |
//     |           |             |---------------------------->|
//     |           |             | SubDelResp   |              |
//     |           |             |<----------------------------|
//     |           |             | RouteUpdate  |              |
//     |           |             |------------->|              |
//     |           |             | RouteUpdate  |              |
//     |           |             |  status:400  |              |
//     |           |             |(Bad request) |              |
//     |           |             |<-------------|              |
//     |           |             |              |              |
//     |           | RESTSubDelReq2             |              |
//     |           |------------>|              |              |
//     |           | RESTSubDelResp2            |              |
//     |           |<------------|              |              |
//     |           |             | SubDelReq    |              |
//     |           |             |---------------------------->|
//     |           |             | SubdelResp   |              |
//     |           |             |<----------------------------|
//     |           |             | RouteDelete  |              |
//     |           |             |------------->|              |
//     |           |             | Deleteresp   |              |
//     |           |             |<-------------|              |

//-----------------------------------------------------------------------------

func TestRESTSubMergeDelAndRouteUpdateNok(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cRouteDeleteUpdateFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 2},
		Counter{cUnmergedSubscriptions, 1},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})
	restSubId2, e2SubsId2 := createXapp2MergedSubscription(t, "RAN_NAME_1")

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560", "localhost:13660"})

	//Del1, this shall fail on rtmgr side
	waiter := rtmgrHttp.AllocNextEvent(false)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	waiter.WaitResult(t)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13660"})

	//Del2
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId2, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqRetransmission
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq1     |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq1      |
//     |                 |------------->|
//     |                 |              |
//     | RESTSubReq2     |              |
//     | (retrans)       |              |
//     |---------------->|              |
//     | RESTSubResp(201)|              |
//     |<----------------|              |
//     |                 |              |
//     |                 |     SubResp1 |
//     |                 |<-------------|
//     |      RESTNotif1 |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetransmission(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})
	// Retry/duplicate will get the same way as the first request.
	// Contianed duplicate messages from same xapp will not be merged. Here we use xappConn2 to simulate sending
	// second request from same xapp as doing it from xappConn1 would not work as notification would not be received

	// Subs Create
	const subReqCount int = 1

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	xappConn1.SendRESTSubsReq(t, params)
	<-time.After(time.Second * 1)

	xappConn1.WaitListedRestNotifications(t, []string{restSubId})

	// Depending one goroutine scheduling order, we cannot say for sure which xapp reaches e2term first. Thus
	// the order is not significant here.
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	e2SubsId := <-xappConn1.ListedRESTNotifications

	xapp.Logger.Debug("TEST: XAPP notification received e2SubsId=%v", e2SubsId)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	mainCtrl.wait_multi_subs_clean(t, []uint32{e2SubsId.E2SubsId}, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [RETRANS1]      |              |             |
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [RETRANS2]      |              |             |
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetransmissionV2(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 3},
		Counter{cDuplicateE2SubReq, 2},
		Counter{cRestSubRespToXapp, 3},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 3},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	mainCtrl.WaitOngoingRequestMapEmpty()

	//1.st resend
	restSubId_resend := xappConn1.SendRESTSubsReq(t, params)

	assert.Equal(t, restSubId_resend, restSubId)

	mainCtrl.WaitOngoingRequestMapEmpty()

	//2.nd resend
	restSubId_resend2 := xappConn1.SendRESTSubsReq(t, params)

	assert.Equal(t, restSubId_resend2, restSubId)

	mainCtrl.WaitOngoingRequestMapEmpty()

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [RETRANS, with RESTsubsId]     |             |
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [RETRANS, without RESTsubsId]  |             |
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqRetransmissionV3(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 3},
		Counter{cDuplicateE2SubReq, 2},
		Counter{cRestSubRespToXapp, 3},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 3},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	mainCtrl.WaitOngoingRequestMapEmpty()

	//1.st resend with subscription ID
	params.SetSubscriptionID(&restSubId)
	restSubId_resend := xappConn1.SendRESTSubsReq(t, params)

	assert.Equal(t, restSubId_resend, restSubId)

	mainCtrl.WaitOngoingRequestMapEmpty()

	//2.nd resend without subscription ID (faking app restart)
	params = xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId_resend2 := xappConn1.SendRESTSubsReq(t, params)

	assert.Equal(t, restSubId_resend2, restSubId)

	mainCtrl.WaitOngoingRequestMapEmpty()

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [with RestSUbsId + one additional e2 subDetail]
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     | [for initial e2 subDetail]     |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [with RESTsubsId initial request]            |
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetransmissionV4(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 3},
		Counter{cDuplicateE2SubReq, 2},
		Counter{cRestSubRespToXapp, 3},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 4},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	mainCtrl.WaitOngoingRequestMapEmpty()

	// Send modified  requst, this time with e2 subscriptions.
	params2 := xappConn1.GetRESTSubsReqReportParams(subReqCount + 1)
	params2.SetSubscriptionID(&restSubId)

	xapp.Subscription.SetResponseCB(xappConn1.SubscriptionRespHandler)
	xappConn1.ExpectAnyNotification(t)
	// Resend the original request with an additional e2 subscription (detail), this time with restsubsid
	restSubId_resend := xappConn1.SendRESTSubsReq(t, params2)
	e2SubsId1 := xappConn1.WaitAnyRESTNotification(t)
	assert.Equal(t, e2SubsId, e2SubsId1)

	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	xappConn1.DecrementRequestCount()
	xappConn1.ExpectRESTNotification(t, restSubId_resend)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.WaitRESTNotification(t, restSubId_resend)
	assert.NotEqual(t, e2SubsId2, 0)

	mainCtrl.WaitOngoingRequestMapEmpty()

	xapp.Subscription.SetResponseCB(xappConn1.SubscriptionRespHandler)
	params = xappConn1.GetRESTSubsReqReportParams(subReqCount)
	params.SetSubscriptionID(&restSubId)
	xappConn1.ExpectAnyNotification(t)
	// Resend the original request again with only one e2 subscription (detail), this time with restsubsid
	restSubId_resend2 := xappConn1.SendRESTSubsReq(t, params)
	assert.Equal(t, restSubId_resend, restSubId_resend2)

	e2SubsId1 = xappConn1.WaitAnyRESTNotification(t)
	assert.Equal(t, e2SubsId, e2SubsId1)

	mainCtrl.WaitOngoingRequestMapEmpty()

	// Delete both e2 subscriptions
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	e2SubsIds := []uint32{e2SubsId, e2SubsId2}
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [with RestSUbsId + one additional e2 subDetail]
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     | [for initial e2 subDetail]     |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [without RESTsubsId initial request]         |
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetransmissionV5(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 3},
		Counter{cDuplicateE2SubReq, 2},
		Counter{cRestSubRespToXapp, 3},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 4},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	mainCtrl.WaitOngoingRequestMapEmpty()

	// Send modified  request, this time with e2 subscriptions.
	params2 := xappConn1.GetRESTSubsReqReportParams(subReqCount + 1)
	params2.SetSubscriptionID(&restSubId)

	xapp.Subscription.SetResponseCB(xappConn1.SubscriptionRespHandler)
	xappConn1.ExpectAnyNotification(t)
	// Resend the original request with an additional e2 subscription (detail), this time with restsubsid
	restSubId_resend := xappConn1.SendRESTSubsReq(t, params2)

	e2SubsId1 := xappConn1.WaitAnyRESTNotification(t)
	assert.Equal(t, e2SubsId, e2SubsId1)
	// The first E2 subscription returns immediately, manually decrement expected request count for the remaining request handling
	xappConn1.DecrementRequestCount()

	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId_resend)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.WaitRESTNotification(t, restSubId_resend)
	assert.NotEqual(t, e2SubsId2, 0)

	mainCtrl.WaitOngoingRequestMapEmpty()

	xapp.Subscription.SetResponseCB(xappConn1.SubscriptionRespHandler)
	params = xappConn1.GetRESTSubsReqReportParams(subReqCount)
	xappConn1.ExpectAnyNotification(t)
	// Resend the original request again with only one e2 subscription (detail), WITHOUT restsubsid
	// md5sum shall find the original request
	restSubId_resend2 := xappConn1.SendRESTSubsReq(t, params)
	assert.Equal(t, restSubId_resend, restSubId_resend2)

	e2SubsId1 = xappConn1.WaitAnyRESTNotification(t)
	assert.Equal(t, e2SubsId, e2SubsId1)

	mainCtrl.WaitOngoingRequestMapEmpty()

	// Delete both e2 subscriptions
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	e2SubsIds := []uint32{e2SubsId, e2SubsId2}
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubReq      |              |             |
//     | [with RestSUbsId + one additional e2 subDetail]
//     |---------------->|              |             |
//     |      RESTNotif1 |              |             |
//     | [for initial e2 subDetail]     |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     | RESTSubReq      |              |             |
//     | [with RESTsubsId initial request]            |
//     |---------------->|              |             |
//     |     RESTSubResp |              |             |
//     |<----------------|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|
//     |                 | SubReq       |             |
//     |                 |------------->|             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqRetransmissionV6(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 3},
		Counter{cDuplicateE2SubReq, 1},
		Counter{cRestSubRespToXapp, 3},
		Counter{cSubReqToE2, 3},
		Counter{cSubRespFromE2, 3},
		Counter{cRestSubNotifToXapp, 4},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 3},
		Counter{cSubDelRespFromE2, 3},
		Counter{cRestSubDelRespToXapp, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	mainCtrl.WaitOngoingRequestMapEmpty()

	// Send modified  requst, this time with e2 subscriptions.
	params2 := xappConn1.GetRESTSubsReqReportParams(subReqCount + 1)
	params2.SetSubscriptionID(&restSubId)

	xapp.Subscription.SetResponseCB(xappConn1.SubscriptionRespHandler)
	xappConn1.ExpectAnyNotification(t)
	// Resend the original request with an additional e2 subscription (detail), this time with restsubsid
	restSubId_resend := xappConn1.SendRESTSubsReq(t, params2)

	e2SubsId1 := xappConn1.WaitAnyRESTNotification(t)
	assert.Equal(t, e2SubsId, e2SubsId1)

	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId_resend)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.WaitRESTNotification(t, restSubId_resend)
	assert.NotEqual(t, e2SubsId2, 0)

	mainCtrl.WaitOngoingRequestMapEmpty()

	// Delete both e2 subscriptions
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	e2SubsIds := []uint32{e2SubsId, e2SubsId2}
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	waitSubsCleanup(t, e2SubsId, 10)

	// Resend the original request, we shall find it's previous md5sum/restsubs
	// but the restsubscription has been already removed. This shall trigger a
	// fresh create.
	restSubId, e2SubsId = createSubscription(t, xappConn1, e2termConn1, params)

	mainCtrl.WaitOngoingRequestMapEmpty()

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqRetransmission
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |      SubResp |
//     |                 |<-------------|
//     |      RESTNotif1 |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubDelReqRetransmission(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 2},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	//Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	//Resend delete req
	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// Del resp
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqDelReq
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |     unsuccess   |              |
//     |<----------------|              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |      RESTNotif1 |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqDelReq(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cRestSubDelFailToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	// Del. This will fail as processing of the subscription
	// is still ongoing in submgr. Deletion is not allowed before
	// subscription creation has been completed.
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	// Retry del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkTwoParallel
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndSubDelOkTwoParallel(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	//Req1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send 1st REST subscriber request for subscriberId : %v", restSubId1)

	//Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send 2nd REST subscriber request for subscriberId : %v", restSubId2)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	//XappConn1 receives both of the  responses
	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	//Resp2
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)

	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Debug("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Debug("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	//Del1
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId1)
	//Del2
	deleteSubscription(t, xappConn2, e2termConn1, &restSubId2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsIdA.E2SubsId, 10)
	waitSubsCleanup(t, e2SubsIdB.E2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSameSubsDiffRan
// Same subscription to different RANs
//
//   stub                          stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq(r1)  |              |
//     |---------------->|              |
//     | RESTSubResp(r1) |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq(r1)   |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubResp(r1)  |
//     |                 |<-------------|
//     |                 |              |
//     | RESTNotif(r1)   |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq(r2)  |              |
//     |---------------->|              |
//     |                 |              |
//     | RESTSubResp(r2) |              |
//     |<----------------|              |
//     |                 | SubReq(r2)   |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubResp(r2)  |
//     |                 |<-------------|
//     |                 |              |
//     | RESTNotif(r2)   |              |
//     |<----------------|              |
//     |                 |              |
//     |          [SUBS r1 DELETE]      |
//     |                 |              |
//     |          [SUBS r2 DELETE]      |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSameSubsDiffRan(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send 1st REST subscriber request for subscriberId : %v", restSubId1)

	params = xappConn1.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_2")
	restSubId2, e2SubsId2 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send 2nd REST subscriber request for subscriberId : %v", restSubId2)

	//Del1
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId1)
	//Del2
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqRetryInSubmgr
//
//   stub                          stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     | RESTSubResp     |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 | SubResp      |
//     |                 |<-------------|
//     |                 |              |
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetryInSubmgr(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	xapp.Logger.Debug("Send REST subscriber request for subscriber : %v", restSubId)

	// Catch the first message and ignore it
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore REST subscriber request for subscriber : %v", restSubId)

	// The second request is being handled normally
	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// Del
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqTwoRetriesNoRespSubDelRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetryNoRespSubDelRespInSubmgr(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore 1st REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore 2nd REST subscriber request for subscriber : %v", restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.WaitRESTNotification(t, restSubId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	waitSubsCleanup(t, delreq.RequestId.InstanceId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestREST2eTermNotRespondingToSubReq
//
//   stub                          stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     | RESTSubResp     |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     | RESTNotif(Unsuccessful)        |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     | RESTSubDelResp  |              |
//     |<----------------|              |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestREST2eTermNotRespondingToSubReq(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReReqToE2, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore 1st REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore 2nd REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Debug("Ignore 1st INTERNAL delete request for subscriber : %v", restSubId)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")
	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Debug("Ignore 2nd INTERNAL delete request for subscriber : %v", restSubId)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqTwoRetriesNoRespSubDelRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqTwoRetriesNoRespAtAllInSubmgr(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore 1st REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Debug("Ignore 2nd REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Debug("Ignore 1st INTERNAL delete request for subscriber : %v", restSubId)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")
	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Debug("Ignore 2nd INTERNAL delete request for subscriber : %v", restSubId)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqSubFailRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqSubFailRespInSubmgr(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const e2Timeout int64 = 2
	const e2RetryCount int64 = 1
	const routingNeeded bool = true

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// REST subscription sill there to be deleted
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPartialResp
//
//   stub                          stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     | RESTSubResp     |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 | SubResp      | Partially accepted
//     |                 |<-------------|
//     |                 |              |
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPartialResp(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cPartialSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	actionId := int64(2)
	actionType := "report"
	actionDefinition := []int64{5678, 1}
	subsequestActionType := "continue"
	timeToWait := "w10ms"
	params.AppendActionToActionToBeSetupList(actionId, actionType, actionDefinition, subsequestActionType, timeToWait)

	restSubId := xappConn1.SendRESTSubsReq(t, params)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)

	actionNotAdmittedItem := e2ap.ActionNotAdmittedItem{}
	actionNotAdmittedItem.ActionId = 1
	actionNotAdmittedItem.Cause.Content = 1
	actionNotAdmittedItem.Cause.Value = 8
	actionNotAdmittedList := e2ap.ActionNotAdmittedList{}
	actionNotAdmittedList.Items = append(actionNotAdmittedList.Items, actionNotAdmittedItem)
	e2termConn1.SendPartialSubsResp(t, crereq, cremsg, actionNotAdmittedList)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// Del
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqRetryInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubDelReqRetryInSubmgr(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})
	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqTwoRetriesNoRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubDelReqTwoRetriesNoRespInSubmgr(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqSubDelFailRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubDelReqSubDelFailRespInSubmgr(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkSameAction
//
//   stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     |             |                 |    SubResp1  |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |                  RESTSubResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |      RESTNotif2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |                 | SubDelReq2   |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |  SubDelResp2 |
//     |             |                 |<-------------|
//     |             |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndSubDelOkSameAction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")

	xapp.Subscription.SetResponseCB(xappConn2.SubscriptionRespHandler)
	xappConn2.ExpectAnyNotification(t)
	waiter := rtmgrHttp.AllocNextSleep(10, true)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	waiter.WaitResult(t)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	e2SubsId2 := xappConn2.WaitAnyRESTNotification(t)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId2)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560", "localhost:13660"})

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	deleteXapp2Subscription(t, &restSubId2)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId2, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              | SubReq2      |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              |    SubResp2  |
//     |             |              |<-------------|
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
func TestRESTSubReqAndSubDelOkSameActionParallel(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	params2 := xappConn2.GetRESTSubsReqReportParams(subReqCount)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params2)

	xappConn1.ExpectRESTNotification(t, restSubId1)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.WaitRESTNotification(t, restSubId1)

	xappConn2.ExpectRESTNotification(t, restSubId2)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)

	waitSubsCleanup(t, e2SubsId2, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelNoAnswerSameActionParallel
//
//   stub          stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             |                 |              |
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubResp2    |              |
//     |<------------------------------|              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |              |
//     |             |                 | SubDelReq    |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |   SubDelResp |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |       unsuccess |              |
//     |             |<----------------|              |
//     |                    RESTNotif2 |              |
//     |             |       unsuccess |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndSubDelNoAnswerSameActionParallel(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubReReqToE2, 1},
		Counter{cRestSubFailNotifToXapp, 2},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 2},
	})
	const subReqCount int = 1

	// Req1
	params1 := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params1)
	crereq1, _ := e2termConn1.RecvSubsReq(t)

	// Req2
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	params2 := xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params2.SetMeid("RAN_NAME_1")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params2)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	//Req1 (retransmitted)
	e2termConn1.RecvSubsReq(t)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)

	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Debug("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Debug("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)

	mainCtrl.wait_multi_subs_clean(t, []uint32{e2SubsIdA.E2SubsId, e2SubsIdB.E2SubsId}, 10)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsIdA.E2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelNokSameActionParallel
//
//   stub          stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             |                 |              |
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 |    SubFail1  |
//     |             |                 |<-------------|
//     |             |                 |              |
//     |             |      RESTNotif1 |              |
//     |             |       unsuccess |              |
//     |             |<----------------|              |
//     |                    RESTNotif2 |              |
//     |             |       unsuccess |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |   There is no need for xApp to send delete for failed subscriptions but some xApp might do so.
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndSubDelNokSameActionParallel(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cRestSubFailNotifToXapp, 2},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	const subReqCount int = 1

	// Req1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	// Req2
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	params2 := xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params2.SetMeid("RAN_NAME_1")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params2)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	// E2t: send SubsFail (first)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})
	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Debug("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Debug("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsIdA.E2SubsId, 10)
	waitSubsCleanup(t, e2SubsIdB.E2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyAndSubDelOk
//
//   stub                          stub
// +-------+       +---------+    +---------+
// | xapp  |       | submgr  |    | e2term  |
// +-------+       +---------+    +---------+
//     |                |              |
//     | RESTSubReq     |              |
//     |--------------->|              |
//     |  RESTSubResp   |              |
//     |<---------------|              |
//     |                |              |
//     |                | SubReq       |
//     |                |------------->|
//     |                |              |
//     |                |      SubResp |
//     |                |<-------------|
//     |                |              |
//     |  RESTNotif     |              |
//     |<---------------|              |
//     |                |              |
//     |                |              |
//     | RESTSubDelReq  |              |
//     |--------------->|              |
//     | RESTSubDelResp |              |
//     |<---------------|              |
//     |                |              |
//     |                | SubDelReq    |
//     |                |------------->|
//     |                |              |
//     |                |   SubDelResp |
//     |                |<-------------|
//     |                |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPolicyAndSubDelOk(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyChangeAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPolicyChangeAndSubDelOk(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const e2Timeout int64 = 1
	const e2RetryCount int64 = 0
	const routingNeeded bool = true

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Policy change
	// GetRESTSubsReqPolicyParams sets some counters on tc side.

	params = xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	params.SetSubscriptionID(&restSubId)
	params.SetTimeToWait("w200ms")
	restSubId, e2SubsId = createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyChangeNokAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPolicyChangeNokAndSubDelOk(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const e2Timeout int64 = 1
	const e2RetryCount int64 = 0
	const routingNeeded bool = false

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)
	fmt.Printf("restSubId: %v", restSubId)

	// Policy change
	// GetRESTSubsReqPolicyParams sets some counters on tc side.
	params = xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	params.SetSubscriptionID(&restSubId)
	params.SetTimeToWait("w200ms")

	restSubId = xappConn1.SendRESTSubsReq(t, params)
	fmt.Printf("restSubId: %v", restSubId)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Gnb sends RICSubscriptionFailure
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 5) // CauseRIC / function-resource-limit
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyChangeNOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |         RESTSubUpdateFail(400 Bad request)
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPolicyChangeNOk(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubFailToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Policy change
	params = xappConn1.GetRESTSubsReqPolicyParams(subReqCount)

	restSubIdUpd := strings.ToUpper(restSubId) // This makes RESTSubReq to fail
	params.SetSubscriptionID(&restSubIdUpd)
	params.SetTimeToWait("w200ms")

	restSubId2 := xappConn1.SendRESTSubsReq(t, params)
	assert.Equal(t, restSubId2, "")

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkTwoE2termParallel
//
//   stub                             stub           stub
// +-------+        +---------+    +---------+    +---------+
// | xapp  |        | submgr  |    | e2term1 |    | e2term2 |
// +-------+        +---------+    +---------+    +---------+
//     |                 |              |              |
//     |                 |              |              |
//     |                 |              |              |
//     | RESTSubReq1     |              |              |
//     |---------------->|              |              |
//     |                 |              |              |
//     |    RESTSubResp1 |              |              |
//     |<----------------|              |              |
//     |                 | SubReq1      |              |
//     |                 |------------->|              |
//     |                 |              |              |
//     | RESTSubReq2     |              |              |
//     |---------------->|              |              |
//     |                 |              |              |
//     |    RESTSubResp2 |              |              |
//     |<----------------|              |              |
//     |                 | SubReq2      |              |
//     |                 |---------------------------->|
//     |                 |              |              |
//     |                 |    SubResp1  |              |
//     |                 |<-------------|              |
//     |      RESTNotif1 |              |              |
//     |<----------------|              |              |
//     |                 |    SubResp2  |              |
//     |                 |<----------------------------|
//     |      RESTNotif2 |              |              |
//     |<----------------|              |              |
//     |                 |              |              |
//     |           [SUBS 1 DELETE]      |              |
//     |                 |              |              |
//     |           [SUBS 2 DELETE]      |              |
//     |                 |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndSubDelOkTwoE2termParallel(t *testing.T) {

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	const subReqCount int = 1

	// Req1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_11")
	// Here we use xappConn2 to simulate sending second request from same xapp as doing it from xappConn1
	// would not work as notification would not be received
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	crereq2, cremsg2 := e2termConn2.RecvSubsReq(t)

	// Resp1
	xappConn1.ExpectRESTNotification(t, restSubId1)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.WaitRESTNotification(t, restSubId1)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId1=%v", e2SubsId1)

	// Resp2
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn2.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId2=%v", e2SubsId2)

	// Delete1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Delete2
	xappConn1.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq2, delmsg2)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqInsertAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |       ...       |     ...      |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqInsertAndSubDelOk(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	params.SetSubActionTypes("insert")

	// Req
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqNokAndSubDelOkWithRestartInMiddle
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | RESTSubReq   |              |
//     |------------->|              |
//     |              |              |
//     |  RESTSubResp |              |
//     |<-------------|              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |                        <----|
//     |                             |
//     |        Submgr restart       |
//     |                             |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |    RESTNotif |              |
//     |    unsuccess |              |
//     |<-------------|              |
//     |              |              |
//     | RESTSubDelReq|              |
//     |------------->|              |
//     |              |              |
//     |RESTSubDelResp|              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqNokAndSubDelOkWithRestartInMiddle(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)

	//Req
	mainCtrl.SetResetTestFlag(t, true) // subs.DoNotWaitSubResp will be set TRUE for the subscription
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)

	mainCtrl.SetResetTestFlag(t, false)

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// Deleletion of uncompleted subscription
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkWithRestartInMiddle
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     |                                |
//     |           Submgr restart       |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndSubDelOkWithRestartInMiddle(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Create subscription
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriber : %v", restSubId)

	// Check subscription
	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// That needs to be completed before successful subscription query is possible
	<-time.After(time.Second * 1)

	// Check subscription
	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// Delete subscription
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkSameActionWithRestartsInMiddle
//
//   stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     |             |                 |    SubResp1  |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |                  RESTSubResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |      RESTNotif2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |           Submgr restart       |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |             |           Submgr restart       |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |                 | SubDelReq2   |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |  SubDelResp2 |
//     |             |                 |<-------------|
//     |             |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelOkSameActionWithRestartsInMiddle(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Create subscription 1
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriber 1 : %v", restSubId1)

	// Create subscription 2 with same action
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	xapp.Subscription.SetResponseCB(xappConn2.SubscriptionRespHandler)
	xappConn2.ExpectAnyNotification(t)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	e2SubsId2 := xappConn2.WaitAnyRESTNotification(t)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId2)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560", "localhost:13660"})

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done 1")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// That needs to be completed before successful subscription delete is possible
	<-time.After(time.Second * 1)

	// Delete subscription 1, and wait until it has removed the first endpoint
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	mainCtrl.WaitRESTSubscriptionDelete(restSubId1)
	// Above wait does not work correctly anymore as this delay makes this test case work

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done 2")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// That needs to be completed before successful subscription query is possible
	<-time.After(time.Second * 1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13660"})

	// Delete subscription 2
	deleteXapp2Subscription(t, &restSubId2)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTReportSubReqAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 | SubReq       |   // Only one request sent in the teat case
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |       ...       |     ...      |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTReportSubReqAndSubDelOk(t *testing.T) {
	const subReqCount int = 1
	testIndex := 1
	RESTReportSubReqAndSubDelOk(t, subReqCount, testIndex)
}

func RESTReportSubReqAndSubDelOk(t *testing.T, subReqCount int, testIndex int) {
	xapp.Logger.Debug("TEST: TestRESTReportSubReqAndSubDelOk with testIndex %v", testIndex)

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	var e2SubsId []uint32
	for i := 0; i < subReqCount; i++ {
		crereq, cremsg := e2termConn1.RecvSubsReq(t)
		xappConn1.ExpectRESTNotification(t, restSubId)

		e2termConn1.SendSubsResp(t, crereq, cremsg)
		instanceId := xappConn1.WaitRESTNotification(t, restSubId)
		xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)
		e2SubsId = append(e2SubsId, instanceId)
		resp, _ := xapp.Subscription.QuerySubscriptions()
		assert.Equal(t, resp[i].SubscriptionID, (int64)(instanceId))
		assert.Equal(t, resp[i].Meid, "RAN_NAME_1")
		assert.Equal(t, resp[i].ClientEndpoint, []string{"localhost:13560"})

	}

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	for i := 0; i < subReqCount; i++ {
		delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
		e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	}

	// Wait that subs is cleaned
	for i := 0; i < subReqCount; i++ {
		mainCtrl.wait_subs_clean(t, e2SubsId[i], 10)
	}

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTTwoPolicySubReqAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTTwoPolicySubReqAndSubDelOk(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 2

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), 2)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTPolicySubReqAndSubDelOk19E2Subs
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |  ------
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |   E2 subscription x 19
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |  ------
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//     |                 | SubDelReq    |  ------
//     |                 |------------->|
//     |                 |              |   E2 subscription delete x 19
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |  ------
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTPolicySubReqAndSubDelOk19E2Subs(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 19},
		Counter{cSubRespFromE2, 19},
		Counter{cRestSubNotifToXapp, 19},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 19},
		Counter{cSubDelRespFromE2, 19},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 19
	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), 19)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTTwoPolicySubReqAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTTwoReportSubReqAndSubDelOk(t *testing.T) {

	subReqCount := 2

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cRestSubNotifToXapp, uint64(subReqCount)},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), subReqCount)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTTwoReportSubReqAndSubDelOkNoActParams
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTTwoReportSubReqAndSubDelOkNoActParams(t *testing.T) {

	subReqCount := 2

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cRestSubNotifToXapp, uint64(subReqCount)},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), subReqCount)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTReportSubReqAndSubDelOk19E2Subs
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |  ------
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |   E2 subscription x 19
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 |              |  ------
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//     |                 | SubDelReq    |  ------
//     |                 |------------->|
//     |                 |              |   E2 subscription delete x 19
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |  ------
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTReportSubReqAndSubDelOk19E2Subs(t *testing.T) {

	subReqCount := 19

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cRestSubNotifToXapp, uint64(subReqCount)},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
		Counter{cRestSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), subReqCount)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqReportSameActionDiffEventTriggerDefinitionLen
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqReportSameActionDiffEventTriggerDefinitionLen(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	eventTriggerDefinition := []int64{1234, 1}
	params.SetSubEventTriggerDefinition(eventTriggerDefinition)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqReportSameActionDiffActionListLen
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqReportSameActionDiffActionListLen(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")

	actionId := int64(1)
	actionType := "report"
	actionDefinition := []int64{5678, 1}
	subsequestActionType := "continue"
	timeToWait := "w10ms"
	params.AppendActionToActionToBeSetupList(actionId, actionType, actionDefinition, subsequestActionType, timeToWait)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqReportSameActionDiffActionID
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqReportSameActionDiffActionID(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	params.SetSubActionIDs(int64(2))

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqDiffActionType
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqDiffActionType(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	const e2Timeout int64 = 2
	const e2RetryCount int64 = 2
	const routingNeeded bool = true

	// Req1
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	params.SetMeid("RAN_NAME_1")

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyAndSubDelOkSameAction
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPolicyAndSubDelOkSameAction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	const e2Timeout int64 = 2
	const e2RetryCount int64 = 2
	const routingNeeded bool = true

	// Req1
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	params.SetMeid("RAN_NAME_1")

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqReportSameActionDiffActionDefinitionLen
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqReportSameActionDiffActionDefinitionLen(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	actionDefinition := []int64{5678, 1}
	params.SetSubActionDefinition(actionDefinition)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqReportSameActionDiffActionDefinitionContents
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqReportSameActionDiffActionDefinitionContents(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	actionDefinition := []int64{56782}
	params.SetSubActionDefinition(actionDefinition)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqReportSameActionDiffSubsAction
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp2 |  | xapp1 |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          | RESTSubReq1  |              |
//     |          |------------->|              |
//     |          | RESTSubResp1 |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       RESTSubReq2       |              |
//     |------------------------>|              |
//     |       RESTSubResp2      |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          | RESTNotif1   |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       RESTNotif2        |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqReportSameActionDiffSubsAction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cRestSubNotifToXapp, 2},
		Counter{cRestSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cRestSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount)
	params.SetMeid("RAN_NAME_1")
	params.SetTimeToWait("w200ms")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp | ASN.1 decode fails
//     |                 |<-------------| Decode failed. More data needed. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionResponseDecodeFail(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, cremsg, xapp.RIC_SUB_RESP)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp | Unknown instanceId
//     |                 |<-------------| No valid subscription found with subIds [0]
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------| No valid subscription found with subIds [0]
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionResponseUnknownInstanceId(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	const subReqCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Unknown instanceId 0 in this response which will result resending original request
	orgInstanceId := crereq.RequestId.InstanceId
	crereq.RequestId.InstanceId = 0
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Subscription already created in E2 Node. E2 Node responds with failure but there is also same unknown instanceId 0
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, orgInstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp | No transaction for the response
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|Ongoing transaction not found. This will result timer expiry and sending delete
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------| Ongoing transaction not found.
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionResponseNoTransaction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 2},
	})

	const subReqCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	mainCtrl.MakeTransactionNil(t, crereq.RequestId.InstanceId)
	// No transaction exist for this response which will result resending original request
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Resending happens because there no transaction
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionFailureDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | ASN.1 decode fails
//     |                 |<-------------| Decode failed. More data needed. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionFailureDecodeFail(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubFailFromE2, 2},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, cremsg, xapp.RIC_SUB_FAILURE)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Unknown instanceId
//     |                 |<-------------| No valid subscription found with subIds [0]. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|No valid subscription found with subIds [0]. This will result timer expiry and sending delete
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionFailureUnknownInstanceId(t *testing.T) {

	const subReqCount int = 1

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubReReqToE2, 1},
		Counter{cSubFailFromE2, 2},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Unknown instanceId 0 in this response which will result resending original request
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.Fail.RequestId.InstanceId = 0
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Subscription already created in E2 Node. E2 Node responds with failure but there is also same unknown instanceId 0
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionFailureNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | No transaction for the response
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and sending delete
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------| Ongoing transaction not found.
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionFailureNoTransaction(t *testing.T) {

	const subReqCount int = 1

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubReReqToE2, 1},
		Counter{cSubFailFromE2, 2},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 2},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	mainCtrl.MakeTransactionNil(t, crereq.RequestId.InstanceId)

	// No transaction exist for this response which will result resending original request
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	// Subscription already created in E2 Node.
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Resending happens because there no transaction
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", instanceId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteResponseDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp | ASN.1 decode fails.
//     |                 |<-------------| Decode failed. More data needed. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more in E2 node
//     |                 |<-------------|
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionDeleteResponseDecodeFail(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, delmsg, xapp.RIC_SUB_DEL_RESP)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteResponseUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp | Unknown instanceId
//     |                 |<-------------| No valid subscription found with subIds [0]. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more in E2 node
//     |                 |<-------------|
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionDeleteResponseUnknownInstanceId(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelFailFromE2, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Unknown instanceId in this response which will result resending original request
	delreq.RequestId.InstanceId = 0
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteResponseNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp | No transaction for the response
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more in E2 node
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionDeleteResponseNoTransaction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelFailFromE2, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	mainCtrl.MakeTransactionNil(t, e2SubsId)

	// No transaction exist for this response which will result resending original request
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteFailureDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | ASN.1 decode fails
//     |                 |<-------------| Decode failed. More data needed. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more in E2 node
//     |                 |<-------------|
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionDeleteFailureDecodeFail(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelFailFromE2, 2},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, delmsg, xapp.RIC_SUB_DEL_FAILURE)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteailureUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Unknown instanceId
//     |                 |<-------------| No valid subscription found with subIds [0]. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more in E2 node
//     |                 |<-------------| No valid subscription found with subIds [0].
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionDeleteailureUnknownInstanceId(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelFailFromE2, 2},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Unknown instanceId 0 in this response which will result resending original request
	delreq.RequestId.InstanceId = 0
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node. E2 Node responds with failure but there is also same unknown instanceId 0
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteFailureNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | No transaction for the response
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry and resending
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more in E2 node
//     |                 |<-------------| Ongoing transaction not found. This will result timer expiry
//     |                 |              |
//     |           [SUBS DELETE]        |
//     |                 |              |
//-----------------------------------------------------------------------------

func TestRESTUnpackSubscriptionDeleteFailureNoTransaction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelFailFromE2, 2},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	mainCtrl.MakeTransactionNil(t, e2SubsId)

	// No transaction exist for this response which will result resending original request
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqFailAsn1PackSubReqError
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |        ASN.1 encode fails      |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |  SubDelFail  |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqFailAsn1PackSubReqError(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const subReqCount int = 1

	var params *teststube2ap.RESTSubsReqParams = nil
	params = xappConn1.GetRESTSubsReqReportParams(subReqCount)
	e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_REQ, false)

	// Req
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId)

	// E2t: Receive SubsDelReq
	xappConn1.ExpectRESTNotificationNok(t, restSubId, "allFail")

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", e2SubsId)

	e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_REQ, true)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyUpdateTimeoutAndSubDelOkSameAction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |  Policy modification
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |       RESTNotif(Unsuccessful)  |  E2 timeout
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqPolicyUpdateTimeoutAndSubDelOkSameAction(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubReqTimerExpiry, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const e2Timeout int64 = 1
	const e2RetryCount int64 = 0
	const routingNeeded bool = false

	// Req1
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)

	// Subs Create
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscribe request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	// Policy change
	params = xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	params.SetSubscriptionID(&restSubId)
	params.SetTimeToWait("w200ms")
	restSubId = xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscribe request for subscriberId : %v", restSubId)

	crereq1, cremsg1 = e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	// SubsResp is missing, e2SubsId will be 0
	zeroE2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", zeroE2SubsId)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyAllClean(t)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestPolicyUpdateRESTSubReqAndSubDelOkWithRestartInMiddle
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                                |
//     |           Submgr restart       |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestPolicyUpdateRESTSubReqAndSubDelOkWithRestartInMiddle(t *testing.T) {

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cRestSubReqFromXapp, 2},
		Counter{cRestSubRespToXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cRestSubDelRespToXapp, 1},
	})

	const e2Timeout int64 = 1
	const e2RetryCount int64 = 0
	const routingNeeded bool = false

	// Req1
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	// Create subscription
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscribe request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	// Check subscription
	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// Policy change
	params = xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	params.SetSubscriptionDirectives(e2Timeout, e2RetryCount, routingNeeded)
	params.SetSubscriptionID(&restSubId)
	params.SetTimeToWait("w200ms")
	mainCtrl.SetResetTestFlag(t, true) // subs.DoNotWaitSubResp will be set TRUE for the subscription
	restSubId = xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscribe request for subscriberId : %v", restSubId)

	crereq1, cremsg1 = e2termConn1.RecvSubsReq(t)
	mainCtrl.SetResetTestFlag(t, false)

	// SubsResp is missing due to submgr restart

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// ReadE2Subscriptions() for testing is running in own go routine (go mainCtrl.c.ReadE2Subscriptions())
	// That needs to be completed before successful subscription query is possible
	<-time.After(time.Second * 1)

	// Check subscription
	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// Delete subscription
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
	mainCtrl.VerifyAllClean(t)
}

////////////////////////////////////////////////////////////////////////////////////
//   Services for UT cases
////////////////////////////////////////////////////////////////////////////////////
const subReqCount int = 1
const host string = "localhost"

func createSubscription(t *testing.T, fromXappConn *teststube2ap.E2Stub, toE2termConn *teststube2ap.E2Stub, params *teststube2ap.RESTSubsReqParams) (string, uint32) {
	if params == nil {
		params = fromXappConn.GetRESTSubsReqReportParams(subReqCount)
	}
	restSubId := fromXappConn.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := toE2termConn.RecvSubsReq(t)
	fromXappConn.ExpectRESTNotification(t, restSubId)
	toE2termConn.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := fromXappConn.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func createXapp2MergedSubscription(t *testing.T, meid string) (string, uint32) {

	params := xappConn2.GetRESTSubsReqReportParams(subReqCount)
	if meid != "" {
		params.SetMeid(meid)
	}
	xapp.Subscription.SetResponseCB(xappConn2.SubscriptionRespHandler)
	restSubId := xappConn2.SendRESTSubsReq(t, params)
	xappConn2.ExpectRESTNotification(t, restSubId)
	xapp.Logger.Debug("Send REST subscriber request for subscriberId : %v", restSubId)
	e2SubsId := xappConn2.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func createXapp1PolicySubscription(t *testing.T) (string, uint32) {

	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Debug("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func createXapp1ReportSubscriptionE2Fail(t *testing.T) (string, uint32) {
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Debug("TEST: REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func deleteSubscription(t *testing.T, fromXappConn *teststube2ap.E2Stub, toE2termConn *teststube2ap.E2Stub, restSubId *string) {
	fromXappConn.SendRESTSubsDelReq(t, restSubId)
	delreq, delmsg := toE2termConn.RecvSubsDelReq(t)
	toE2termConn.SendSubsDelResp(t, delreq, delmsg)
}

func deleteXapp1Subscription(t *testing.T, restSubId *string) {
	xappConn1.SendRESTSubsDelReq(t, restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
}

func deleteXapp2Subscription(t *testing.T, restSubId *string) {
	xappConn2.SendRESTSubsDelReq(t, restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
}

func queryXappSubscription(t *testing.T, e2SubsId int64, meid string, endpoint []string) {
	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, e2SubsId, resp[0].SubscriptionID)
	assert.Equal(t, meid, resp[0].Meid)
	assert.Equal(t, endpoint, resp[0].ClientEndpoint)
}

func waitSubsCleanup(t *testing.T, e2SubsId uint32, timeout int) {
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, timeout)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, timeout)
}

func sendAndReceiveMultipleE2SubReqs(t *testing.T, count int, fromXappConn *teststube2ap.E2Stub, toE2termConn *teststube2ap.E2Stub, restSubId string) []uint32 {

	var e2SubsId []uint32

	for i := 0; i < count; i++ {
		xapp.Logger.Debug("TEST: %d ===================================== BEGIN CRE ============================================", i+1)
		crereq, cremsg := toE2termConn.RecvSubsReq(t)
		fromXappConn.ExpectRESTNotification(t, restSubId)
		toE2termConn.SendSubsResp(t, crereq, cremsg)
		instanceId := fromXappConn.WaitRESTNotification(t, restSubId)
		e2SubsId = append(e2SubsId, instanceId)
		xapp.Logger.Debug("TEST: %v", e2SubsId)
		xapp.Logger.Debug("TEST: %d ===================================== END CRE ============================================", i+1)
		<-time.After(100 * time.Millisecond)
	}
	return e2SubsId
}

func sendAndReceiveMultipleE2DelReqs(t *testing.T, e2SubsIds []uint32, toE2termConn *teststube2ap.E2Stub) {

	for i := 0; i < len(e2SubsIds); i++ {
		xapp.Logger.Debug("TEST: %d ===================================== BEGIN DEL ============================================", i+1)
		delreq, delmsg := toE2termConn.RecvSubsDelReq(t)
		toE2termConn.SendSubsDelResp(t, delreq, delmsg)
		<-time.After(1 * time.Second)
		xapp.Logger.Debug("TEST: %d ===================================== END DEL ============================================", i+1)
		<-time.After(100 * time.Millisecond)
	}

	// Wait that subs is cleaned
	for i := 0; i < len(e2SubsIds); i++ {
		mainCtrl.wait_subs_clean(t, e2SubsIds[i], 10)
	}

}
