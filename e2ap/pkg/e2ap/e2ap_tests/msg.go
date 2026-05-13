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
	"encoding/hex"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"github.com/google/go-cmp/cmp"
	"log"
	"os"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var testLogger *log.Logger

func init() {
	testLogger = log.New(os.Stdout, "TEST: ", log.LstdFlags)
}

type ApTests struct {
	name string
	desc string
}

func (testctxt *ApTests) Name() string { return testctxt.name }

func (testctxt *ApTests) Desc() string { return testctxt.desc }

func (testctxt *ApTests) SetDesc(desc string) { testctxt.desc = desc }

func (testctxt *ApTests) String() string { return testctxt.name + string("-") + testctxt.desc }

func (testctxt *ApTests) testPrint(pattern string, args ...interface{}) {
	testLogger.Printf("(%s): %s", testctxt.String(), fmt.Sprintf(pattern, args...))
}

func (testctxt *ApTests) testError(t *testing.T, pattern string, args ...interface{}) {
	testLogger.Printf("(%s): %s", testctxt.String(), fmt.Sprintf(pattern, args...))
	t.Errorf("(%s): %s", testctxt.String(), fmt.Sprintf(pattern, args...))
}

func (testctxt *ApTests) testValueEquality(t *testing.T, msg string, a interface{}, b interface{}) {
	if !cmp.Equal(a, b) {
		testLogger.Printf("(%s) %s Difference: %s", testctxt.String(), msg, cmp.Diff(a, b))
		t.Errorf("(%s) %s Difference: %s", testctxt.String(), msg, cmp.Diff(a, b))
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type E2ApTests struct {
	ApTests
	packerif e2ap.E2APPackerIf
}

func (testCtxt *E2ApTests) toPackedData(t *testing.T, buffer string) *e2ap.PackedData {
	msg, err := hex.DecodeString(buffer)
	if err != nil {
		testCtxt.testError(t, "Hex DecodeString Failed: %s [%s]", err.Error(), buffer)
		return nil
	}
	packedData := &e2ap.PackedData{}
	packedData.Buf = msg
	return packedData
}

func NewE2ApTests(name string, packerif e2ap.E2APPackerIf) *E2ApTests {
	testCtxt := &E2ApTests{}
	testCtxt.packerif = packerif
	testCtxt.name = name
	return testCtxt
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func RunTests(t *testing.T, e2aptestctxt *E2ApTests) {

	subMsgContent := &SubscriptionTestMsgContent{}
	subMsgContent.NBNRTEventTriggerDefinitionPresent = true
	subMsgContent.ActionDefinitionNRTFormat1Present = true
	subMsgContent.RANParameterValueEnumPresent = true
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionRequest(t, subMsgContent) })

	subMsgContent2 := &SubscriptionTestMsgContent{}
	subMsgContent2.NBX2EventTriggerDefinitionPresent = true
	subMsgContent2.ActionDefinitionX2Format1Present = true
	subMsgContent2.ActionParameterValueBoolPresent = true
	subMsgContent2.RANParameterValueBoolPresent = true
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionRequest(t, subMsgContent2) })

	subMsgContent3 := &SubscriptionTestMsgContent{}
	subMsgContent3.NBX2EventTriggerDefinitionPresent = true
	subMsgContent3.ActionDefinitionX2Format2Present = true
	subMsgContent3.ActionParameterValueBitSPresent = true
	subMsgContent3.RANParameterValueBitSPresent = true
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionRequest(t, subMsgContent3) })

	subMsgContent4 := &SubscriptionTestMsgContent{}
	subMsgContent4.NBX2EventTriggerDefinitionPresent = true
	subMsgContent4.ActionDefinitionX2Format2Present = true
	subMsgContent4.ActionParameterValueOctSPresent = true
	subMsgContent4.RANParameterValueOctSPresent = true
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionRequest(t, subMsgContent4) })

	subMsgContent5 := &SubscriptionTestMsgContent{}
	subMsgContent5.NBX2EventTriggerDefinitionPresent = true
	subMsgContent5.ActionDefinitionX2Format2Present = true
	subMsgContent5.ActionParameterValuePrtSPresent = true
	subMsgContent5.RANParameterValuePrtSPresent = true
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionRequest(t, subMsgContent5) })

	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionResponse(t) })
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionFailure(t) })
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionDeleteRequest(t) })
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionDeleteResponse(t) })
	t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionDeleteFailure(t) })

	/*
		t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionRequestBuffers(t) })
		t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionResponseBuffers(t) })
		t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionFailureBuffers(t) })
		t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionDeleteRequestBuffers(t) })
		t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionDeleteResponseBuffers(t) })
		t.Run(e2aptestctxt.Name(), func(t *testing.T) { e2aptestctxt.E2ApTestMsgSubscriptionDeleteFailureBuffers(t) })
	*/

}
