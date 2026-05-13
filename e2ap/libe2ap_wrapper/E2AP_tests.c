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


#if DEBUG

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "Cause.h"
#include "E2AP_if.h"

const size_t cDataBufferSize = 2048;

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionRequest() {

    RICSubscriptionRequest_t ricSubscriptionRequest;

    // RICrequestID
    ricSubscriptionRequest.ricRequestID.ricRequestorID = 20206;
    ricSubscriptionRequest.ricRequestID.ricInstanceID = 0;

    // RANfunctionID
    ricSubscriptionRequest.ranFunctionID = 3;

    // RICsubscriptionDetails
    // RICeventTriggerDefinition
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength = 4;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data[0] = 11;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data[1] = 22;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data[2] = 33;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data[3] = 44;

    // RICactions-ToBeSetup-List
    ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = 16;  //1..16
    uint64_t index = 0;
    while (index < ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength) {

        // RICactionID
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID = index;

        // RICactionType
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType = RICActionType_report; //RICActionType_policy;

        // RICactionDefinition, OPTIONAL.
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = true;  // E2AP
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength = 2;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.data[0] = 1;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.data[1] = 2;

        // RICsubsequentAction, OPTIONAL
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = true;  // E2AP
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType = RICSubsequentActionType_Continue;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait = RICTimeToWait_zero;
        index++;
    }

    printRICSubscriptionRequest(&ricSubscriptionRequest);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionRequest(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionRequest) == e2err_OK)
    {
        memset(&ricSubscriptionRequest,0, sizeof (RICSubscriptionRequest_t));
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2InitiatingMessage) {
                if (messageInfo.messageId == cRICSubscriptionRequest) {
                    if ((returnCode = getRICSubscriptionRequestData(pE2AP_PDU, &ricSubscriptionRequest)) == e2err_OK) {
                        printRICSubscriptionRequest(&ricSubscriptionRequest);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionRequestData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionRequest\n");
            }
            else
                printf("Not InitiatingMessage\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);

    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionResponse() {

    RICSubscriptionResponse_t ricSubscriptionResponse;

    // RICrequestID
    ricSubscriptionResponse.ricRequestID.ricRequestorID = 1;
    ricSubscriptionResponse.ricRequestID.ricInstanceID = 22;

    // RANfunctionID
    ricSubscriptionResponse.ranFunctionID = 33;

    // RICaction-Admitted-List
    ricSubscriptionResponse.ricActionAdmittedList.contentLength = 16;
    uint64_t index = 0;
    while (index < ricSubscriptionResponse.ricActionAdmittedList.contentLength) {

        // RICactionID
        ricSubscriptionResponse.ricActionAdmittedList.ricActionID[index] = index;
        index++;
    }
    // RICaction-NotAdmitted-List, OPTIONAL
    ricSubscriptionResponse.ricActionNotAdmittedListPresent = true;
    ricSubscriptionResponse.ricActionNotAdmittedList.contentLength = 16;
    index = 0;
    while (index < ricSubscriptionResponse.ricActionNotAdmittedList.contentLength) {

        // RICactionID
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = index;  //0..255

        // Cause
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_ricService; //cCauseRICService;
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal = 2;
        index++;
    }

    printRICSubscriptionResponse(&ricSubscriptionResponse);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionResponse(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionResponse) == e2err_OK)
    {
        memset(&ricSubscriptionResponse,0, sizeof ricSubscriptionResponse);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2SuccessfulOutcome) {
                if (messageInfo.messageId == cRICSubscriptionResponse) {
                    if ((returnCode = getRICSubscriptionResponseData(pE2AP_PDU, &ricSubscriptionResponse)) == e2err_OK) {
                        printRICSubscriptionResponse(&ricSubscriptionResponse);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionResponseData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionResponse\n");
            }
            else
                printf("Not SuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionFailure() {

    RICSubscriptionFailure_t ricSubscriptionFailure;

    // RICrequestID
    ricSubscriptionFailure.ricRequestID.ricRequestorID = 1;
    ricSubscriptionFailure.ricRequestID.ricInstanceID = 22;

    // RANfunctionID
    ricSubscriptionFailure.ranFunctionID = 33;

    // Cause
    ricSubscriptionFailure.cause.content = Cause_PR_ricService;
    ricSubscriptionFailure.cause.causeVal =

    // CriticalityDiagnostics, OPTIONAL. Not used in RIC
    ricSubscriptionFailure.criticalityDiagnosticsPresent = false;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCodePresent = true;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCode = 1;
    ricSubscriptionFailure.criticalityDiagnostics.triggeringMessagePresent = true;
    ricSubscriptionFailure.criticalityDiagnostics.triggeringMessage = TriggeringMessage__initiating_message;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCriticalityPresent = true;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCriticality = Criticality__reject;

    ricSubscriptionFailure.criticalityDiagnostics.iEsCriticalityDiagnosticsPresent = false;
    ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIELength = 256;
    uint16_t index2 = 0;
    while (index2 < ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIELength) {
        ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iECriticality = Criticality__reject;
        ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iE_ID = index2;
        ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].typeOfError = TypeOfError_missing;
        index2++;
    }

    printRICSubscriptionFailure(&ricSubscriptionFailure);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionFailure(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionFailure) == e2err_OK)
    {
        memset(&ricSubscriptionFailure,0, sizeof ricSubscriptionFailure);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2UnsuccessfulOutcome) {
                if (messageInfo.messageId == cRICSubscriptionFailure) {
                    if ((returnCode = getRICSubscriptionFailureData(pE2AP_PDU, &ricSubscriptionFailure)) == e2err_OK) {
                        printRICSubscriptionFailure(&ricSubscriptionFailure);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionFailureData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionFailure\n");
            }
            else
                printf("Not UnuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionDeleteRequest() {

    RICSubscriptionDeleteRequest_t ricSubscriptionDeleteRequest;

    // RICrequestID
    ricSubscriptionDeleteRequest.ricRequestID.ricRequestorID = 1;
    ricSubscriptionDeleteRequest.ricRequestID.ricInstanceID = 22;

     // RANfunctionID
   ricSubscriptionDeleteRequest.ranFunctionID = 33;

    printRICSubscriptionDeleteRequest(&ricSubscriptionDeleteRequest);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[cDataBufferSize];
    if ((packRICSubscriptionDeleteRequest(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionDeleteRequest)) == e2err_OK)
    {
        memset(&ricSubscriptionDeleteRequest,0, sizeof ricSubscriptionDeleteRequest);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2InitiatingMessage) {
                if (messageInfo.messageId == cRICSubscriptionDeleteRequest) {
                    if ((returnCode = getRICSubscriptionDeleteRequestData(pE2AP_PDU, &ricSubscriptionDeleteRequest)) == e2err_OK) {
                        printRICSubscriptionDeleteRequest(&ricSubscriptionDeleteRequest);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionDeleteRequestData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionDeleteRequest\n");
            }
            else
                printf("Not InitiatingMessage\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionDeleteResponse() {

    RICSubscriptionDeleteResponse_t ricSubscriptionDeleteResponse;

    // RICrequestID
    ricSubscriptionDeleteResponse.ricRequestID.ricRequestorID = 1;
    ricSubscriptionDeleteResponse.ricRequestID.ricInstanceID = 22;

    // RANfunctionID
    ricSubscriptionDeleteResponse.ranFunctionID = 33;

    printRICSubscriptionDeleteResponse(&ricSubscriptionDeleteResponse);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if ((packRICSubscriptionDeleteResponse(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionDeleteResponse)) == e2err_OK)
    {
        memset(&ricSubscriptionDeleteResponse,0, sizeof ricSubscriptionDeleteResponse);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2SuccessfulOutcome) {
                if (messageInfo.messageId == cRICsubscriptionDeleteResponse) {
                    if ((returnCode = getRICSubscriptionDeleteResponseData(pE2AP_PDU, &ricSubscriptionDeleteResponse)) == e2err_OK) {
                        printRICSubscriptionDeleteResponse(&ricSubscriptionDeleteResponse);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionDeleteResponseData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionDeleteResponse\n");
            }
            else
                printf("Not SuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionDeleteFailure() {

    RICSubscriptionDeleteFailure_t ricSubscriptionDeleteFailure;

    // RICrequestID
    ricSubscriptionDeleteFailure.ricRequestID.ricRequestorID = 1;
    ricSubscriptionDeleteFailure.ricRequestID.ricInstanceID = 22;

    // RANfunctionID
    ricSubscriptionDeleteFailure.ranFunctionID = 33;

    // Cause
    ricSubscriptionDeleteFailure.cause.content = Cause_PR_ricService;//cCauseRICService;
    ricSubscriptionDeleteFailure.cause.causeVal = 2;

    printRICSubscriptionDeleteFailure(&ricSubscriptionDeleteFailure);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if ((packRICSubscriptionDeleteFailure(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionDeleteFailure)) == e2err_OK)
    {
        memset(&ricSubscriptionDeleteFailure,0, sizeof ricSubscriptionDeleteFailure);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2UnsuccessfulOutcome) {
                if (messageInfo.messageId == cRICsubscriptionDeleteFailure) {
                    if ((returnCode = getRICSubscriptionDeleteFailureData(pE2AP_PDU, &ricSubscriptionDeleteFailure)) == e2err_OK) {
                        printRICSubscriptionDeleteFailure(&ricSubscriptionDeleteFailure);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionDeleteFailureData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionDeleteFailure\n");
            }
            else
                printf("Not UnuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
void printDataBuffer(const size_t byteCount, const uint8_t* pData) {

    uint64_t index = 0;
    while (index < byteCount) {
        if (index > 0 && index % 50 == 0) {
            printf("\n");
        }
        printf("%u ",pData[index]);
        index++;
    }
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionRequest(const RICSubscriptionRequest_t* pRICSubscriptionRequest) {
    printf("pRICSubscriptionRequest->ricRequestID.ricRequestorID = %u\n", pRICSubscriptionRequest->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionRequest->ricRequestID.ricInstanceID = %u\n", pRICSubscriptionRequest->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionRequest->ranFunctionID = %u\n",pRICSubscriptionRequest->ranFunctionID);

    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength = %li\n",
         pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength);
    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data = ");
    printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength,
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data);
    printf("\n");

    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength);

    uint64_t index = 0;
    while (index < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength) {
        printf("index = %lu\n", index);
        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID = %li\n",
             pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID);
        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType = %li\n",
             pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType);
        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = %i\n",
             pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent);
        if(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent)
        {
            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength = %li\n",
                 pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength);
            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.data = ");
            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength,
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.data);
            printf("\n");
        }

        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = %i\n",
          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent);
        if(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent)
        {
            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType = %li\n",
                 pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType);
            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait = %li\n",
                 pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait);
        }
        printf("\n\n");
        index++;
    }
    printf("\n\n");
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionResponse(const RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    printf("pRICSubscriptionResponse->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionResponse->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionResponse->ricRequestID.ricInstanceID = %u\n", pRICSubscriptionResponse->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionResponse->ranFunctionID = %u\n",pRICSubscriptionResponse->ranFunctionID);
    printf("pRICSubscriptionResponse->ricActionAdmittedList.contentLength = %u\n",(unsigned)pRICSubscriptionResponse->ricActionAdmittedList.contentLength);
    uint64_t index = 0;
    while (index < pRICSubscriptionResponse->ricActionAdmittedList.contentLength) {
        printf("pRICSubscriptionResponse->ricActionAdmittedList.ricActionID[index] = %lu\n",pRICSubscriptionResponse->ricActionAdmittedList.ricActionID[index]);
        index++;
    }
    printf("pRICSubscriptionResponse->ricActionNotAdmittedListPresent = %u\n",pRICSubscriptionResponse->ricActionNotAdmittedListPresent);
    printf("pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength = %u\n",(unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength);
    index = 0;
    while (index < pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength) {
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = %lu\n",
             pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID);
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = %u\n",
             (unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content);
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.cause = %u\n",
             (unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal);
        index++;
    }
    printf("\n");
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionFailure(const RICSubscriptionFailure_t* pRICSubscriptionFailure) {

    printf("pRICSubscriptionFailure->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionFailure->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionFailure->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionFailure->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionFailure->ranFunctionID = %i\n",pRICSubscriptionFailure->ranFunctionID);
    printf("pRICSubscriptionFailure->ricActionNotAdmittedList.content = %u\n",pRICSubscriptionFailure->cause.content);
    printf("pRICSubscriptionFailure->ricActionNotAdmittedList.content = %u\n",pRICSubscriptionFailure->cause.causeVal);
    if (pRICSubscriptionFailure->criticalityDiagnosticsPresent) {
        printf("pRICSubscriptionFailure->criticalityDiagnosticsPresent = %u\n",pRICSubscriptionFailure->criticalityDiagnosticsPresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCodePresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.procedureCodePresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCode = %u\n",(unsigned)pRICSubscriptionFailure->criticalityDiagnostics.procedureCode);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessagePresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessagePresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessage = %u\n",(unsigned)pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessage);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticalityPresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticalityPresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticality = %u\n",(unsigned)pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticality);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.iEsCriticalityDiagnosticsPresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.iEsCriticalityDiagnosticsPresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIELength = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIELength);
        unsigned int index = 0;
        while (index < pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIELength) {
            printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iECriticality = %u\n",
                 (unsigned)pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iECriticality);
            printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iE_ID = %u\n",
                 pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iE_ID);
            printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].typeOfError = %u\n",
                 (unsigned)pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].typeOfError);
            index++;
        }
    }
    printf("\n");
}

void printRICSubscriptionDeleteRequest(const RICSubscriptionDeleteRequest_t* pRICSubscriptionDeleteRequest) {

    printf("\npRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteRequest->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionDeleteRequest->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionDeleteRequest->ranFunctionID = %i\n",pRICSubscriptionDeleteRequest->ranFunctionID);
    printf("\n");
}

void printRICSubscriptionDeleteResponse(const RICSubscriptionDeleteResponse_t* pRICSubscriptionDeleteResponse) {

    printf("\npRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteResponse->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionDeleteResponse->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionDeleteResponse->ranFunctionID = %i\n",pRICSubscriptionDeleteResponse->ranFunctionID);
    printf("\n");
}

void printRICSubscriptionDeleteFailure(const RICSubscriptionDeleteFailure_t* pRICSubscriptionDeleteFailure) {

    printf("\npRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteFailure->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionDeleteFailure->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionDeleteFailure->ranFunctionID = %i\n",pRICSubscriptionDeleteFailure->ranFunctionID);
    printf("pRICSubscriptionDeleteFailure->ricCause.content = %i\n",pRICSubscriptionDeleteFailure->cause.content);
    printf("pRICSubscriptionDeleteFailure->ricCause.cause = %i\n",pRICSubscriptionDeleteFailure->cause.causeVal);
    printf("\n");
}

#endif
