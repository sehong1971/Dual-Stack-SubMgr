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
	"testing"
)

func TestAddAllCountersOnce(t *testing.T) {

	// Check that all counters can be added correctly
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubFailToXapp, 1},
		Counter{cRestSubReqFromXapp, 1},
		Counter{cRestSubRespToXapp, 1},
		Counter{cRestSubFailToXapp, 1},
		Counter{cRestReqRejDueE2Down, 1},
		Counter{cRestSubNotifToXapp, 1},
		Counter{cRestSubFailNotifToXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cPartialSubRespFromE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cSubReqTimerExpiry, 1},
		Counter{cRouteCreateFail, 1},
		Counter{cRouteCreateUpdateFail, 1},
		Counter{cMergedSubscriptions, 1},
		Counter{cDuplicateE2SubReq, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelRespToXapp, 1},
		Counter{cRestSubDelReqFromXapp, 1},
		Counter{cRestSubDelRespToXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cSubDelReqTimerExpiry, 1},
		Counter{cSubDelRequFromE2, 1},
		Counter{cRouteDeleteFail, 1},
		Counter{cRouteDeleteUpdateFail, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cSDLWriteFailure, 1},
		Counter{cSDLReadFailure, 1},
		Counter{cSDLRemoveFailure, 1},
		Counter{cE2StateChangedToUp, 1},
		Counter{cE2StateChangedToDown, 1},
		Counter{cE2StateUnderReset, 1},
	})

	mainCtrl.c.UpdateCounter(cSubReqFromXapp)
	mainCtrl.c.UpdateCounter(cSubRespToXapp)
	mainCtrl.c.UpdateCounter(cSubFailToXapp)
	mainCtrl.c.UpdateCounter(cRestSubReqFromXapp)
	mainCtrl.c.UpdateCounter(cRestSubRespToXapp)
	mainCtrl.c.UpdateCounter(cRestSubFailToXapp)
	mainCtrl.c.UpdateCounter(cRestReqRejDueE2Down)
	mainCtrl.c.UpdateCounter(cRestSubNotifToXapp)
	mainCtrl.c.UpdateCounter(cRestSubFailNotifToXapp)
	mainCtrl.c.UpdateCounter(cSubReqToE2)
	mainCtrl.c.UpdateCounter(cSubReReqToE2)
	mainCtrl.c.UpdateCounter(cSubRespFromE2)
	mainCtrl.c.UpdateCounter(cPartialSubRespFromE2)
	mainCtrl.c.UpdateCounter(cSubFailFromE2)
	mainCtrl.c.UpdateCounter(cSubReqTimerExpiry)
	mainCtrl.c.UpdateCounter(cRouteCreateFail)
	mainCtrl.c.UpdateCounter(cRouteCreateUpdateFail)
	mainCtrl.c.UpdateCounter(cMergedSubscriptions)
	mainCtrl.c.UpdateCounter(cDuplicateE2SubReq)
	mainCtrl.c.UpdateCounter(cSubDelReqFromXapp)
	mainCtrl.c.UpdateCounter(cSubDelRespToXapp)
	mainCtrl.c.UpdateCounter(cRestSubDelReqFromXapp)
	mainCtrl.c.UpdateCounter(cRestSubDelRespToXapp)
	mainCtrl.c.UpdateCounter(cSubDelReqToE2)
	mainCtrl.c.UpdateCounter(cSubDelReReqToE2)
	mainCtrl.c.UpdateCounter(cSubDelRespFromE2)
	mainCtrl.c.UpdateCounter(cSubDelFailFromE2)
	mainCtrl.c.UpdateCounter(cSubDelReqTimerExpiry)
	mainCtrl.c.UpdateCounter(cSubDelRequFromE2)
	mainCtrl.c.UpdateCounter(cRouteDeleteFail)
	mainCtrl.c.UpdateCounter(cRouteDeleteUpdateFail)
	mainCtrl.c.UpdateCounter(cUnmergedSubscriptions)
	mainCtrl.c.UpdateCounter(cSDLWriteFailure)
	mainCtrl.c.UpdateCounter(cSDLReadFailure)
	mainCtrl.c.UpdateCounter(cSDLRemoveFailure)
	mainCtrl.c.UpdateCounter(cE2StateChangedToUp)
	mainCtrl.c.UpdateCounter(cE2StateChangedToDown)
	mainCtrl.c.UpdateCounter(cE2StateUnderReset)

	mainCtrl.VerifyCounterValues(t)
}
