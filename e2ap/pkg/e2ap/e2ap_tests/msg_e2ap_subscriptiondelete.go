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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteRequest(t *testing.T) {

	testCtxt.SetDesc("SubsDeleteReq")

	e2SubsReq := testCtxt.packerif.NewPackerSubscriptionDeleteRequest()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	areqenc := e2ap.E2APSubscriptionDeleteRequest{}
	areqenc.RequestId.Id = 1
	areqenc.RequestId.InstanceId = 22
	areqenc.FunctionId = 33

	testCtxt.testPrint("pack")
	err, packedMsg := e2SubsReq.Pack(&areqenc)
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
	testCtxt.testValueEquality(t, "msg", &areqenc, areqdec)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteResponse(t *testing.T) {

	testCtxt.SetDesc("SubsDeleteResp")

	e2SubsResp := testCtxt.packerif.NewPackerSubscriptionDeleteResponse()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	arespenc := e2ap.E2APSubscriptionDeleteResponse{}
	arespenc.RequestId.Id = 1
	arespenc.RequestId.InstanceId = 22
	arespenc.FunctionId = 33

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

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteFailure(t *testing.T) {

	testCtxt.SetDesc("SubsDeleteFail")

	e2SubsFail := testCtxt.packerif.NewPackerSubscriptionDeleteFailure()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	afailenc := e2ap.E2APSubscriptionDeleteFailure{}
	afailenc.RequestId.Id = 1
	afailenc.RequestId.InstanceId = 22
	afailenc.FunctionId = 33
	afailenc.Cause.Content = 1
	afailenc.Cause.Value = 1
	// NOT SUPPORTED CURRENTLY
	//	afailenc.CriticalityDiagnostics.Present = false
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

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteRequestBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionDeleteRequest()
		err, _ := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubDelReqBuffer")

	testfunc("00ca4012000002ea7e000500000106e7ea6300020001")
	testfunc("00ca4012000002ea7e000500000106e8ea6300020001")
	testfunc("00ca4012000002ea7e000500000106e9ea6300020001")
	testfunc("00ca4012000002ea7e000500000106eaea6300020001")
	testfunc("00ca4012000002ea7e000500000106ebea6300020001")
	testfunc("00ca4012000002ea7e000500000106ecea6300020001")
	testfunc("00ca4012000002ea7e000500000106edea6300020001")
	testfunc("00ca4012000002ea7e000500000106eeea6300020001")
	testfunc("00ca4012000002ea7e000500000106efea6300020001")
	testfunc("00ca4012000002ea7e000500000106f0ea6300020001")
	testfunc("00ca4012000002ea7e000500000106f4ea6300020001")
	testfunc("00ca4012000002ea7e000500000106f5ea6300020001")
	testfunc("00ca4012000002ea7e000500000106f6ea6300020001")
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteResponseBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionDeleteResponse()
		err, _ := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubDelRespBuffer")
	testfunc("20ca4012000002ea7e000500000106e7ea6300020001")

}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteFailureBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionDeleteFailure()
		err, _ := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubDelFailBuffer")
	testfunc("40ca4017000003ea7e000500000106f6ea6300020001ea74000124")
}
