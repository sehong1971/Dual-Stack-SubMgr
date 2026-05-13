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

package e2ap_tests

import (
	"testing"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionRequestWithData(t *testing.T, areqenc *e2ap.E2APSubscriptionRequest) {

	e2SubsReq := testCtxt.packerif.NewPackerSubscriptionRequest()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")
	testCtxt.testPrint("pack")

	err, packedMsg := e2SubsReq.Pack(areqenc)
	if err != nil {
		testCtxt.testError(t, "Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsReq.String())
	testCtxt.testPrint("unpack")

	err, areqdec := e2SubsReq.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsReq.String())

	testCtxt.testValueEquality(t, "msg", areqenc, areqdec)
	testCtxt.testValueEquality(t, "EventTriggerDefinition", &areqenc.EventTriggerDefinition, &areqdec.EventTriggerDefinition)
}
func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionRequest(t *testing.T, msgContent *SubscriptionTestMsgContent) {

	areqenc := e2ap.E2APSubscriptionRequest{}
	areqenc.RequestId.Id = 1
	areqenc.RequestId.InstanceId = 22
	areqenc.FunctionId = 33

	if msgContent.NBX2EventTriggerDefinitionPresent {
		areqenc.EventTriggerDefinition.Data.Length = 1
		areqenc.EventTriggerDefinition.Data.Data = make([]uint8, areqenc.EventTriggerDefinition.Data.Length)
		areqenc.EventTriggerDefinition.Data.Data[0] = 1
	} else if msgContent.NBNRTEventTriggerDefinitionPresent {
		areqenc.EventTriggerDefinition.Data.Length = 1
		areqenc.EventTriggerDefinition.Data.Data = make([]uint8, areqenc.EventTriggerDefinition.Data.Length)
		areqenc.EventTriggerDefinition.Data.Data[0] = 100
	}

	for index := 0; index < 1; /*16*/ index++ {
		item := e2ap.ActionToBeSetupItem{}
		item.ActionId = uint64(index)
		item.ActionType = e2ap.E2AP_ActionTypeInsert

		item.RicActionDefinitionPresent = true

		if item.RicActionDefinitionPresent {
			if msgContent.ActionDefinitionX2Format1Present {
				item.ActionDefinitionChoice.Data.Length = 1
				item.ActionDefinitionChoice.Data.Data = make([]uint8, item.ActionDefinitionChoice.Data.Length)
				item.ActionDefinitionChoice.Data.Data[0] = 1
			} else if msgContent.ActionDefinitionX2Format2Present {
				item.ActionDefinitionChoice.Data.Length = 1
				item.ActionDefinitionChoice.Data.Data = make([]uint8, item.ActionDefinitionChoice.Data.Length)
				item.ActionDefinitionChoice.Data.Data[0] = 2
			} else if msgContent.ActionDefinitionNRTFormat1Present {
				item.ActionDefinitionChoice.Data.Length = 1
				item.ActionDefinitionChoice.Data.Data = make([]uint8, item.ActionDefinitionChoice.Data.Length)
				item.ActionDefinitionChoice.Data.Data[0] = 3
			}
		}
		item.SubsequentAction.Present = true
		item.SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
		item.SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitW100ms
		areqenc.ActionSetups = append(areqenc.ActionSetups, item)
	}
	testCtxt.E2ApTestMsgSubscriptionRequestWithData(t, &areqenc)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionResponse(t *testing.T) {

	testCtxt.SetDesc("SubsResp")

	e2SubsResp := testCtxt.packerif.NewPackerSubscriptionResponse()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	arespenc := e2ap.E2APSubscriptionResponse{}
	arespenc.RequestId.Id = 1
	arespenc.RequestId.InstanceId = 22
	arespenc.FunctionId = 33
	for index := uint64(0); index < 16; index++ {
		item := e2ap.ActionAdmittedItem{}
		item.ActionId = index
		arespenc.ActionAdmittedList.Items = append(arespenc.ActionAdmittedList.Items, item)
	}
	for index := uint64(0); index < 16; index++ {
		item := e2ap.ActionNotAdmittedItem{}
		item.ActionId = index
		item.Cause.Content = 1
		item.Cause.Value = 1
		arespenc.ActionNotAdmittedList.Items = append(arespenc.ActionNotAdmittedList.Items, item)
	}

	testCtxt.testPrint("pack")
	err, packedMsg := e2SubsResp.Pack(&arespenc)
	if err != nil {
		testCtxt.testError(t, "Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsResp.String())
	testCtxt.testPrint("unpack")
	err, arespdec := e2SubsResp.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsResp.String())
	testCtxt.testValueEquality(t, "msg", &arespenc, arespdec)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionFailure(t *testing.T) {

	testCtxt.SetDesc("SubsFail")

	e2SubsFail := testCtxt.packerif.NewPackerSubscriptionFailure()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	afailenc := e2ap.E2APSubscriptionFailure{}
	afailenc.RequestId.Id = 1
	afailenc.RequestId.InstanceId = 22
	afailenc.Cause.Content = e2ap.E2AP_CauseContent_RICrequest
	afailenc.Cause.Value = e2ap.E2AP_CauseValue_RICrequest_control_message_invalid

	// NOT SUPPORTED CURRENTLY
	afailenc.CriticalityDiagnostics.Present = false
	//	afailenc.CriticalityDiagnostics.ProcCodePresent = true
	//	afailenc.CriticalityDiagnostics.ProcCode = 1
	//	afailenc.CriticalityDiagnostics.TrigMsgPresent = true
	//	afailenc.CriticalityDiagnostics.TrigMsg = 2
	//	afailenc.CriticalityDiagnostics.ProcCritPresent = true
	//	afailenc.CriticalityDiagnostics.ProcCrit = e2ap.E2AP_CriticalityReject
	//	for index := uint32(0); index < 256; index++ {
	//		ieitem := e2ap.CriticalityDiagnosticsIEListItem{}
	//		ieitem.IeCriticality = e2ap.E2AP_CriticalityReject
	//		ieitem.IeID = index
	//		ieitem.TypeOfError = 1
	//		afailenc.CriticalityDiagnostics.CriticalityDiagnosticsIEList.Items = append(afailenc.CriticalityDiagnostics.CriticalityDiagnosticsIEList.Items, ieitem)
	//	}

	testCtxt.testPrint("pack")
	err, packedMsg := e2SubsFail.Pack(&afailenc)
	if err != nil {
		testCtxt.testError(t, "Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsFail.String())
	testCtxt.testPrint("unpack")
	err, afaildec := e2SubsFail.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsFail.String())
	testCtxt.testValueEquality(t, "msg", &afailenc, afaildec)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionRequestBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionRequest()
		err, _ := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubReqBuffer")
	testfunc("00c9402c000003ea7e00050000010000ea6300020001ea810016000b00130051407b000000054000ea6b000420000000")
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionResponseBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionResponse()
		err, _ := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubRespBuffer")
	testfunc("20c9402a000004ea7e00050000018009ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9401d000003ea7e0005004eec0004ea6300020001ea6c000700ea6d40020000")

}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionFailureBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionFailure()
		err, _ := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubFailBuffer")
	testfunc("40c94017000003ea7e000500000106f3ea6300020001ea6e000100")
}
