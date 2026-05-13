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

package teststube2ap

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	clientmodel "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/clientmodel"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrTransactionId struct {
	xid  string
	meid *xapp.RMRMeid
}

type E2RestIds struct {
	RestSubsId  string
	E2SubsId    uint32
	ErrorCause  string
	ErrorSource string
	TimeoutType string
}

func (trans *RmrTransactionId) String() string {
	meidstr := "N/A"
	if trans.meid != nil {
		meidstr = trans.meid.String()
	}
	return "trans(" + trans.xid + "/" + meidstr + ")"
}

type E2Stub struct {
	teststub.RmrStubControl
	xid_seq                     uint64
	subscriptionId              string
	requestCount                int
	CallBackNotification        chan int64
	RESTNotification            chan uint32
	CallBackListedNotifications chan E2RestIds
	ListedRESTNotifications     chan E2RestIds
	clientEndpoint              clientmodel.SubscriptionParamsClientEndpoint
	meid                        string
	restSubsIdList              []string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewE2Stub(desc string, srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc, stat string, mtypeseed int, ranName string, host string, RMRPort int64, HTTPPort int64) *E2Stub {
	tc := &E2Stub{}
	tc.RmrStubControl.Init(desc, srcId, rtgSvc, stat, mtypeseed)
	tc.xid_seq = 1
	tc.SetCheckXid(true)
	tc.CallBackNotification = make(chan int64)
	tc.RESTNotification = make(chan uint32)
	tc.CallBackListedNotifications = make(chan E2RestIds)
	tc.ListedRESTNotifications = make(chan E2RestIds, 2)
	var endPoint clientmodel.SubscriptionParamsClientEndpoint
	endPoint.Host = host
	endPoint.HTTPPort = &HTTPPort
	endPoint.RMRPort = &RMRPort
	tc.clientEndpoint = endPoint
	tc.meid = ranName
	return tc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewE2termStub(desc string, srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc, stat string, mtypeseed int) *E2Stub {
	tc := &E2Stub{}
	tc.RmrStubControl.Init(desc, srcId, rtgSvc, stat, mtypeseed)
	tc.xid_seq = 1
	tc.SetCheckXid(false)
	return tc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) NewRmrTransactionId(xid string, ranname string) *RmrTransactionId {
	trans := &RmrTransactionId{}
	if len(xid) == 0 {
		trans.xid = tc.GetDesc() + "_XID_" + strconv.FormatUint(uint64(tc.xid_seq), 10)
		tc.xid_seq++
	} else {
		trans.xid = xid
	}
	trans.meid = &xapp.RMRMeid{RanName: ranname}
	tc.Debug("New test %s", trans.String())
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2StubSubsReqParams struct {
	Req *e2ap.E2APSubscriptionRequest
}

func (p *E2StubSubsReqParams) Init() {
	p.Req = &e2ap.E2APSubscriptionRequest{}

	p.Req.RequestId.Id = 1
	p.Req.RequestId.InstanceId = 0
	p.Req.FunctionId = 1

	// gnb -> enb outgoing
	// enb -> gnb incoming
	// X2 36423-f40.doc
	p.Req.EventTriggerDefinition.Data.Length = 1
	p.Req.EventTriggerDefinition.Data.Data = make([]uint8, p.Req.EventTriggerDefinition.Data.Length)
	p.Req.EventTriggerDefinition.Data.Data[0] = 1

	p.Req.ActionSetups = make([]e2ap.ActionToBeSetupItem, 1)

	p.Req.ActionSetups[0].ActionId = 0
	p.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeReport
	p.Req.ActionSetups[0].RicActionDefinitionPresent = true

	p.Req.ActionSetups[0].ActionDefinitionChoice.Data.Length = 1
	p.Req.ActionSetups[0].ActionDefinitionChoice.Data.Data = make([]uint8, p.Req.ActionSetups[0].ActionDefinitionChoice.Data.Length)
	p.Req.ActionSetups[0].ActionDefinitionChoice.Data.Data[0] = 1

	p.Req.ActionSetups[0].SubsequentAction.Present = true
	p.Req.ActionSetups[0].SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
	p.Req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitZero
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type E2StubSubsFailParams struct {
	Req  *e2ap.E2APSubscriptionRequest
	Fail *e2ap.E2APSubscriptionFailure
}

func (p *E2StubSubsFailParams) Set(req *e2ap.E2APSubscriptionRequest) {
	p.Req = req

	p.Fail = &e2ap.E2APSubscriptionFailure{}
	p.Fail.RequestId.Id = p.Req.RequestId.Id
	p.Fail.RequestId.InstanceId = p.Req.RequestId.InstanceId
	p.Fail.FunctionId = p.Req.FunctionId
	p.Fail.Cause.Content = e2ap.E2AP_CauseContent_RICrequest
	p.Fail.Cause.Value = e2ap.E2AP_CauseValue_RICrequest_control_message_invalid
}

func (p *E2StubSubsFailParams) SetCauseVal(ind int, content uint8, causeval uint8) {

	p.Fail.Cause.Content = content
	p.Fail.Cause.Value = causeval
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (tc *E2Stub) SendSubsReq(t *testing.T, rparams *E2StubSubsReqParams, oldTrans *RmrTransactionId) *RmrTransactionId {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewRmrTransactionId("", "RAN_NAME_1")
	}

	tc.Debug("SendSubsReq %s", trans.String())
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// xapp activity: Send Subs Req
	//---------------------------------
	myparams := rparams

	if myparams == nil {
		myparams = &E2StubSubsReqParams{}
		myparams.Init()
	}

	err, packedMsg := e2SubsReq.Pack(myparams.Req)
	if err != nil {
		tc.TestError(t, "pack NOK %s %s", trans.String(), err.Error())
		return nil
	}
	tc.Debug("%s %s", trans.String(), e2SubsReq.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_REQ
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	tc.Debug("SEND SUB REQ: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s %s", trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsReq(t *testing.T) (*e2ap.E2APSubscriptionRequest, *xapp.RMRParams) {
	tc.Debug("RecvSubsReq")
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// e2term activity: Recv Subs Req
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_REQ"] {
			tc.TestError(t, "Received wrong mtype expected %s got %s, error", "RIC_SUB_REQ", xapp.RicMessageTypeToName[msg.Mtype])
		} else {
			tc.Debug("Recv Subs Req")
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, req := e2SubsReq.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_REQ unpack failed err: %s", unpackerr.Error())
			}
			return req, msg
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}

	return nil, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsResp(t *testing.T, req *e2ap.E2APSubscriptionRequest, msg *xapp.RMRParams) {
	tc.Debug("SendSubsResp")
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()

	//---------------------------------
	// e2term activity: Send Subs Resp
	//---------------------------------
	resp := &e2ap.E2APSubscriptionResponse{}

	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId

	resp.ActionAdmittedList.Items = make([]e2ap.ActionAdmittedItem, len(req.ActionSetups))
	for index := int(0); index < len(req.ActionSetups); index++ {
		resp.ActionAdmittedList.Items[index].ActionId = req.ActionSetups[index].ActionId
	}

	packerr, packedMsg := e2SubsResp.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Debug("%s", e2SubsResp.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_RESP
	//params.SubId = msg.SubId
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	//params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Debug("SEND SUB RESP: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendPartialSubsResp(t *testing.T, req *e2ap.E2APSubscriptionRequest, msg *xapp.RMRParams, actionNotAdmittedList e2ap.ActionNotAdmittedList) {
	tc.Debug("SendPartialSubsResp")

	if len(actionNotAdmittedList.Items) == 0 {
		tc.TestError(t, "SendPartialSubsResp() Empty actionNotAdmittedList.Items")
		return
	}

	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()

	//---------------------------------
	// e2term activity: Send Subs Resp
	//---------------------------------
	resp := &e2ap.E2APSubscriptionResponse{}

	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId

	for index, actionNotAdmittedItem := range actionNotAdmittedList.Items {
		for _, ActionToBeSetupItem := range req.ActionSetups {
			if ActionToBeSetupItem.ActionId == actionNotAdmittedItem.ActionId {
				actionNotAdmittedItem := e2ap.ActionNotAdmittedItem{}
				actionNotAdmittedItem.ActionId = ActionToBeSetupItem.ActionId
				actionNotAdmittedItem.Cause.Content = 1
				actionNotAdmittedItem.Cause.Value = 8
				resp.ActionNotAdmittedList.Items = append(resp.ActionNotAdmittedList.Items, actionNotAdmittedItem)
				// Remove the element
				req.ActionSetups = append(req.ActionSetups[:index], req.ActionSetups[index+1:]...)

			}
		}
	}
	for _, ActionToBeSetupItem := range req.ActionSetups {
		actionAdmittedItem := e2ap.ActionAdmittedItem{}
		actionAdmittedItem.ActionId = ActionToBeSetupItem.ActionId
		resp.ActionAdmittedList.Items = append(resp.ActionAdmittedList.Items, actionAdmittedItem)
	}

	packerr, packedMsg := e2SubsResp.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Debug("%s", e2SubsResp.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_RESP
	//params.SubId = msg.SubId
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	//params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Debug("SEND SUB RESP: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendInvalidE2Asn1Resp(t *testing.T, msg *xapp.RMRParams, msgType int) {

	params := &xapp.RMRParams{}
	params.Mtype = msgType
	params.SubId = -1
	params.Payload = []byte{1, 2, 3, 4, 5}
	params.PayloadLen = 5
	params.Meid = msg.Meid
	params.Xid = ""
	params.Mbuf = nil

	if params.Mtype == xapp.RIC_SUB_RESP {
		tc.Debug("SEND INVALID ASN.1 SUB RESP")

	} else if params.Mtype == xapp.RIC_SUB_FAILURE {
		tc.Debug("SEND INVALID ASN.1 SUB FAILURE")

	} else if params.Mtype == xapp.RIC_SUB_DEL_RESP {
		tc.Debug("SEND INVALID ASN.1 SUB DEL RESP")

	} else if params.Mtype == xapp.RIC_SUB_DEL_FAILURE {
		tc.Debug("SEND INVALID ASN.1 SUB DEL FAILURE")
	}
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsResp(t *testing.T, trans *RmrTransactionId) uint32 {
	tc.Debug("RecvSubsResp")
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()
	var e2SubsId uint32

	//---------------------------------
	// xapp activity: Recv Subs Resp
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_RESP"] {
			tc.TestError(t, "Received RIC_SUB_RESP wrong mtype expected %s got %s, error", "RIC_SUB_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return 0
		} else if msg.Xid != trans.xid {
			tc.TestError(t, "Received RIC_SUB_RESP wrong xid expected %s got %s, error", trans.xid, msg.Xid)
			return 0
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			if msg.SubId > 0 {
				e2SubsId = uint32(msg.SubId)
			} else {
				e2SubsId = 0
			}
			unpackerr, resp := e2SubsResp.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_RESP unpack failed err: %s", unpackerr.Error())
			}
			tc.Debug("Recv Subs Resp rmr: xid=%s subid=%d, asn: instanceid=%d", msg.Xid, msg.SubId, resp.RequestId.InstanceId)
			return e2SubsId
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return 0
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (tc *E2Stub) SendSubsFail(t *testing.T, fparams *E2StubSubsFailParams, msg *xapp.RMRParams) {
	tc.Debug("SendSubsFail")
	e2SubsFail := e2asnpacker.NewPackerSubscriptionFailure()

	//---------------------------------
	// e2term activity: Send Subs Fail
	//---------------------------------
	packerr, packedMsg := e2SubsFail.Pack(fparams.Fail)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Debug("%s", e2SubsFail.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_FAILURE
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Debug("SEND SUB FAIL: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsFail(t *testing.T, trans *RmrTransactionId) uint32 {
	tc.Debug("RecvSubsFail")
	e2SubsFail := e2asnpacker.NewPackerSubscriptionFailure()
	var e2SubsId uint32

	//-------------------------------
	// xapp activity: Recv Subs Fail
	//-------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_FAILURE"] {
			tc.TestError(t, "Received RIC_SUB_FAILURE wrong mtype expected %s got %s, error", "RIC_SUB_FAILURE", xapp.RicMessageTypeToName[msg.Mtype])
			return 0
		} else if msg.Xid != trans.xid {
			tc.TestError(t, "Received RIC_SUB_FAILURE wrong xid expected %s got %s, error", trans.xid, msg.Xid)
			return 0
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			if msg.SubId > 0 {
				e2SubsId = uint32(msg.SubId)
			} else {
				e2SubsId = 0
			}
			unpackerr, resp := e2SubsFail.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_FAILURE unpack failed err: %s", unpackerr.Error())
			}
			tc.Debug("Recv Subs Fail rmr: xid=%s subid=%d, asn: instanceid=%d", msg.Xid, msg.SubId, resp.RequestId.InstanceId)
			return e2SubsId
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return 0
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsDelReq(t *testing.T, oldTrans *RmrTransactionId, e2SubsId uint32) *RmrTransactionId {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewRmrTransactionId("", "RAN_NAME_1")
	}

	tc.Debug("SendSubsDelReq %s", trans.String())
	e2SubsDelReq := e2asnpacker.NewPackerSubscriptionDeleteRequest()
	//---------------------------------
	// xapp activity: Send Subs Del Req
	//---------------------------------
	req := &e2ap.E2APSubscriptionDeleteRequest{}
	req.RequestId.Id = 1
	req.RequestId.InstanceId = e2SubsId
	req.FunctionId = 1

	err, packedMsg := e2SubsDelReq.Pack(req)
	if err != nil {
		tc.TestError(t, "pack NOK %s %s", trans.String(), err.Error())
		return nil
	}
	tc.Debug("%s %s", trans.String(), e2SubsDelReq.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_DEL_REQ
	params.SubId = int(e2SubsId)
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	tc.Debug("SEND SUB DEL REQ: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s %s", trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsDelReq(t *testing.T) (*e2ap.E2APSubscriptionDeleteRequest, *xapp.RMRParams) {
	tc.Debug("RecvSubsDelReq")
	e2SubsDelReq := e2asnpacker.NewPackerSubscriptionDeleteRequest()

	//---------------------------------
	// e2term activity: Recv Subs Del Req
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_REQ"] {
			tc.TestError(t, "Received wrong mtype expected %s got %s, error", "RIC_SUB_DEL_REQ", xapp.RicMessageTypeToName[msg.Mtype])
		} else {
			tc.Debug("Recv Subs Del Req")

			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, req := e2SubsDelReq.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_DEL_REQ unpack failed err: %s", unpackerr.Error())
			}
			return req, msg
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return nil, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsDelResp(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapp.RMRParams) {
	tc.Debug("SendSubsDelResp")
	e2SubsDelResp := e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// e2term activity: Send Subs Del Resp
	//---------------------------------
	resp := &e2ap.E2APSubscriptionDeleteResponse{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId

	packerr, packedMsg := e2SubsDelResp.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Debug("%s", e2SubsDelResp.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_DEL_RESP
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Debug("SEND SUB DEL RESP: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsDelResp(t *testing.T, trans *RmrTransactionId) {
	tc.Debug("RecvSubsDelResp")
	e2SubsDelResp := e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// xapp activity: Recv Subs Del Resp
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_RESP"] {
			tc.TestError(t, "Received RIC_SUB_DEL_RESP wrong mtype expected %s got %s, error", "RIC_SUB_DEL_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return
		} else if trans != nil && msg.Xid != trans.xid {
			tc.TestError(t, "Received RIC_SUB_DEL_RESP wrong xid expected %s got %s, error", trans.xid, msg.Xid)
			return
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, resp := e2SubsDelResp.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_DEL_RESP unpack failed err: %s", unpackerr.Error())
			}
			tc.Debug("Recv Subs Del Resp rmr: xid=%s subid=%d, asn: instanceid=%d", msg.Xid, msg.SubId, resp.RequestId.InstanceId)
			return
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsDelFail(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapp.RMRParams) {
	tc.Debug("SendSubsDelFail")
	e2SubsDelFail := e2asnpacker.NewPackerSubscriptionDeleteFailure()

	//---------------------------------
	// e2term activity: Send Subs Del Fail
	//---------------------------------
	resp := &e2ap.E2APSubscriptionDeleteFailure{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId
	resp.Cause.Content = 5 // CauseMisc
	resp.Cause.Value = 3   // unspecified

	packerr, packedMsg := e2SubsDelFail.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Debug("%s", e2SubsDelFail.String())

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_DEL_FAILURE
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Debug("SEND SUB DEL FAIL: %s", params.String())
	snderr := tc.SendWithRetry(params, false, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

// REST code below all

/*****************************************************************************/
// REST interface specific functions are below

//-----------------------------------------------------------------------------
// Callback handler for subscription response notifications
//-----------------------------------------------------------------------------
func (tc *E2Stub) SubscriptionRespHandler(resp *clientmodel.SubscriptionResponse) {

	if tc.subscriptionId == "SUBSCRIPTIONID NOT SET" {
		tc.Debug("REST notification received for %v while no SubscriptionID was not set for E2EventInstanceID=%v, XappEventInstanceID=%v",
			*resp.SubscriptionID, *resp.SubscriptionInstances[0].E2EventInstanceID, *resp.SubscriptionInstances[0].XappEventInstanceID)
		tc.CallBackNotification <- *resp.SubscriptionInstances[0].E2EventInstanceID
	} else if tc.subscriptionId == *resp.SubscriptionID {
		tc.Debug("REST notification received SubscriptionID=%s, E2EventInstanceID=%v, RequestorID=%v",
			*resp.SubscriptionID, *resp.SubscriptionInstances[0].E2EventInstanceID, *resp.SubscriptionInstances[0].XappEventInstanceID)
		tc.CallBackNotification <- *resp.SubscriptionInstances[0].E2EventInstanceID
	} else {
		tc.Debug("MISMATCHING REST notification received SubscriptionID=%s<>%s, E2EventInstanceID=%v, XappEventInstanceID=%v",
			tc.subscriptionId, *resp.SubscriptionID, *resp.SubscriptionInstances[0].E2EventInstanceID, *resp.SubscriptionInstances[0].XappEventInstanceID)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (tc *E2Stub) ExpectRESTNotification(t *testing.T, restSubsId string) {
	tc.expectNotification(t, restSubsId, "")
}

func (tc *E2Stub) ExpectRESTNotificationOk(t *testing.T, restSubsId string) {
	tc.expectNotification(t, restSubsId, "allOk")
}

func (tc *E2Stub) ExpectRESTNotificationNok(t *testing.T, restSubsId string, expectError string) {
	tc.expectNotification(t, restSubsId, expectError)
}

func (tc *E2Stub) expectNotification(t *testing.T, restSubsId string, expectError string) {

	tc.Debug("### Started to wait REST notification for restSubsId= %v, RMRPort=%v, requestCount=%v responses expected", restSubsId, *tc.clientEndpoint.RMRPort, tc.requestCount)
	tc.restSubsIdList = []string{restSubsId}
	xapp.Subscription.SetResponseCB(tc.ListedRestNotifHandler)
	if tc.requestCount == 0 {
		tc.TestError(t, "### NO REST notifications SET received for restSubsId= %v, RMRPort=%v, request count ZERO!", restSubsId, *tc.clientEndpoint.RMRPort)
	}
	go func() {
		select {
		case e2Ids := <-tc.CallBackListedNotifications:
			if tc.requestCount == 0 {
				tc.TestError(t, "### REST notification count unexpectedly ZERO for restSubsId= %v", restSubsId)
			} else if e2Ids.RestSubsId != restSubsId {
				tc.TestError(t, "### Unexpected REST notifications received, expected %s but got %s instead", e2Ids.RestSubsId, restSubsId)
			} else if e2Ids.ErrorCause == "" && expectError == "allFail" {
				tc.TestError(t, "### Unexpected ok cause received from REST notifications |%s:%s|", e2Ids.RestSubsId, restSubsId)
			} else if e2Ids.ErrorCause != "" && expectError == "allOk" {
				tc.TestError(t, "### Unexpected ErrorCause: (%s), ErrorSource: (%s), TimeoutType: (%s) received from REST notifications |%s:%s|", e2Ids.ErrorCause, e2Ids.ErrorSource, e2Ids.TimeoutType, e2Ids.RestSubsId, restSubsId)
			} else {
				tc.requestCount--
				if tc.requestCount == 0 {
					tc.Debug("### All expected REST notifications received for restSubsId= %v)", e2Ids.RestSubsId)
				} else {
					tc.Debug("### Expected REST notifications received for restSubsId= %v", e2Ids.RestSubsId)
				}
				if e2Ids.ErrorCause != "" && expectError == "allFail" {
					tc.Debug("### REST Notification: RestSubsId: %s, ErrorCause: %s, ErrorSource: (%s), TimeoutType: (%s)", e2Ids.RestSubsId, e2Ids.ErrorCause, e2Ids.ErrorSource, e2Ids.TimeoutType)
				}
				tc.Debug("### REST Notification received Notif for %s : %v", e2Ids.RestSubsId, e2Ids.E2SubsId)
				tc.ListedRESTNotifications <- e2Ids
				if len(tc.ListedRESTNotifications) > 1 {
					panic("expectNotification - ListedRESTNotifications stacking up")
				}
			}
		case <-time.After(15 * time.Second):
			err := fmt.Errorf("### Timeout 15s expired while expecting REST notification for subsId: %v", restSubsId)
			tc.TestError(t, "%s", err.Error())
		}
	}()
}

func (tc *E2Stub) WaitRESTNotification(t *testing.T, restSubsId string) uint32 {

	stack := string(debug.Stack())

	select {
	case e2SubsId := <-tc.ListedRESTNotifications:
		if e2SubsId.RestSubsId == restSubsId {
			tc.Debug("### Expected REST notifications received %s, e2SubsId %v", e2SubsId.RestSubsId, e2SubsId.E2SubsId)
			return e2SubsId.E2SubsId
		} else {
			tc.TestError(t, "### Unexpected REST notification %s received, expected %v", e2SubsId.RestSubsId, restSubsId)
			xapp.Logger.Debug("CALL STACK:\n %s", stack)
			return 0
		}
	case <-time.After(15 * time.Second):
		err := fmt.Errorf("### Timeout 15s expired while waiting REST notification for subsId: %v", restSubsId)
		tc.TestError(t, "%s", err.Error())
		xapp.Logger.Debug("CALL STACK:\n %s", stack)
		panic("WaitRESTNotification - timeout error")
	}
}

// Note, this function should be followed by a handling of <-xappConn1.RESTNotification.
func (tc *E2Stub) ExpectAnyNotification(t *testing.T) {
	go func() {
		tc.Debug("### Started waiting ANY REST notifications received for RMRPort=%v", *tc.clientEndpoint.RMRPort)
		select {
		case e2SubsId := <-tc.CallBackNotification:
			tc.Debug("### ANY REST notifications received for e2SubsId %v RMRPort=%v", e2SubsId, *tc.clientEndpoint.RMRPort)
			tc.RESTNotification <- (uint32)(e2SubsId)
		case <-time.After(15 * time.Second):
			err := fmt.Errorf("### Timeout 15s expired while waiting ANY REST notification")
			tc.TestError(t, "%s", err.Error())
			tc.RESTNotification <- 0
		}
	}()
}

func (tc *E2Stub) WaitAnyRESTNotification(t *testing.T) uint32 {

	select {
	case e2SubsId := <-tc.RESTNotification:
		tc.Debug("### Expected ANY REST notification received for e2SubsId %v", e2SubsId)
		return e2SubsId

	case <-time.After(15 * time.Second):
		err := fmt.Errorf("### Timeout 15s expired while waiting ANY REST notification")
		tc.TestError(t, "%s", err.Error())
		panic("WaitRESTNotification - timeout error")
	}
}

func (tc *E2Stub) ListedRestNotifHandler(resp *clientmodel.SubscriptionResponse) {

	if len(tc.restSubsIdList) == 0 {
		tc.Error("Unexpected listed REST notifications received for SubscriptionID=%s, expected restSubsId list size was ZERO!", *resp.SubscriptionID)
	} else {
		for i, subsId := range tc.restSubsIdList {
			if *resp.SubscriptionID == subsId {
				//tc.Debug("Listed REST notifications received SubscriptionID=%s, InstanceID=%v, XappEventInstanceID=%v",
				//	*resp.SubscriptionID, *resp.SubscriptionInstances[0].InstanceID, *resp.SubscriptionInstances[0].XappEventInstanceID)

				tc.restSubsIdList = append(tc.restSubsIdList[:i], tc.restSubsIdList[i+1:]...)
				//tc.Debug("Removed %s from Listed REST notifications, %v entries left", *resp.SubscriptionID, len(tc.restSubsIdList))

				//				if resp.SubscriptionInstances[0].ErrorCause != nil {
				tc.CallBackListedNotifications <- *tc.GetE2RestIds(resp)
				//				} else {
				//					tc.CallBackListedNotifications <- E2RestIds{*resp.SubscriptionID, uint32(*resp.SubscriptionInstances[0].E2EventInstanceID), "", "", ""}
				//				}

				if len(tc.restSubsIdList) == 0 {
					tc.Debug("All listed REST notifications received for SubscriptionID=%s", *resp.SubscriptionID)
				}

				return
			}
		}
		tc.Error("UNKONWN REST notification received SubscriptionID=%s<>%s, InstanceID=%v, XappEventInstanceID=%v",
			tc.subscriptionId, *resp.SubscriptionID, *resp.SubscriptionInstances[0].E2EventInstanceID, *resp.SubscriptionInstances[0].XappEventInstanceID)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) GetE2RestIds(resp *clientmodel.SubscriptionResponse) *E2RestIds {

	e2RestIds := &E2RestIds{}
	if resp != nil {
		if resp.SubscriptionID != nil {
			e2RestIds.RestSubsId = *resp.SubscriptionID
		}
		if resp.SubscriptionInstances != nil {
			if resp.SubscriptionInstances[0].E2EventInstanceID != nil {
				e2RestIds.E2SubsId = uint32(*resp.SubscriptionInstances[0].E2EventInstanceID)
			}
			e2RestIds.ErrorCause = resp.SubscriptionInstances[0].ErrorCause
			e2RestIds.ErrorSource = resp.SubscriptionInstances[0].ErrorSource
			e2RestIds.TimeoutType = resp.SubscriptionInstances[0].TimeoutType
		}
	}
	return e2RestIds
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) WaitListedRestNotifications(t *testing.T, restSubsIds []string) {
	tc.Debug("Started to wait REST notifications for restSubsIds=%v, RMRPort=%v", restSubsIds, *tc.clientEndpoint.RMRPort)

	tc.restSubsIdList = restSubsIds
	xapp.Subscription.SetResponseCB(tc.ListedRestNotifHandler)

	for i := 0; i < len(restSubsIds); i++ {
		go func() {
			select {
			case e2Ids := <-tc.CallBackListedNotifications:
				tc.Debug("Listed Notification waiter received Notif for %s : %v", e2Ids.RestSubsId, e2Ids.E2SubsId)
				tc.ListedRESTNotifications <- e2Ids
			case <-time.After(15 * time.Second):
				err := fmt.Errorf("Timeout 15s expired while waiting Listed REST notification")
				tc.TestError(t, "%s", err.Error())
				tc.RESTNotification <- 0
			}
		}()
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendRESTSubsReq(t *testing.T, params *RESTSubsReqParams) string {
	tc.Debug("======== Posting REST subscriptions to Submgr ======")

	if params == nil {
		tc.Debug("SendRESTReportSubsReq: params == nil")
		return ""
	}

	tc.subscriptionId = "SUBSCRIPTIONID NOT SET"

	resp, err := xapp.Subscription.Subscribe(&params.SubsReqParams)
	if err != nil {
		// Swagger generated code makes checks for the values that are inserted the subscription function
		// If error cause is unknown and POST is not done, the problem is in the inserted values
		tc.Error("======== REST subscription request failed %s ========", err.Error())
		return ""
	}
	tc.subscriptionId = *resp.SubscriptionID
	tc.Debug("======== REST subscriptions posted successfully. SubscriptionID = %s, RequestCount = %v ========", *resp.SubscriptionID, tc.requestCount)
	return *resp.SubscriptionID
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) GetRESTSubsReqReportParams(subReqCount int) *RESTSubsReqParams {

	reportParams := RESTSubsReqParams{}
	reportParams.GetRESTSubsReqReportParams(subReqCount, &tc.clientEndpoint, &tc.meid)
	tc.requestCount = subReqCount
	return &reportParams
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RESTSubsReqParams struct {
	SubsReqParams clientmodel.SubscriptionParams
}

func (p *RESTSubsReqParams) GetRESTSubsReqReportParams(subReqCount int, clientEndpoint *clientmodel.SubscriptionParamsClientEndpoint, meid *string) {

	// E2SM-gNB-X2
	p.SubsReqParams.ClientEndpoint = clientEndpoint
	p.SubsReqParams.Meid = meid
	var rANFunctionID int64 = 33
	p.SubsReqParams.RANFunctionID = &rANFunctionID

	actionId := int64(1)
	actionType := "report"
	subsequestActioType := "continue"
	timeToWait := "w10ms"

	for requestCount := 0; requestCount < subReqCount; requestCount++ {
		reqId := int64(requestCount) + 1
		subscriptionDetail := &clientmodel.SubscriptionDetail{
			XappEventInstanceID: &reqId,
			EventTriggers: clientmodel.EventTriggerDefinition{
				int64(1234 + requestCount),
			},
			ActionToBeSetupList: clientmodel.ActionsToBeSetup{
				&clientmodel.ActionToBeSetup{
					ActionID:   &actionId,
					ActionType: &actionType,
					ActionDefinition: clientmodel.ActionDefinition{
						int64(5678 + requestCount),
					},
					SubsequentAction: &clientmodel.SubsequentAction{
						SubsequentActionType: &subsequestActioType,
						TimeToWait:           &timeToWait,
					},
				},
			},
		}
		p.SubsReqParams.SubscriptionDetails = append(p.SubsReqParams.SubscriptionDetails, subscriptionDetail)
	}

}

func (p *RESTSubsReqParams) SetMeid(MEID string) {
	p.SubsReqParams.Meid = &MEID
}

func (p *RESTSubsReqParams) SetEndpoint(HTTP_port int64, RMR_port int64, host string) {
	var endpoint clientmodel.SubscriptionParamsClientEndpoint
	endpoint.HTTPPort = &HTTP_port
	endpoint.RMRPort = &RMR_port
	endpoint.Host = host
	p.SubsReqParams.ClientEndpoint = &endpoint
}

func (p *RESTSubsReqParams) SetEndpointHost(host string) {

	if p.SubsReqParams.ClientEndpoint.Host != "" {
		if p.SubsReqParams.ClientEndpoint.Host != host {
			// Renaming toto, print something tc.Debug("Posting REST subscription request to Submgr")
			err := fmt.Errorf("hostname change attempt: %s -> %s", p.SubsReqParams.ClientEndpoint.Host, host)
			panic(err)
		}
	}
	p.SubsReqParams.ClientEndpoint.Host = host
}

func (p *RESTSubsReqParams) SetHTTPEndpoint(HTTP_port int64, host string) {

	p.SubsReqParams.ClientEndpoint.HTTPPort = &HTTP_port

	p.SetEndpointHost(host)

	if p.SubsReqParams.ClientEndpoint.RMRPort == nil {
		var RMR_port int64
		p.SubsReqParams.ClientEndpoint.RMRPort = &RMR_port
	}
}

func (p *RESTSubsReqParams) SetE2SubscriptionDirectives(E2RetryCount int64, E2TimeoutTimerValue int64, RMRRoutingNeeded bool) {

	E2SubscriptionDirectives := &clientmodel.SubscriptionParamsE2SubscriptionDirectives{}
	p.SubsReqParams.E2SubscriptionDirectives = E2SubscriptionDirectives
	p.SubsReqParams.E2SubscriptionDirectives.E2RetryCount = &E2RetryCount
	p.SubsReqParams.E2SubscriptionDirectives.E2TimeoutTimerValue = E2TimeoutTimerValue
	p.SubsReqParams.E2SubscriptionDirectives.RMRRoutingNeeded = RMRRoutingNeeded
}

func (p *RESTSubsReqParams) RemoveE2SubscriptionDirectives() {

	p.SubsReqParams.E2SubscriptionDirectives = nil
}

func (p *RESTSubsReqParams) SetSubActionTypes(actionType string) {

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		for _, action := range subDetail.ActionToBeSetupList {
			if action != nil {
				action.ActionType = &actionType
			}
		}
	}
}

func (p *RESTSubsReqParams) SetSubActionIDs(actionId int64) {

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		for _, action := range subDetail.ActionToBeSetupList {
			if action != nil {
				action.ActionID = &actionId
			}
		}
	}
}

func (p *RESTSubsReqParams) SetSubActionDefinition(actionDefinition []int64) {

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		for _, action := range subDetail.ActionToBeSetupList {
			if action != nil {
				action.ActionDefinition = actionDefinition
			}
		}
	}
}

func (p *RESTSubsReqParams) SetSubEventTriggerDefinition(eventTriggerDefinition []int64) {

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		if subDetail != nil {
			subDetail.EventTriggers = eventTriggerDefinition
		}
	}
}

func (p *RESTSubsReqParams) AppendActionToActionToBeSetupList(actionId int64, actionType string, actionDefinition []int64, subsequentActionType string, timeToWait string) {

	actionToBeSetup := &clientmodel.ActionToBeSetup{
		ActionID:         &actionId,
		ActionType:       &actionType,
		ActionDefinition: clientmodel.ActionDefinition(actionDefinition),
		SubsequentAction: &clientmodel.SubsequentAction{
			SubsequentActionType: &subsequentActionType,
			TimeToWait:           &timeToWait,
		},
	}

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		if subDetail != nil {
			subDetail.ActionToBeSetupList = append(subDetail.ActionToBeSetupList, actionToBeSetup)
		}
	}
}

func (p *RESTSubsReqParams) SetRMREndpoint(RMR_port int64, host string) {

	p.SubsReqParams.ClientEndpoint.RMRPort = &RMR_port

	p.SetEndpointHost(host)

	if p.SubsReqParams.ClientEndpoint.HTTPPort == nil {
		var HTTP_port int64
		p.SubsReqParams.ClientEndpoint.HTTPPort = &HTTP_port
	}
}

func (p *RESTSubsReqParams) SetTimeToWait(timeToWait string) {

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		for _, action := range subDetail.ActionToBeSetupList {
			if action != nil && action.SubsequentAction != nil {
				action.SubsequentAction.TimeToWait = &timeToWait
			}
		}
	}
}

func (p *RESTSubsReqParams) SetSubscriptionID(SubscriptionID *string) {

	if *SubscriptionID == "" {
		return
	}
	p.SubsReqParams.SubscriptionID = *SubscriptionID
}

func (p *RESTSubsReqParams) SetXappEventInstanceID(xappEventInstanceId int64) {

	for _, subDetail := range p.SubsReqParams.SubscriptionDetails {
		if subDetail != nil {
			subDetail.XappEventInstanceID = &xappEventInstanceId
		}
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendRESTSubsDelReq(t *testing.T, subscriptionID *string) {

	tc.Debug("======== Posting REST DELETE subscription(s) to Submgr ======")

	if *subscriptionID == "" {
		tc.Error("REST error in deleting subscriptions. Empty SubscriptionID = %s", *subscriptionID)
	}
	tc.Debug("REST deleting E2 subscriptions. SubscriptionID = %s", *subscriptionID)

	err := xapp.Subscription.Unsubscribe(*subscriptionID)
	if err != nil {
		tc.Error("REST Delete subscription failed %s", err.Error())
	}
	tc.Debug("REST delete subscription pushed to submgr successfully. SubscriptionID = %s", *subscriptionID)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) GetRESTSubsReqPolicyParams(subReqCount int) *RESTSubsReqParams {

	policyParams := RESTSubsReqParams{}
	policyParams.GetRESTSubsReqPolicyParams(subReqCount, &tc.clientEndpoint, &tc.meid)
	tc.requestCount = subReqCount
	return &policyParams
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) DecrementRequestCount() {
	if tc.requestCount > 0 {
		tc.requestCount -= 1
	} else {
		tc.Error("FAILED to decrement request count, count already ZERO")
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (p *RESTSubsReqParams) GetRESTSubsReqPolicyParams(subReqCount int, clientEndpoint *clientmodel.SubscriptionParamsClientEndpoint, meid *string) {

	p.SubsReqParams.ClientEndpoint = clientEndpoint
	p.SubsReqParams.Meid = meid
	var rANFunctionID int64 = 33
	p.SubsReqParams.RANFunctionID = &rANFunctionID
	actionId := int64(1)
	actionType := "policy"
	subsequestActioType := "continue"
	timeToWait := "w10ms"

	for requestCount := 0; requestCount < subReqCount; requestCount++ {
		reqId := int64(requestCount) + 1
		subscriptionDetail := &clientmodel.SubscriptionDetail{
			XappEventInstanceID: &reqId,
			EventTriggers: clientmodel.EventTriggerDefinition{
				int64(1234 + requestCount),
			},
			ActionToBeSetupList: clientmodel.ActionsToBeSetup{
				&clientmodel.ActionToBeSetup{
					ActionID:   &actionId,
					ActionType: &actionType,
					ActionDefinition: clientmodel.ActionDefinition{
						int64(5678 + requestCount),
					},
					SubsequentAction: &clientmodel.SubsequentAction{
						SubsequentActionType: &subsequestActioType,
						TimeToWait:           &timeToWait,
					},
				},
			},
		}
		p.SubsReqParams.SubscriptionDetails = append(p.SubsReqParams.SubscriptionDetails, subscriptionDetail)
	}
}

func (p *RESTSubsReqParams) SetSubscriptionDirectives(e2Timeout int64, e2RetryCount int64, routingNeeded bool) {

	e2SubscriptionDirectives := &clientmodel.SubscriptionParamsE2SubscriptionDirectives{}
	e2SubscriptionDirectives.E2TimeoutTimerValue = e2Timeout
	e2SubscriptionDirectives.E2RetryCount = &e2RetryCount
	e2SubscriptionDirectives.RMRRoutingNeeded = routingNeeded
	p.SubsReqParams.E2SubscriptionDirectives = e2SubscriptionDirectives

}
