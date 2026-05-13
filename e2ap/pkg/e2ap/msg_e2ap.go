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

package e2ap

import (
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/conv"
	"strconv"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_InitiatingMessage   uint64 = 1
	E2AP_SuccessfulOutcome   uint64 = 2
	E2AP_UnsuccessfulOutcome uint64 = 3
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
// E2AP messages
// Initiating message
const (
	E2AP_RICSubscriptionRequest       uint64 = 1
	E2AP_RICSubscriptionDeleteRequest uint64 = 2

	// E2AP_RICServiceUpdate uint64 = 3
	// E2AP_RICControlRequest uint64 = 4
	//
	// //E2AP_X2SetupRequest uint64 = 5;
	// E2AP_ENDCX2SetupRequest uint64 = 6
	// E2AP_ResourceStatusRequest uint64 = 7
	// E2AP_ENBConfigurationUpdate uint64 = 8
	// E2AP_ENDCConfigurationUpdate uint64 = 9
	// E2AP_ResetRequest uint64 = 10
	// E2AP_RICIndication uint64 = 11

	// E2AP_RICServiceQuery uint64 = 12
	// E2AP_LoadInformation uint64 = 13
	// E2AP_GNBStatusIndication uint64 = 14
	// E2AP_ResourceStatusUpdate uint64 = 15
	// E2AP_ErrorIndication uint64 = 16
	//
)

// E2AP messages
// Successful outcome
const (
	E2AP_RICSubscriptionResponse       uint64 = 1
	E2AP_RICSubscriptionDeleteResponse uint64 = 2

	// E2AP_RICserviceUpdateAcknowledge uint64 = 3
	// E2AP_RICcontrolAcknowledge uint64 = 4
	//
	// //E2AP_X2SetupResponse uint64 = 5;
	// E2AP_ENDCX2SetupResponse uint64 = 6
	// E2AP_ResourceStatusResponse uint64 = 7
	// E2AP_ENBConfigurationUpdateAcknowledge uint64 = 8
	// E2AP_ENDCConfigurationUpdateAcknowledge uint64 = 9
	// E2AP_ResetResponse uint64 = 10
	//
)

// E2AP messages
// Unsuccessful outcome
const (
	E2AP_RICSubscriptionFailure       uint64 = 1
	E2AP_RICSubscriptionDeleteFailure uint64 = 2

	// E2AP_RICserviceUpdateFailure uint64 = 3
	// E2AP_RICcontrolFailure uint64 = 4
	//
	// //E2AP_X2SetupFailure uint64 = 5;
	// E2AP_ENDCX2SetupFailure uint64 = 6
	// E2AP_ResourceStatusFailure uint64 = 7
	// E2AP_ENBConfigurationUpdateFailure uint64 = 8
	// E2AP_ENDCConfigurationUpdateFailure uint64 = 9
	//
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type PackedData struct {
	Buf []byte
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type MessageInfo struct {
	MsgType uint64
	MsgId   uint64
}

func (msgInfo *MessageInfo) String() string {
	return "msginfo(" + strconv.FormatUint((uint64)(msgInfo.MsgType), 10) + string(":") + strconv.FormatUint((uint64)(msgInfo.MsgId), 10) + ")"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RequestId struct {
	Id         uint32 //RequestorId
	InstanceId uint32 //Same as SubId in many place in the code. Name changed in E2 spec SubId -> InstanceId
}

func (rid *RequestId) String() string {
	return strconv.FormatUint((uint64)(rid.Id), 10) + string(":") + strconv.FormatUint((uint64)(rid.InstanceId), 10)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type FunctionId uint16

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

const (
	E2AP_ENBIDMacroPBits20    uint8 = 20
	E2AP_ENBIDHomeBits28      uint8 = 28
	E2AP_ENBIDShortMacroits18 uint8 = 18
	E2AP_ENBIDlongMacroBits21 uint8 = 21
)

type NodeId struct {
	Bits uint8
	Id   uint32
}

func (nid *NodeId) String() string {
	return strconv.FormatUint((uint64)(nid.Id), 10)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type GlobalNodeId struct {
	Present      bool
	PlmnIdentity conv.PlmnIdentityTbcd
	NodeId       NodeId
}

func (gnid *GlobalNodeId) String() string {
	return gnid.PlmnIdentity.String() + string(":") + gnid.NodeId.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type InterfaceId struct {
	GlobalEnbId GlobalNodeId
	GlobalGnbId GlobalNodeId
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

const (
	E2AP_InterfaceDirectionIncoming uint32 = 0
	E2AP_InterfaceDirectionOutgoing uint32 = 1
)

type EventTriggerDefinition struct {
	Data OctetString
}

/*
//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CallProcessId struct {
  CallProcessIDVal uint32
}
*/

type ActionDefinitionChoice struct {
	Data OctetString
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type BitString struct {
	UnusedBits uint8
	Length     uint64
	Data       []uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type OctetString struct {
	Length uint64
	Data   []uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_SubSeqActionTypeContinue uint64 = 0
	E2AP_SubSeqActionTypeWait     uint64 = 1
)

var E2AP_SubSeqActionTypeStrMap = map[string]uint64{
	"continue": E2AP_SubSeqActionTypeContinue,
	"wait":     E2AP_SubSeqActionTypeWait,
}

const (
	E2AP_TimeToWaitZero   uint64 = 0
	E2AP_TimeToWaitW1ms   uint64 = 1
	E2AP_TimeToWaitW2ms   uint64 = 2
	E2AP_TimeToWaitW5ms   uint64 = 3
	E2AP_TimeToWaitW10ms  uint64 = 4
	E2AP_TimeToWaitW20ms  uint64 = 5
	E2AP_TimeToWaitW30ms  uint64 = 6
	E2AP_TimeToWaitW40ms  uint64 = 7
	E2AP_TimeToWaitW50ms  uint64 = 8
	E2AP_TimeToWaitW100ms uint64 = 9
	E2AP_TimeToWaitW200ms uint64 = 10
	E2AP_TimeToWaitW500ms uint64 = 11
	E2AP_TimeToWaitW1s    uint64 = 12
	E2AP_TimeToWaitW2s    uint64 = 13
	E2AP_TimeToWaitW5s    uint64 = 14
	E2AP_TimeToWaitW10s   uint64 = 15
	E2AP_TimeToWaitW20s   uint64 = 16
	E2AP_TimeToWaitW60    uint64 = 17
)

var E2AP_TimeToWaitStrMap = map[string]uint64{
	"zero":   E2AP_TimeToWaitZero,
	"w1ms":   E2AP_TimeToWaitW1ms,
	"w2ms":   E2AP_TimeToWaitW2ms,
	"w5ms":   E2AP_TimeToWaitW5ms,
	"w10ms":  E2AP_TimeToWaitW10ms,
	"w20ms":  E2AP_TimeToWaitW20ms,
	"w30ms":  E2AP_TimeToWaitW30ms,
	"w40ms":  E2AP_TimeToWaitW40ms,
	"w50ms":  E2AP_TimeToWaitW50ms,
	"w100ms": E2AP_TimeToWaitW100ms,
	"w200ms": E2AP_TimeToWaitW200ms,
	"w500ms": E2AP_TimeToWaitW500ms,
	"w1s":    E2AP_TimeToWaitW1s,
	"w2s":    E2AP_TimeToWaitW2s,
	"w5s":    E2AP_TimeToWaitW5s,
	"w10s":   E2AP_TimeToWaitW10s,
	"w20s":   E2AP_TimeToWaitW20s,
	"w60s":   E2AP_TimeToWaitW60,
}

type SubsequentAction struct {
	Present    bool
	Type       uint64
	TimetoWait uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

const (
	E2AP_ActionTypeReport  uint64 = 0
	E2AP_ActionTypeInsert  uint64 = 1
	E2AP_ActionTypePolicy  uint64 = 2
	E2AP_ActionTypeInvalid uint64 = 99 // For RIC internal usage only
)

var E2AP_ActionTypeStrMap = map[string]uint64{
	"report": E2AP_ActionTypeReport,
	"insert": E2AP_ActionTypeInsert,
	"policy": E2AP_ActionTypePolicy,
}

type ActionToBeSetupItem struct {
	ActionId                   uint64
	ActionType                 uint64
	RicActionDefinitionPresent bool
	ActionDefinitionChoice
	SubsequentAction
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

// Cause CHOICE, E2AP-v02.00
const (
	E2AP_CauseContent_RICrequest uint8 = 1
	E2AP_CauseContent_RICservice uint8 = 2
	E2AP_CauseContent_E2node     uint8 = 3
	E2AP_CauseContent_Transport  uint8 = 4
	E2AP_CauseContent_Protocol   uint8 = 5
	E2AP_CauseContent_Misc       uint8 = 6
)

// CauseRICrequest ENUMERATED, E2AP-v02.00
const (
	E2AP_CauseValue_RICrequest_function_id_Invalid                            uint8 = 0
	E2AP_CauseValue_RICrequest_action_not_supported                           uint8 = 1
	E2AP_CauseValue_RICrequest_excessive_actions                              uint8 = 2
	E2AP_CauseValue_RICrequest_duplicate_action                               uint8 = 3
	E2AP_CauseValue_RICrequest_duplicate_event_trigger                        uint8 = 4
	E2AP_CauseValue_RICrequest_function_resource_limit                        uint8 = 5
	E2AP_CauseValue_RICrequest_request_id_unknown                             uint8 = 6
	E2AP_CauseValue_RICrequest_inconsistent_action_subsequent_action_sequence uint8 = 7
	E2AP_CauseValue_RICrequest_control_message_invalid                        uint8 = 8
	E2AP_CauseValue_RICrequest_ric_call_process_id_invalid                    uint8 = 9
	E2AP_CauseValue_RICrequest_control_timer_expired                          uint8 = 10
	E2AP_CauseValue_RICrequest_control_failed_to_execute                      uint8 = 11
	E2AP_CauseValue_RICrequest_system_not_ready                               uint8 = 12
	E2AP_CauseValue_RICrequest_unspecified                                    uint8 = 13
)

type Cause struct {
	Content uint8
	Value   uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionAdmittedItem struct {
	ActionId uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionAdmittedList struct {
	Items []ActionAdmittedItem //16
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionNotAdmittedItem struct {
	ActionId uint64
	Cause    Cause
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionNotAdmittedList struct {
	Items []ActionNotAdmittedItem //16
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_CriticalityReject uint8 = 0
	E2AP_CriticalityIgnore uint8 = 1
	E2AP_CriticalityNotify uint8 = 2
)

type CriticalityDiagnosticsIEListItem struct {
	IeCriticality uint8 //Crit
	IeID          uint32
	TypeOfError   uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CriticalityDiagnosticsIEList struct {
	Items []CriticalityDiagnosticsIEListItem //256
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CriticalityDiagnostics struct {
	Present         bool
	ProcCodePresent bool
	ProcCode        uint64
	TrigMsgPresent  bool
	TrigMsg         uint64
	ProcCritPresent bool
	ProcCrit        uint8 //Crit
	CriticalityDiagnosticsIEList
}
