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

package e2ap_wrapper

// #cgo LDFLAGS: -le2ap_wrapper -le2ap  -lstdc++
// #include <stdlib.h>
// #include <c_types.h>
// #include <E2AP_if.h>
// #include <strings.h>
//
// void initSubsRequest(RICSubscriptionRequest_t *data){
//   bzero(data,sizeof(RICSubscriptionRequest_t));
// }
// void initSubsResponse(RICSubscriptionResponse_t *data){
//   bzero(data,sizeof(RICSubscriptionResponse_t));
// }
// void initSubsFailure(RICSubscriptionFailure_t *data){
//   bzero(data,sizeof(RICSubscriptionFailure_t));
// }
// void initSubsDeleteRequest(RICSubscriptionDeleteRequest_t *data){
//   bzero(data,sizeof(RICSubscriptionDeleteRequest_t));
// }
// void initSubsDeleteResponse(RICSubscriptionDeleteResponse_t *data){
//   bzero(data,sizeof(RICSubscriptionDeleteResponse_t));
// }
// void initSubsDeleteFailure(RICSubscriptionDeleteFailure_t *data){
//   bzero(data,sizeof(RICSubscriptionDeleteFailure_t));
// }
// void initSubsDeleteRequired(RICSubsDeleteRequired_t *data){
//	 bzero(data,sizeof(RICSubsDeleteRequired_t));
// }
//
import "C"

import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/conv"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
)

const cLogBufferMaxSize = 40960
const cMsgBufferMaxSize = 40960
const cMsgBufferExtra = 512

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func cMessageInfoToMessageInfo(minfo *C.E2MessageInfo_t) *e2ap.MessageInfo {

	msgInfo := &e2ap.MessageInfo{}

	switch minfo.messageType {
	case C.cE2InitiatingMessage:
		msgInfo.MsgType = e2ap.E2AP_InitiatingMessage
		switch minfo.messageId {
		case C.cRICSubscriptionRequest:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionRequest
			return msgInfo
		case C.cRICSubscriptionDeleteRequest:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteRequest
			return msgInfo
		}
	case C.cE2SuccessfulOutcome:
		msgInfo.MsgType = e2ap.E2AP_SuccessfulOutcome
		switch minfo.messageId {
		case C.cRICSubscriptionResponse:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionResponse
			return msgInfo
		case C.cRICsubscriptionDeleteResponse:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteResponse
			return msgInfo
		}
	case C.cE2UnsuccessfulOutcome:
		msgInfo.MsgType = e2ap.E2AP_UnsuccessfulOutcome
		switch minfo.messageId {
		case C.cRICSubscriptionFailure:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionFailure
			return msgInfo
		case C.cRICsubscriptionDeleteFailure:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteFailure
			return msgInfo
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRequestID struct {
	entry *C.RICRequestID_t
}

func (e2Item *e2apEntryRequestID) set(id *e2ap.RequestId) error {
	e2Item.entry.ricRequestorID = (C.uint32_t)(id.Id)
	e2Item.entry.ricInstanceID = (C.uint32_t)(id.InstanceId)
	return nil
}

func (e2Item *e2apEntryRequestID) get(id *e2ap.RequestId) error {
	id.Id = (uint32)(e2Item.entry.ricRequestorID)
	id.InstanceId = (uint32)(e2Item.entry.ricInstanceID)
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionToBeSetupItem struct {
	entry *C.RICActionToBeSetupItem_t
}

func (e2Item *e2apEntryActionToBeSetupItem) set(id *e2ap.ActionToBeSetupItem) error {

	e2Item.entry.ricActionID = (C.ulong)(id.ActionId)
	e2Item.entry.ricActionType = (C.uint64_t)(id.ActionType)
	if id.RicActionDefinitionPresent {
		e2Item.entry.ricActionDefinitionPresent = true
		if err := (&e2apEntryActionDefinitionChoice{entry: &e2Item.entry.ricActionDefinitionChoice}).set(&id.ActionDefinitionChoice); err != nil {
			return err
		}
	}
	if id.SubsequentAction.Present {
		e2Item.entry.ricSubsequentActionPresent = true
		e2Item.entry.ricSubsequentAction.ricSubsequentActionType = (C.uint64_t)(id.SubsequentAction.Type)
		e2Item.entry.ricSubsequentAction.ricTimeToWait = (C.uint64_t)(id.SubsequentAction.TimetoWait)
	}
	return nil
}

func (e2Item *e2apEntryActionToBeSetupItem) get(id *e2ap.ActionToBeSetupItem) error {

	id.ActionId = (uint64)(e2Item.entry.ricActionID)
	id.ActionType = (uint64)(e2Item.entry.ricActionType)
	if e2Item.entry.ricActionDefinitionPresent {
		id.RicActionDefinitionPresent = true
		if err := (&e2apEntryActionDefinitionChoice{entry: &e2Item.entry.ricActionDefinitionChoice}).get(&id.ActionDefinitionChoice); err != nil {
			return err
		}
	}
	if e2Item.entry.ricSubsequentActionPresent {
		id.SubsequentAction.Present = true
		id.SubsequentAction.Type = (uint64)(e2Item.entry.ricSubsequentAction.ricSubsequentActionType)
		id.SubsequentAction.TimetoWait = (uint64)(e2Item.entry.ricSubsequentAction.ricTimeToWait)
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionDefinitionChoice struct {
	entry *C.RICActionDefinitionChoice_t
}

func (e2Item *e2apEntryActionDefinitionChoice) set(id *e2ap.ActionDefinitionChoice) error {
	if id.Data.Length > 0 {
		e2Item.entry.octetString.contentLength = C.size_t(id.Data.Length)
		C.memcpy(unsafe.Pointer(&e2Item.entry.octetString.data[0]), unsafe.Pointer(&id.Data.Data[0]), C.size_t(e2Item.entry.octetString.contentLength))
	}
	return nil
}

func (e2Item *e2apEntryActionDefinitionChoice) get(id *e2ap.ActionDefinitionChoice) error {
	id.Data.Length = (uint64)(e2Item.entry.octetString.contentLength)
	if id.Data.Length > 0 {
		id.Data.Data = make([]uint8, id.Data.Length)
		C.memcpy(unsafe.Pointer(&id.Data.Data[0]), unsafe.Pointer(&e2Item.entry.octetString.data[0]), C.size_t(e2Item.entry.octetString.contentLength))
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryPlmnIdentity struct {
	entry *C.PLMNIdentity_t
}

func (plmnId *e2apEntryPlmnIdentity) set(id conv.PlmnIdentityIf) error {

	buf := new(bytes.Buffer)
	id.EncodeTo(buf)
	data := buf.Bytes()
	plmnId.entry.contentLength = (C.uint8_t)(len(data))
	for i := 0; i < len(data); i++ {
		plmnId.entry.pLMNIdentityVal[i] = (C.uint8_t)(data[i])
	}
	return nil
}

func (plmnId *e2apEntryPlmnIdentity) get(id conv.PlmnIdentityIf) error {
	conlen := (int)(plmnId.entry.contentLength)
	bcdBuf := make([]uint8, conlen)
	for i := 0; i < conlen; i++ {
		bcdBuf[i] = (uint8)(plmnId.entry.pLMNIdentityVal[i])
	}
	reader := bytes.NewReader(bcdBuf)
	id.DecodeFrom(reader)
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryGlobalEnbId struct {
	entry *C.GlobalNodeID_t
}

func (enbId *e2apEntryGlobalEnbId) checkbits(bits uint8) error {
	switch bits {
	case e2ap.E2AP_ENBIDMacroPBits20:
		return nil
	case e2ap.E2AP_ENBIDHomeBits28:
		return nil
	case e2ap.E2AP_ENBIDShortMacroits18:
		return nil
	case e2ap.E2AP_ENBIDlongMacroBits21:
		return nil
	}
	return fmt.Errorf("GlobalEnbId: given bits %d not match allowed: 20,28,18,21", bits)
}

func (enbId *e2apEntryGlobalEnbId) set(id *e2ap.GlobalNodeId) error {
	if err := enbId.checkbits(id.NodeId.Bits); err != nil {
		return err
	}
	enbId.entry.nodeID.bits = (C.uchar)(id.NodeId.Bits)
	enbId.entry.nodeID.nodeID = (C.uint32_t)(id.NodeId.Id)
	return (&e2apEntryPlmnIdentity{entry: &enbId.entry.pLMNIdentity}).set(&id.PlmnIdentity)
}

func (enbId *e2apEntryGlobalEnbId) get(id *e2ap.GlobalNodeId) error {
	if err := enbId.checkbits((uint8)(enbId.entry.nodeID.bits)); err != nil {
		return err
	}
	id.NodeId.Bits = (uint8)(enbId.entry.nodeID.bits)
	id.NodeId.Id = (uint32)(enbId.entry.nodeID.nodeID)
	return (&e2apEntryPlmnIdentity{entry: &enbId.entry.pLMNIdentity}).get(&id.PlmnIdentity)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryGlobalGnbId struct {
	entry *C.GlobalNodeID_t
}

func (gnbId *e2apEntryGlobalGnbId) checkbits(bits uint8) error {
	if bits < 22 || bits > 32 {
		return fmt.Errorf("GlobalGnbId: given bits %d not match allowed: 22-32", bits)
	}
	return nil
}

func (gnbId *e2apEntryGlobalGnbId) set(id *e2ap.GlobalNodeId) error {
	if err := gnbId.checkbits(id.NodeId.Bits); err != nil {
		return err
	}
	gnbId.entry.nodeID.bits = (C.uchar)(id.NodeId.Bits)
	gnbId.entry.nodeID.nodeID = (C.uint32_t)(id.NodeId.Id)
	return (&e2apEntryPlmnIdentity{entry: &gnbId.entry.pLMNIdentity}).set(&id.PlmnIdentity)
}

func (gnbId *e2apEntryGlobalGnbId) get(id *e2ap.GlobalNodeId) error {
	if err := gnbId.checkbits((uint8)(gnbId.entry.nodeID.bits)); err != nil {
		return err
	}
	id.NodeId.Bits = (uint8)(gnbId.entry.nodeID.bits)
	id.NodeId.Id = (uint32)(gnbId.entry.nodeID.nodeID)
	return (&e2apEntryPlmnIdentity{entry: &gnbId.entry.pLMNIdentity}).get(&id.PlmnIdentity)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryInterfaceId struct {
	entry *C.InterfaceID_t
}

func (indId *e2apEntryInterfaceId) set(id *e2ap.InterfaceId) error {
	if id.GlobalEnbId.Present {
		indId.entry.globalENBIDPresent = true
		if err := (&e2apEntryGlobalEnbId{entry: &indId.entry.globalENBID}).set(&id.GlobalEnbId); err != nil {
			return err
		}
	}

	if id.GlobalGnbId.Present {
		indId.entry.globalGNBIDPresent = true
		if err := (&e2apEntryGlobalGnbId{entry: &indId.entry.globalGNBID}).set(&id.GlobalGnbId); err != nil {
			return err
		}
	}
	return nil
}

func (indId *e2apEntryInterfaceId) get(id *e2ap.InterfaceId) error {
	if indId.entry.globalENBIDPresent == true {
		id.GlobalEnbId.Present = true
		if err := (&e2apEntryGlobalEnbId{entry: &indId.entry.globalENBID}).get(&id.GlobalEnbId); err != nil {
			return err
		}
	}

	if indId.entry.globalGNBIDPresent == true {
		id.GlobalGnbId.Present = true
		if err := (&e2apEntryGlobalGnbId{entry: &indId.entry.globalGNBID}).get(&id.GlobalGnbId); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryEventTrigger struct {
	entry *C.RICEventTriggerDefinition_t
}

func (evtTrig *e2apEntryEventTrigger) set(id *e2ap.EventTriggerDefinition) error {

	if id.Data.Length > 0 {
		evtTrig.entry.octetString.contentLength = C.size_t(id.Data.Length)
		C.memcpy(unsafe.Pointer(&evtTrig.entry.octetString.data[0]), unsafe.Pointer(&id.Data.Data[0]), C.size_t(evtTrig.entry.octetString.contentLength))
	}
	return nil
}

func (evtTrig *e2apEntryEventTrigger) get(id *e2ap.EventTriggerDefinition) error {

	id.Data.Length = (uint64)(evtTrig.entry.octetString.contentLength)
	if id.Data.Length > 0 {
		id.Data.Data = make([]uint8, id.Data.Length)
		C.memcpy(unsafe.Pointer(&id.Data.Data[0]), unsafe.Pointer(&evtTrig.entry.octetString.data[0]), C.size_t(evtTrig.entry.octetString.contentLength))
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryAdmittedList struct {
	entry *C.RICActionAdmittedList_t
}

func (item *e2apEntryAdmittedList) set(data *e2ap.ActionAdmittedList) error {

	if len(data.Items) > 16 {
		return fmt.Errorf("ActionAdmittedList: too long %d while allowed %d", len(data.Items), 16)
	}

	item.entry.contentLength = 0
	for i := 0; i < len(data.Items); i++ {
		item.entry.ricActionID[item.entry.contentLength] = (C.ulong)(data.Items[i].ActionId)
		item.entry.contentLength++
	}
	return nil
}

func (item *e2apEntryAdmittedList) get(data *e2ap.ActionAdmittedList) error {
	conlen := (int)(item.entry.contentLength)
	data.Items = make([]e2ap.ActionAdmittedItem, conlen)
	for i := 0; i < conlen; i++ {
		data.Items[i].ActionId = (uint64)(item.entry.ricActionID[i])
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryNotAdmittedList struct {
	entry *C.RICActionNotAdmittedList_t
}

func (item *e2apEntryNotAdmittedList) set(data *e2ap.ActionNotAdmittedList) error {

	if len(data.Items) > 16 {
		return fmt.Errorf("e2apEntryNotAdmittedList: too long %d while allowed %d", len(data.Items), 16)
	}

	item.entry.contentLength = 0
	for i := 0; i < len(data.Items); i++ {
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].ricActionID = (C.ulong)(data.Items[i].ActionId)
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].cause.content = (C.uchar)(data.Items[i].Cause.Content)
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].cause.causeVal = (C.uchar)(data.Items[i].Cause.Value)
		item.entry.contentLength++
	}

	return nil
}

func (item *e2apEntryNotAdmittedList) get(data *e2ap.ActionNotAdmittedList) error {
	conlen := (int)(item.entry.contentLength)
	data.Items = make([]e2ap.ActionNotAdmittedItem, conlen)
	for i := 0; i < conlen; i++ {
		data.Items[i].ActionId = (uint64)(item.entry.RICActionNotAdmittedItem[i].ricActionID)
		data.Items[i].Cause.Content = (uint8)(item.entry.RICActionNotAdmittedItem[i].cause.content)
		data.Items[i].Cause.Value = (uint8)(item.entry.RICActionNotAdmittedItem[i].cause.causeVal)
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryCriticalityDiagnostic struct {
	entry *C.CriticalityDiagnostics__t
}

func (item *e2apEntryCriticalityDiagnostic) set(data *e2ap.CriticalityDiagnostics) error {

	item.entry.procedureCodePresent = (C.bool)(data.ProcCodePresent)
	item.entry.procedureCode = (C.uchar)(data.ProcCode)

	item.entry.triggeringMessagePresent = (C.bool)(data.TrigMsgPresent)
	item.entry.triggeringMessage = (C.uchar)(data.TrigMsg)

	item.entry.procedureCriticalityPresent = (C.bool)(data.ProcCritPresent)
	item.entry.procedureCriticality = (C.uchar)(data.ProcCrit)

	item.entry.criticalityDiagnosticsIELength = 0
	item.entry.iEsCriticalityDiagnosticsPresent = false
	for i := 0; i < len(data.CriticalityDiagnosticsIEList.Items); i++ {
		item.entry.criticalityDiagnosticsIEListItem[i].iECriticality = (C.uint8_t)(data.CriticalityDiagnosticsIEList.Items[i].IeCriticality)
		item.entry.criticalityDiagnosticsIEListItem[i].iE_ID = (C.uint32_t)(data.CriticalityDiagnosticsIEList.Items[i].IeID)
		item.entry.criticalityDiagnosticsIEListItem[i].typeOfError = (C.uint8_t)(data.CriticalityDiagnosticsIEList.Items[i].TypeOfError)
		item.entry.criticalityDiagnosticsIELength++
		item.entry.iEsCriticalityDiagnosticsPresent = true
	}
	return nil
}

func (item *e2apEntryCriticalityDiagnostic) get(data *e2ap.CriticalityDiagnostics) error {

	data.ProcCodePresent = (bool)(item.entry.procedureCodePresent)
	data.ProcCode = (uint64)(item.entry.procedureCode)

	data.TrigMsgPresent = (bool)(item.entry.triggeringMessagePresent)
	data.TrigMsg = (uint64)(item.entry.triggeringMessage)

	data.ProcCritPresent = (bool)(item.entry.procedureCriticalityPresent)
	data.ProcCrit = (uint8)(item.entry.procedureCriticality)

	if item.entry.iEsCriticalityDiagnosticsPresent == true {
		conlen := (int)(item.entry.criticalityDiagnosticsIELength)
		data.CriticalityDiagnosticsIEList.Items = make([]e2ap.CriticalityDiagnosticsIEListItem, conlen)
		for i := 0; i < conlen; i++ {
			data.CriticalityDiagnosticsIEList.Items[i].IeCriticality = (uint8)(item.entry.criticalityDiagnosticsIEListItem[i].iECriticality)
			data.CriticalityDiagnosticsIEList.Items[i].IeID = (uint32)(item.entry.criticalityDiagnosticsIEListItem[i].iE_ID)
			data.CriticalityDiagnosticsIEList.Items[i].TypeOfError = (uint8)(item.entry.criticalityDiagnosticsIEListItem[i].typeOfError)
		}
	}
	return nil
}

/*
//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryCallProcessId struct {
	entry *C.RICCallProcessID_t
}

func (callProcId *e2apEntryCallProcessId) set(data *e2ap.CallProcessId) error {
	callProcId.entry.ricCallProcessIDVal = (C.uint64_t)(data.CallProcessIDVal)
	return nil
}

func (callProcId *e2apEntryCallProcessId) get(data *e2ap.CallProcessId) error {
	data.CallProcessIDVal = (uint32)(callProcId.entry.ricCallProcessIDVal)
	return nil
}
*/

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type e2apMessagePacker struct {
	expectedInfo C.E2MessageInfo_t
	pduMsgInfo   C.E2MessageInfo_t
	pdu          *C.e2ap_pdu_ptr_t
	lb           []byte
	p            unsafe.Pointer
	plen         C.size_t
}

func (e2apMsg *e2apMessagePacker) init(minfo C.E2MessageInfo_t) {
	e2apMsg.expectedInfo = minfo
	e2apMsg.lb = make([]byte, cLogBufferMaxSize)
	e2apMsg.lb[0] = 0
	e2apMsg.p = C.malloc(C.size_t(cMsgBufferMaxSize))
	e2apMsg.plen = C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
}

func (e2apMsg *e2apMessagePacker) fini() {
	C.free(e2apMsg.p)
	e2apMsg.plen = 0
	e2apMsg.p = nil
}

func (e2apMsg *e2apMessagePacker) lbString() string {
	return "logbuffer(" + string(e2apMsg.lb[:strings.Index(string(e2apMsg.lb[:]), "\000")]) + ")"
}

func (e2apMsg *e2apMessagePacker) packeddata() *e2ap.PackedData {
	return &e2ap.PackedData{C.GoBytes(e2apMsg.p, C.int(e2apMsg.plen))}
}

func (e2apMsg *e2apMessagePacker) checkerr(errorNro C.uint64_t) error {
	if errorNro != C.e2err_OK {
		return fmt.Errorf("e2err(%s) %s", C.GoString(C.getE2ErrorString(errorNro)), e2apMsg.lbString())
	}
	return nil
}

func (e2apMsg *e2apMessagePacker) unpacktopdu(data *e2ap.PackedData) error {
	e2apMsg.pdu = C.unpackE2AP_pdu((C.size_t)(len(data.Buf)), (*C.uchar)(unsafe.Pointer(&data.Buf[0])), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), &e2apMsg.pduMsgInfo)
	if e2apMsg.pduMsgInfo.messageType != e2apMsg.expectedInfo.messageType || e2apMsg.pduMsgInfo.messageId != e2apMsg.expectedInfo.messageId {
		return fmt.Errorf("unpack e2ap %s %s", e2apMsg.lbString(), e2apMsg.String())
	}
	return nil
}

func (e2apMsg *e2apMessagePacker) messageInfoPdu() *e2ap.MessageInfo {
	return cMessageInfoToMessageInfo(&e2apMsg.pduMsgInfo)
}

func (e2apMsg *e2apMessagePacker) messageInfoExpected() *e2ap.MessageInfo {
	return cMessageInfoToMessageInfo(&e2apMsg.expectedInfo)
}

func (e2apMsg *e2apMessagePacker) String() string {
	var ret string
	pduInfo := e2apMsg.messageInfoPdu()
	if pduInfo != nil {
		ret += "pduinfo(" + pduInfo.String() + ")"
	} else {
		ret += "pduinfo(N/A)"
	}
	expInfo := e2apMsg.messageInfoExpected()
	if expInfo != nil {
		ret += " expinfo(" + expInfo.String() + ")"
	} else {
		ret += " expinfo(N/A)"
	}
	return ret
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type e2apMsgPackerSubscriptionRequest struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionRequest_t
	msgG *e2ap.E2APSubscriptionRequest
}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2InitiatingMessage, C.cRICSubscriptionRequest})
	e2apMsg.msgC = &C.RICSubscriptionRequest_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionRequest{}
	C.initSubsRequest(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) Pack(data *e2ap.E2APSubscriptionRequest) (error, *e2ap.PackedData) {

	e2apMsg.init()

	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryEventTrigger{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition}).set(&e2apMsg.msgG.EventTriggerDefinition); err != nil {
		return err, nil
	}
	if len(e2apMsg.msgG.ActionSetups) > 16 {
		return fmt.Errorf("IndicationMessage.InterfaceMessage: too long %d while allowed %d", len(e2apMsg.msgG.ActionSetups), 16), nil
	}
	e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = 0
	for i := 0; i < len(e2apMsg.msgG.ActionSetups); i++ {
		item := &e2apEntryActionToBeSetupItem{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength]}
		e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength += 1
		if err := item.set(&e2apMsg.msgG.ActionSetups[i]); err != nil {
			return err, nil
		}
	}
	errorNro := C.packRICSubscriptionRequest(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionRequest) {

	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}

	errorNro := C.getRICSubscriptionRequestData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	if err := (&e2apEntryEventTrigger{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition}).get(&e2apMsg.msgG.EventTriggerDefinition); err != nil {
		return err, e2apMsg.msgG
	}
	conlen := (int)(e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength)
	e2apMsg.msgG.ActionSetups = make([]e2ap.ActionToBeSetupItem, conlen)
	for i := 0; i < conlen; i++ {
		item := &e2apEntryActionToBeSetupItem{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[i]}
		if err := item.get(&e2apMsg.msgG.ActionSetups[i]); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG

}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionRequest.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "     ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "     ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricSubscriptionDetails.")
	fmt.Fprintln(&b, "    ricActionToBeSetupItemIEs.")
	fmt.Fprintln(&b, "      contentLength =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength {
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionID =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID)
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionType =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType)
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinitionPresent =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent)
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentActionPresent =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent)
		if e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent {
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType)
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait)
		}
		index++
	}
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionResponse struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionResponse_t
	msgG *e2ap.E2APSubscriptionResponse
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2SuccessfulOutcome, C.cRICSubscriptionResponse})
	e2apMsg.msgC = &C.RICSubscriptionResponse_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionResponse{}
	C.initSubsResponse(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) Pack(data *e2ap.E2APSubscriptionResponse) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryAdmittedList{entry: &e2apMsg.msgC.ricActionAdmittedList}).set(&e2apMsg.msgG.ActionAdmittedList); err != nil {
		return err, nil
	}
	e2apMsg.msgC.ricActionNotAdmittedListPresent = false
	if len(e2apMsg.msgG.ActionNotAdmittedList.Items) > 0 {
		e2apMsg.msgC.ricActionNotAdmittedListPresent = true
		if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).set(&e2apMsg.msgG.ActionNotAdmittedList); err != nil {
			return err, nil
		}
	}

	errorNro := C.packRICSubscriptionResponse(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionResponse) {

	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}

	errorNro := C.getRICSubscriptionResponseData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	if err := (&e2apEntryAdmittedList{entry: &e2apMsg.msgC.ricActionAdmittedList}).get(&e2apMsg.msgG.ActionAdmittedList); err != nil {
		return err, e2apMsg.msgG
	}
	if e2apMsg.msgC.ricActionNotAdmittedListPresent == true {
		if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).get(&e2apMsg.msgG.ActionNotAdmittedList); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionResponse.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricActionAdmittedList.")
	fmt.Fprintln(&b, "    contentLength =", e2apMsg.msgC.ricActionAdmittedList.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricActionAdmittedList.contentLength {
		fmt.Fprintln(&b, "    ricActionAdmittedList.ricActionID[index] =", e2apMsg.msgC.ricActionAdmittedList.ricActionID[index])
		index++
	}
	if e2apMsg.msgC.ricActionNotAdmittedListPresent {
		fmt.Fprintln(&b, "  ricActionNotAdmittedListPresent =", e2apMsg.msgC.ricActionNotAdmittedListPresent)
		fmt.Fprintln(&b, "    ricActionNotAdmittedList.")
		fmt.Fprintln(&b, "    contentLength =", e2apMsg.msgC.ricActionNotAdmittedList.contentLength)
		index = 0
		for (C.uchar)(index) < e2apMsg.msgC.ricActionNotAdmittedList.contentLength {
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].ricActionID =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID)
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].cause.content =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content)
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].cause.causeVal =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal)
			index++
		}
	}
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionFailure struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionFailure_t
	msgG *e2ap.E2APSubscriptionFailure
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2UnsuccessfulOutcome, C.cRICSubscriptionFailure})
	e2apMsg.msgC = &C.RICSubscriptionFailure_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionFailure{}
	C.initSubsFailure(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) Pack(data *e2ap.E2APSubscriptionFailure) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	e2apMsg.msgC.cause.content = (C.uchar)(e2apMsg.msgG.Cause.Content)
	e2apMsg.msgC.cause.causeVal = (C.uchar)(e2apMsg.msgG.Cause.Value)
	e2apMsg.msgC.criticalityDiagnosticsPresent = false
	if e2apMsg.msgG.CriticalityDiagnostics.Present {
		e2apMsg.msgC.criticalityDiagnosticsPresent = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).set(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, nil
		}
	}

	errorNro := C.packRICSubscriptionFailure(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionFailure) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionFailureData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	e2apMsg.msgG.Cause.Content = (uint8)(e2apMsg.msgC.cause.content)
	e2apMsg.msgG.Cause.Value = (uint8)(e2apMsg.msgC.cause.causeVal)
	if e2apMsg.msgC.criticalityDiagnosticsPresent == true {
		e2apMsg.msgG.CriticalityDiagnostics.Present = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).get(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionFailure.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricActionNotAdmittedList.")
	fmt.Fprintln(&b, "    cause.content =", e2apMsg.msgC.cause.content)
	fmt.Fprintln(&b, "    cause.causeVal =", e2apMsg.msgC.cause.causeVal)

	/* NOT SUPPORTED
	if e2apMsg.msgC.criticalityDiagnosticsPresent {
		fmt.Fprintln(&b, "  criticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnosticsPresent)
		fmt.Fprintln(&b, "    criticalityDiagnostics.")
		fmt.Fprintln(&b, "    procedureCodePresent =", e2apMsg.msgC.criticalityDiagnostics.procedureCodePresent)
		fmt.Fprintln(&b, "      procedureCode =", e2apMsg.msgC.criticalityDiagnostics.procedureCode)
		fmt.Fprintln(&b, "    triggeringMessagePresent =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessagePresent)
		fmt.Fprintln(&b, "      triggeringMessage =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessage)
		fmt.Fprintln(&b, "    procedureCriticalityPresent=", e2apMsg.msgC.criticalityDiagnostics.procedureCriticalityPresent)
		fmt.Fprintln(&b, "      procedureCriticality =", e2apMsg.msgC.criticalityDiagnostics.procedureCriticality)
		fmt.Fprintln(&b, "    iEsCriticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnostics.iEsCriticalityDiagnosticsPresent)
		fmt.Fprintln(&b, "      criticalityDiagnosticsIELength =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength)
		var index2 uint16
		index2 = 0
		for (C.ushort)(index2) < e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength {
			fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iECriticality =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iECriticality)
			fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iE_ID =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iE_ID)
			fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].typeOfError =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].typeOfError)
			index2++
		}
	}
	*/
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionDeleteRequest struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionDeleteRequest_t
	msgG *e2ap.E2APSubscriptionDeleteRequest
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2InitiatingMessage, C.cRICSubscriptionDeleteRequest})
	e2apMsg.msgC = &C.RICSubscriptionDeleteRequest_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionDeleteRequest{}
	C.initSubsDeleteRequest(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) Pack(data *e2ap.E2APSubscriptionDeleteRequest) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}

	errorNro := C.packRICSubscriptionDeleteRequest(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteRequest) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteRequestData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	return nil, e2apMsg.msgG

}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteRequest.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "     ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "     ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionDeleteResponse struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionDeleteResponse_t
	msgG *e2ap.E2APSubscriptionDeleteResponse
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2SuccessfulOutcome, C.cRICsubscriptionDeleteResponse})
	e2apMsg.msgC = &C.RICSubscriptionDeleteResponse_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionDeleteResponse{}
	C.initSubsDeleteResponse(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) Pack(data *e2ap.E2APSubscriptionDeleteResponse) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}

	errorNro := C.packRICSubscriptionDeleteResponse(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteResponse) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteResponseData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteResponse.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionDeleteFailure struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionDeleteFailure_t
	msgG *e2ap.E2APSubscriptionDeleteFailure
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2UnsuccessfulOutcome, C.cRICsubscriptionDeleteFailure})
	e2apMsg.msgC = &C.RICSubscriptionDeleteFailure_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionDeleteFailure{}
	C.initSubsDeleteFailure(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) Pack(data *e2ap.E2APSubscriptionDeleteFailure) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	e2apMsg.msgC.cause.content = (C.uchar)(e2apMsg.msgG.Cause.Content)
	e2apMsg.msgC.cause.causeVal = (C.uchar)(e2apMsg.msgG.Cause.Value)
	e2apMsg.msgC.criticalityDiagnosticsPresent = false
	if e2apMsg.msgG.CriticalityDiagnostics.Present {
		e2apMsg.msgC.criticalityDiagnosticsPresent = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).set(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, nil
		}
	}

	errorNro := C.packRICSubscriptionDeleteFailure(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteFailure) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteFailureData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	e2apMsg.msgG.Cause.Content = (uint8)(e2apMsg.msgC.cause.content)
	e2apMsg.msgG.Cause.Value = (uint8)(e2apMsg.msgC.cause.causeVal)
	if e2apMsg.msgC.criticalityDiagnosticsPresent == true {
		e2apMsg.msgG.CriticalityDiagnostics.Present = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).get(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteFailure.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	/*	NOT SUPPORTED
		if e2apMsg.msgC.criticalityDiagnosticsPresent {
			fmt.Fprintln(&b, "  criticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnosticsPresent)
			fmt.Fprintln(&b, "    criticalityDiagnostics.")
			fmt.Fprintln(&b, "    procedureCodePresent =", e2apMsg.msgC.criticalityDiagnostics.procedureCodePresent)
			fmt.Fprintln(&b, "      procedureCode =", e2apMsg.msgC.criticalityDiagnostics.procedureCode)
			fmt.Fprintln(&b, "    triggeringMessagePresent =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessagePresent)
			fmt.Fprintln(&b, "      triggeringMessage =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessage)
			fmt.Fprintln(&b, "    procedureCriticalityPresent=", e2apMsg.msgC.criticalityDiagnostics.procedureCriticalityPresent)
			fmt.Fprintln(&b, "      procedureCriticality =", e2apMsg.msgC.criticalityDiagnostics.procedureCriticality)
			fmt.Fprintln(&b, "    iEsCriticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnostics.iEsCriticalityDiagnosticsPresent)
			fmt.Fprintln(&b, "      criticalityDiagnosticsIELength =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength)
			var index2 uint16
			index2 = 0
			for (C.ushort)(index2) < e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength {
				fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iECriticality =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iECriticality)
				fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iE_ID =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iE_ID)
				fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].typeOfError =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].typeOfError)
				index2++
			}
		}
	*/
	return b.String()
}

//-----------------------------------------------------------------------------
// Changes to support "RIC_SUB_DEL_REQUIRED"
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionDeleteRequired struct {
	e2apMessagePacker
	msgC *C.RICSubsDeleteRequired_t
	msgG *e2ap.SubscriptionDeleteRequiredList
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequired) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2InitiatingMessage, C.cRICSubscriptionDeleteRequired})
	e2apMsg.msgC = &C.RICSubsDeleteRequired_t{}
	e2apMsg.msgG = &e2ap.SubscriptionDeleteRequiredList{}
	C.initSubsDeleteRequired(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequired) Pack(data *e2ap.SubscriptionDeleteRequiredList) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.noOfRanSubscriptions = C.int(len(e2apMsg.msgG.E2APSubscriptionDeleteRequiredRequests))
	for idx, subs := range e2apMsg.msgG.E2APSubscriptionDeleteRequiredRequests {

		// RIC Request ID
		e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ricRequestID.ricInstanceID = (C.uint32_t)(subs.RequestId.InstanceId)
		e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ricRequestID.ricRequestorID = (C.uint32_t)(subs.RequestId.Id)

		// RAN Function ID
		e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ranFunctionID = (C.uint16_t)(subs.FunctionId)

		// RIC Cause
		e2apMsg.msgC.ranSubscriptionsDelRequired[idx].cause.content = (C.uint8_t)(subs.Cause.Content)
		e2apMsg.msgC.ranSubscriptionsDelRequired[idx].cause.causeVal = (C.uint8_t)(subs.Cause.Value)

	}

	errorNro := C.packRICSubscriptionDeleteRequired(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		fmt.Printf("ERROR: %s", err.Error())
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequired) UnPack(msg *e2ap.PackedData) (error, *e2ap.SubscriptionDeleteRequiredList) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteRequiredData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	//TODO: Fill List of RIC Subscriptions to be Removed
	for idx := 0; idx < int(e2apMsg.msgC.noOfRanSubscriptions); idx++ {
		var ricSubsToBeRemove e2ap.E2APSubscriptionDeleteRequired
		// RIC RequestID
		if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ricRequestID}).get(&ricSubsToBeRemove.RequestId); err != nil {
			return err, e2apMsg.msgG
		}
		// RAN Function ID
		ricSubsToBeRemove.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ranFunctionID)

		// RIC Cause
		ricSubsToBeRemove.Cause.Content = (uint8)(e2apMsg.msgC.ranSubscriptionsDelRequired[idx].cause.content)
		ricSubsToBeRemove.Cause.Value = (uint8)(e2apMsg.msgC.ranSubscriptionsDelRequired[idx].cause.causeVal)

		e2apMsg.msgG.E2APSubscriptionDeleteRequiredRequests = append(e2apMsg.msgG.E2APSubscriptionDeleteRequiredRequests, ricSubsToBeRemove)
	}

	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequired) String() string {
	var b bytes.Buffer
	for idx := 0; idx < int(e2apMsg.msgC.noOfRanSubscriptions); idx++ {
		fmt.Fprintln(&b, "ricSubscriptionDeleteRequired.")
		fmt.Fprintln(&b, "  ricRequestID.")
		fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ricRequestID.ricRequestorID)
		fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ricRequestID.ricInstanceID)
		fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranSubscriptionsDelRequired[idx].ranFunctionID)
	}

	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func SetASN1DebugPrintStatus(logLevel int) {
	if logLevel >= 4 {
		//fmt.Println("ASN1 debug prints ON")
		C.allowASN1DebugPrints(true)
	} else {
		//fmt.Println("ASN1 debug prints OFF")
		C.allowASN1DebugPrints(false)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func SetE2IEOrderCheck(ieOrderCheck uint8) {
	C.allowOutOfOrderIEMsg(C.uchar(ieOrderCheck))
}

//-----------------------------------------------------------------------------
// Public E2AP packer creators
//-----------------------------------------------------------------------------

type cppasn1E2APPacker struct{}

func (*cppasn1E2APPacker) NewPackerSubscriptionRequest() e2ap.E2APMsgPackerSubscriptionRequestIf {
	return &e2apMsgPackerSubscriptionRequest{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionResponse() e2ap.E2APMsgPackerSubscriptionResponseIf {
	return &e2apMsgPackerSubscriptionResponse{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionFailure() e2ap.E2APMsgPackerSubscriptionFailureIf {
	return &e2apMsgPackerSubscriptionFailure{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteRequest() e2ap.E2APMsgPackerSubscriptionDeleteRequestIf {
	return &e2apMsgPackerSubscriptionDeleteRequest{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteResponse() e2ap.E2APMsgPackerSubscriptionDeleteResponseIf {
	return &e2apMsgPackerSubscriptionDeleteResponse{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteFailure() e2ap.E2APMsgPackerSubscriptionDeleteFailureIf {
	return &e2apMsgPackerSubscriptionDeleteFailure{}
}

// Changes to support "RIC_SUB_DEL_REQUIRED"
func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteRequired() e2ap.E2APMsgPackerSubscriptionDeleteRequiredIf {
	return &e2apMsgPackerSubscriptionDeleteRequired{}
}

func NewAsn1E2Packer() e2ap.E2APPackerIf {
	return &cppasn1E2APPacker{}
}
