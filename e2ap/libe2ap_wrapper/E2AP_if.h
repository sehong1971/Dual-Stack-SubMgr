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

#ifndef E2AP_IF_H
#define E2AP_IF_H

#ifndef ASN_DISABLE_OER_SUPPORT
#define ASN_DISABLE_OER_SUPPORT
#endif

#include <stdbool.h>
#include <stdint.h>
#include <ProcedureCode.h>

#ifdef	__cplusplus
extern "C" {
#endif

typedef unsigned char byte;

extern const uint64_t cMaxSizeOfOctetString;

typedef struct { // Octet string in ASN.1 does not have maximum length!
    size_t contentLength;
    uint8_t data[102400]; // Table size is const cMaxSizeOfOctetString
} OctetString_t;

typedef struct { // Octet string in ASN.1 does not have maximum length!
    size_t length;
    uint8_t* data;
} DynOctetString_t;

typedef struct {
    uint8_t unusedBits; // Trailing unused bits 0 - 7
    size_t byteLength;  // Length in bytes
    uint8_t* data;
} DynBitString_t;

typedef struct  {
	uint32_t ricRequestorID; // 0..65535
	uint32_t ricInstanceID;  // 0..65535
} RICRequestID_t;

typedef uint16_t RANFunctionID_t; // 0..4095

typedef uint64_t RICActionID_t; // 0..255

enum RICActionType_t {
     RICActionType_report,
     RICActionType_insert,
     RICActionType_policy
};

enum RICSubsequentActionType_t {
	RICSubsequentActionType_Continue,
	RICSubsequentActionType_wait
};

typedef struct {
    OctetString_t octetString;   // This element is E2AP spec format
} RICActionDefinitionChoice_t;

enum RICTimeToWait_t {
	RICTimeToWait_zero,
	RICTimeToWait_w1ms,
	RICTimeToWait_w2ms,
	RICTimeToWait_w5ms,
	RICTimeToWait_w10ms,
	RICTimeToWait_w20ms,
	RICTimeToWait_w30ms,
	RICTimeToWait_w40ms,
	RICTimeToWait_w50ms,
	RICTimeToWait_w100ms,
	RICTimeToWait_w200ms,
    RICTimeToWait_w500ms,
	RICTimeToWait_w1s,
	RICTimeToWait_w2s,
	RICTimeToWait_w5s,
	RICTimeToWait_w10s,
	RICTimeToWait_w20s,
	RICTimeToWait_w60s
};

typedef struct {
	uint64_t ricSubsequentActionType;  // This is type of enum RICSubsequentActionType_t
	uint64_t ricTimeToWait;  // This is type of enum RICTimeToWait_t
} RICSubsequentAction_t;

typedef struct  {
	RICActionID_t ricActionID;
	uint64_t ricActionType;  // This is type of enum RICActionType_t
	bool ricActionDefinitionPresent;
	RICActionDefinitionChoice_t ricActionDefinitionChoice;
	bool ricSubsequentActionPresent;
	RICSubsequentAction_t ricSubsequentAction;
} RICActionToBeSetupItem_t;

static const uint64_t cMaxofRICactionID = 16;

typedef struct  {
    uint8_t contentLength;
    RICActionToBeSetupItem_t ricActionToBeSetupItem[16];  // 1..16 // Table size is const cMaxofRICactionID
} RICActionToBeSetupList_t;

typedef struct {
    uint8_t contentLength;
    uint8_t pLMNIdentityVal[3];
} PLMNIdentity_t;

// size of eNB-id
extern const size_t cMacroENBIDP_20Bits;
extern const size_t cHomeENBID_28Bits;
extern const size_t cShortMacroENBID_18Bits;
extern const size_t clongMacroENBIDP_21Bits;

typedef struct {   // gNB-ID (SIZE 22..32 bits) or eNB-ID (SIZE 18, 20,21 or 28 bits)
    uint8_t bits;
    uint32_t nodeID;
} NodeID_t;

typedef struct {
	PLMNIdentity_t  pLMNIdentity;
	NodeID_t        nodeID;
}  GlobalNodeID_t;

typedef struct {   // CHOICE. Only either value can be present
	bool globalENBIDPresent;
	GlobalNodeID_t globalENBID;
	bool globalGNBIDPresent;
	GlobalNodeID_t globalGNBID;
} InterfaceID_t;

enum InterfaceDirection__t {
	InterfaceDirection__incoming,
	InterfaceDirection__outgoing
};

typedef uint8_t ProcedureCode__t;

typedef struct {
    OctetString_t octetString;   // This element is E2AP spec format
} RICEventTriggerDefinition_t;

typedef struct {
    RICEventTriggerDefinition_t ricEventTriggerDefinition;
    RICActionToBeSetupList_t ricActionToBeSetupItemIEs;
} RICSubscriptionDetails_t;

typedef struct {
    uint8_t contentLength;
	RICActionID_t ricActionID[16]; // Table size is const cMaxofRICactionID
} RICActionAdmittedList_t;

typedef struct {
    uint8_t content; // See above constants
    uint8_t causeVal; // This is type of enum CauseRIC_t
} RICCause_t;

typedef struct {
	RICActionID_t ricActionID;
    RICCause_t cause;
} RICActionNotAdmittedItem_t;

typedef struct {
    uint8_t contentLength;
    RICActionNotAdmittedItem_t RICActionNotAdmittedItem[16];  // Table size is const cMaxofRICactionID
} RICActionNotAdmittedList_t;

enum Criticality_t {
    Criticality__reject,
    Criticality__ignore,
    Criticality__notify
};

typedef uint32_t ProtocolIE_ID__t;

enum TriggeringMessage__t {
    TriggeringMessage__initiating_message,
    TriggeringMessage__successful_outcome,
    TriggeringMessage__unsuccessful_outcome
};

enum TypeOfError_t {
	TypeOfError_not_understood,
	TypeOfError_missing
};

typedef struct {
	uint8_t iECriticality; // This is type of enum Criticality_t
	ProtocolIE_ID__t iE_ID;
	uint8_t typeOfError; // This is type of enum TypeOfError_t
	//iE-Extensions  // This has no content in E2AP ASN.1 specification
} CriticalityDiagnosticsIEListItem_t;

typedef struct {
    bool procedureCodePresent;
	ProcedureCode__t procedureCode;  // OPTIONAL
	bool triggeringMessagePresent;
	uint8_t triggeringMessage;       // OPTIONAL. This is type of enum TriggeringMessage_t
	bool procedureCriticalityPresent;
	uint8_t procedureCriticality;    // OPTIONAL. This is type of enum Criticality_t
	bool ricRequestorIDPresent;
    RICRequestID_t ricRequestorID;   //OPTIONAL
	bool iEsCriticalityDiagnosticsPresent;
    uint16_t criticalityDiagnosticsIELength; // 1..256
	CriticalityDiagnosticsIEListItem_t criticalityDiagnosticsIEListItem[256];  // OPTIONAL. Table size is const cMaxNrOfErrors
} CriticalityDiagnostics__t;

typedef struct {
    OctetString_t octetString;    // E2AP spec format, the other elements for E2SM-X2 format
    uint64_t ricCallProcessIDVal;
} RICCallProcessID_t;

//////////////////////////////////////////////////////////////////////
// E2 Error codes
enum e2err {
    e2err_OK,
    e2err_RICSubscriptionRequestAllocRICrequestIDFail,
    e2err_RICSubscriptionRequestAllocRANfunctionIDFail,
    e2err_RICSubscriptionRequestAllocRICeventTriggerDefinitionBufFail,
    e2err_RICSubscriptionRequestAllocRICaction_ToBeSetup_ItemIEsFail,
    e2err_RICSubscriptionRequestAllocRICactionDefinitionBufFail,
    e2err_RICSubscriptionRequestAllocRICactionDefinitionFail,
    e2err_RICSubscriptionRequestAllocRICsubsequentActionFail,
    e2err_RICSubscriptionRequestAllocRICsubscriptionRequest_IEsFail,
    e2err_RICSubscriptionRequestEncodeFail,
    e2err_RICSubscriptionRequestAllocE2AP_PDUFail,
    e2err_RICSubscriptionResponseAllocRICrequestIDFail,
    e2err_RICSubscriptionResponseAllocRANfunctionIDFail,
    e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail,
    e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail,
    e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail,
    e2err_RICSubscriptionResponseEncodeFail,
    e2err_RICSubscriptionResponseAllocE2AP_PDUFail,
    e2err_RICSubscriptionFailureAllocRICrequestIDFail,
    e2err_RICSubscriptionFailureAllocRANfunctionIDFail,
    e2err_RICSubscriptionFailureAllocRICaction_NotAdmitted_ItemIEsFail,
    e2err_RICSubscriptionFailureAllocCauseFail,
    e2err_RICSubscriptionFailureEncodeFail,
    e2err_RICSubscriptionFailureAllocE2AP_PDUFail,
    e2err_RICSubscriptionDeleteRequestAllocRICrequestIDFail,
    e2err_RICSubscriptionDeleteRequestAllocRANfunctionIDFail,
    e2err_RICSubscriptionDeleteRequestEncodeFail,
    e2err_RICSubscriptionDeleteRequestAllocE2AP_PDUFail,
    e2err_RICSubscriptionDeleteResponseAllocRICrequestIDFail,
    e2err_RICSubscriptionDeleteResponseAllocRANfunctionIDFail,
    e2err_RICSubscriptionDeleteResponseEncodeFail,
    e2err_RICSubscriptionDeleteResponseAllocE2AP_PDUFail,
    e2err_RICSubscriptionDeleteFailureAllocRICrequestIDFail,
    e2err_RICSubscriptionDeleteFailureAllocRANfunctionIDFail,
    e2err_RICSubscriptionDeleteFailureAllocRICcauseFail,
    e2err_RICSubscriptionDeleteFailureEncodeFail,
    e2err_RICSubscriptionDeleteFailureAllocE2AP_PDUFail,
    e2err_RICsubscriptionRequestRICrequestIDMissing,
    e2err_RICsubscriptionRequestRANfunctionIDMissing,
    e2err_RICsubscriptionRequestICsubscriptionMissing,
    e2err_RICsubscriptionResponseRICrequestIDMissing,
    e2err_RICsubscriptionResponseRANfunctionIDMissing,
    e2err_RICsubscriptionResponseRICaction_Admitted_ListMissing,
    e2err_RICsubscriptionFailureRICrequestIDMissing,
    e2err_RICsubscriptionFailureRANfunctionIDMissing,
    e2err_RICsubscriptionFailureCauseMissing,
    e2err_RICsubscriptionDeleteRequestRICrequestIDMissing,
    e2err_RICsubscriptionDeleteRequestRANfunctionIDMissing,
    e2err_RICsubscriptionDeleteResponseRICrequestIDMissing,
    e2err_RICsubscriptionDeleteResponseRANfunctionIDMissing,
    e2err_RICsubscriptionDeleteFailureRICrequestIDMissing,
    e2err_RICsubscriptionDeleteFailureRANfunctionIDMissing,
    e2err_RICsubscriptionDeleteFailureRICcauseMissing,
    e2err_RICSubscriptionDeleteRequiredRICrequestIDMissing,
    e2err_RICSubscriptionDeleteRequiredRANfunctionIDMissing,
    e2err_RICSubscriptionDeleteRequiredRICcauseMissing,
    e2err_RICSubscriptionDeleteRequiredEncodeFail,
    e2err_RICSubscriptionDeleteRequiredAllocE2AP_PDUFail,
    e2err_RICsubscriptionResponseRICrequestIDWrongOrder,
    e2err_RICsubscriptionResponseRANfunctionIDWrongOrder,
    e2err_RICsubscriptionResponseRICaction_Admitted_ListWrongOrder,
    e2err_RICsubscriptionResponseRICaction_NotAdmitted_ListWrongOrder,
    e2err_RICsubscriptionFailureRICrequestIDWrongOrder,
    e2err_RICsubscriptionFailureRANfunctionIDWrongOrder,
    e2err_RICsubscriptionFailureCauseWrongOrder,
    e2err_RICsubscriptionDeleteResponseRICrequestIDWrongOrder,
    e2err_RICsubscriptionDeleteResponseRANfunctionIDWrongOrder,
    e2err_RICsubscriptionDeleteFailureRICrequestIDWrongOrder,
    e2err_RICsubscriptionDeleteFailureRANfunctionIDWrongOrder,
    e2err_RICsubscriptionDeleteFailureRICcauseWrongOrder
};

static const char* const E2ErrorStrings[] = {
    "e2err_OK",
    "e2err_RICSubscriptionRequestAllocRICrequestIDFail",
    "e2err_RICSubscriptionRequestAllocRANfunctionIDFail",
    "e2err_RICSubscriptionRequestAllocRICeventTriggerDefinitionBufFail",
    "e2err_RICSubscriptionRequestAllocRICaction_ToBeSetup_ItemIEsFail",
    "e2err_RICSubscriptionRequestAllocRICactionDefinitionBufFail",
    "e2err_RICSubscriptionRequestAllocRICactionDefinitionFail",
    "e2err_RICSubscriptionRequestAllocRICsubsequentActionFail",
    "e2err_RICSubscriptionRequestAllocRICsubscriptionRequest_IEsFail",
    "e2err_RICSubscriptionRequestEncodeFail",
    "e2err_RICSubscriptionRequestAllocE2AP_PDUFail",
    "e2err_RICSubscriptionResponseAllocRICrequestIDFail",
    "e2err_RICSubscriptionResponseAllocRANfunctionIDFail",
    "e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail",
    "e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail",
    "e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail",
    "e2err_RICSubscriptionResponseEncodeFail",
    "e2err_RICSubscriptionResponseAllocE2AP_PDUFail",
    "e2err_RICSubscriptionFailureAllocRICrequestIDFail",
    "e2err_RICSubscriptionFailureAllocRANfunctionIDFail",
    "e2err_RICSubscriptionFailureAllocRICaction_NotAdmitted_ItemIEsFail",
    "e2err_RICSubscriptionFailureAllocCauseFail",
    "e2err_RICSubscriptionFailureEncodeFail",
    "e2err_RICSubscriptionFailureAllocE2AP_PDUFail",
    "e2err_RICSubscriptionDeleteRequestAllocRICrequestIDFail",
    "e2err_RICSubscriptionDeleteRequestAllocRANfunctionIDFail",
    "e2err_RICSubscriptionDeleteRequestEncodeFail",
    "e2err_RICSubscriptionDeleteRequestAllocE2AP_PDUFail",
    "e2err_RICSubscriptionDeleteResponseAllocRICrequestIDFail",
    "e2err_RICSubscriptionDeleteResponseAllocRANfunctionIDFail",
    "e2err_RICSubscriptionDeleteResponseEncodeFail",
    "e2err_RICSubscriptionDeleteResponseAllocE2AP_PDUFail",
    "e2err_RICSubscriptionDeleteFailureAllocRICrequestIDFail",
    "e2err_RICSubscriptionDeleteFailureAllocRANfunctionIDFail",
    "e2err_RICSubscriptionDeleteFailureAllocRICcauseFail",
    "e2err_RICSubscriptionDeleteFailureEncodeFail",
    "e2err_RICSubscriptionDeleteFailureAllocE2AP_PDUFail",
    "e2err_RICsubscriptionRequestRICrequestIDMissing",
    "e2err_RICsubscriptionRequestRANfunctionIDMissing",
    "e2err_RICsubscriptionRequestICsubscriptionMissing",
    "e2err_RICsubscriptionResponseRICrequestIDMissing",
    "e2err_RICsubscriptionResponseRANfunctionIDMissing",
    "e2err_RICsubscriptionResponseRICaction_Admitted_ListMissing",
    "e2err_RICsubscriptionFailureRICrequestIDMissing",
    "e2err_RICsubscriptionFailureRANfunctionIDMissing",
    "e2err_RICsubscriptionFailureCauseMissing",
    "e2err_RICsubscriptionDeleteRequestRICrequestIDMissing",
    "e2err_RICsubscriptionDeleteRequestRANfunctionIDMissing",
    "e2err_RICsubscriptionDeleteResponseRICrequestIDMissing",
    "e2err_RICsubscriptionDeleteResponseRANfunctionIDMissing",
    "e2err_RICsubscriptionDeleteFailureRICrequestIDMissing",
    "e2err_RICsubscriptionDeleteFailureRANfunctionIDMissing",
    "e2err_RICsubscriptionDeleteFailureRICcauseMissing",
    "e2err_RICSubscriptionDeleteRequiredRICrequestIDMissing",
    "e2err_RICSubscriptionDeleteRequiredRANfunctionIDMissing",
    "e2err_RICSubscriptionDeleteRequiredRICcauseMissing",
    "e2err_RICSubscriptionDeleteRequiredEncodeFail",
    "e2err_RICSubscriptionDeleteRequiredAllocE2AP_PDUFail",
    "e2err_RICsubscriptionResponseRICrequestIDWrongOrder",
    "e2err_RICsubscriptionResponseRANfunctionIDWrongOrder",
    "e2err_RICsubscriptionResponseRICaction_Admitted_ListWrongOrder",
    "e2err_RICsubscriptionResponseRICaction_NotAdmitted_ListWrongOrder",
    "e2err_RICsubscriptionFailureRICrequestIDWrongOrder",
    "e2err_RICsubscriptionFailureRANfunctionIDWrongOrder",
    "e2err_RICsubscriptionFailureCauseWrongOrder",
};

typedef struct {
    uint64_t messageType; // Initiating message or Successful outcome or Unsuccessful outcome
    uint64_t messageId;
} E2MessageInfo_t;

//////////////////////////////////////////////////////////////////////
// Message definitons

// Below constant values are same as in E2AP, E2SM and X2AP specs
extern const uint64_t cE2InitiatingMessage;
extern const uint64_t cE2SuccessfulOutcome;
extern const uint64_t cE2UnsuccessfulOutcome;

// E2AP messages. Below message id constant values are the same as in ASN.1 specification

// Initiating message
extern const uint64_t cRICSubscriptionRequest;
extern const uint64_t cRICSubscriptionDeleteRequest;
extern const uint64_t cRICSubscriptionDeleteRequired;

// Successful outcome
extern const uint64_t cRICSubscriptionResponse;
extern const uint64_t cRICsubscriptionDeleteResponse;

// Unsuccessful outcome
extern const uint64_t cRICSubscriptionFailure;
extern const uint64_t cRICsubscriptionDeleteFailure;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICSubscriptionDetails_t ricSubscriptionDetails;
} RICSubscriptionRequest_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICActionAdmittedList_t ricActionAdmittedList;
    bool ricActionNotAdmittedListPresent;
    RICActionNotAdmittedList_t ricActionNotAdmittedList;
} RICSubscriptionResponse_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICCause_t cause;
    bool criticalityDiagnosticsPresent;
    CriticalityDiagnostics__t criticalityDiagnostics;
} RICSubscriptionFailure_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
} RICSubscriptionDeleteRequest_t;

typedef struct  {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
} RICSubscriptionDeleteResponse_t;

typedef struct  {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICCause_t cause;
    bool criticalityDiagnosticsPresent;
    CriticalityDiagnostics__t criticalityDiagnostics; // OPTIONAL. Not used in RIC currently
} RICSubscriptionDeleteFailure_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICCause_t cause;
} RICSubscriptionDeleteRequired_t;

typedef struct {
    int noOfRanSubscriptions;
    RICSubscriptionDeleteRequired_t ranSubscriptionsDelRequired[1024];

} RICSubsDeleteRequired_t;

//////////////////////////////////////////////////////////////////////
// Function declarations

void allowASN1DebugPrints(bool);

void allowOutOfOrderIEMsg(uint8_t);

const char* getE2ErrorString(uint64_t);

typedef void* e2ap_pdu_ptr_t;

uint64_t packRICSubscriptionRequest(size_t*, byte*, char*,RICSubscriptionRequest_t*);
uint64_t packRICSubscriptionResponse(size_t*, byte*, char*,RICSubscriptionResponse_t*);
uint64_t packRICSubscriptionFailure(size_t*, byte*, char*,RICSubscriptionFailure_t*);
uint64_t packRICSubscriptionDeleteRequest(size_t*, byte*, char*,RICSubscriptionDeleteRequest_t*);
uint64_t packRICSubscriptionDeleteResponse(size_t*, byte*, char*,RICSubscriptionDeleteResponse_t*);
uint64_t packRICSubscriptionDeleteFailure(size_t*, byte*, char*,RICSubscriptionDeleteFailure_t*);
uint64_t packRICSubscriptionDeleteRequired(size_t*, byte*, char*,RICSubsDeleteRequired_t*);

e2ap_pdu_ptr_t* unpackE2AP_pdu(const size_t, const byte*, char*, E2MessageInfo_t*);
uint64_t getRICSubscriptionRequestData(e2ap_pdu_ptr_t*, RICSubscriptionRequest_t*);
uint64_t getRICSubscriptionResponseData(e2ap_pdu_ptr_t*, RICSubscriptionResponse_t*);
uint64_t getRICSubscriptionFailureData(e2ap_pdu_ptr_t*, RICSubscriptionFailure_t*);
uint64_t getRICSubscriptionDeleteRequestData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteRequest_t*);
uint64_t getRICSubscriptionDeleteResponseData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteResponse_t*);
uint64_t getRICSubscriptionDeleteFailureData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteFailure_t*);
uint64_t getRICSubscriptionDeleteRequiredData(e2ap_pdu_ptr_t*, RICSubsDeleteRequired_t*);

#if DEBUG
bool TestRICSubscriptionRequest();
bool TestRICSubscriptionResponse();
bool TestRICSubscriptionFailure();
bool TestRICSubscriptionDeleteRequest();
bool TestRICSubscriptionDeleteResponse();
bool TestRICSubscriptionDeleteFailure();
bool TestRICSubscriptionDeleteRequired();

void printRICSubscriptionRequest(const RICSubscriptionRequest_t*);
void printRICSubscriptionResponse(const RICSubscriptionResponse_t*);
void printRICSubscriptionFailure(const RICSubscriptionFailure_t*);
void printRICSubscriptionDeleteRequest(const RICSubscriptionDeleteRequest_t*);
void printRICSubscriptionDeleteResponse(const RICSubscriptionDeleteResponse_t*);
void printRICSubscriptionDeleteFailure(const RICSubscriptionDeleteFailure_t*);
void printRICSubscriptionDeleteRequired(const RICSubsDeleteRequired_t*);
#endif

#ifdef	__cplusplus
}
#endif

#endif
