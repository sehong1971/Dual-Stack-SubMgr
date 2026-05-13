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
	"fmt"
	"strconv"
	"strings"
	"time"

	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	rtmgrhandle "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client/handle"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type SubRouteInfo struct {
	EpList xapp.RmrEndpointList
	SubID  uint16
}

func (sri *SubRouteInfo) String() string {
	return "routeinfo(" + strconv.FormatUint(uint64(sri.SubID), 10) + "/[" + sri.EpList.String() + "])"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RtmgrClient struct {
	rtClient *rtmgrclient.RoutingManager
}

func (rc *RtmgrClient) SubscriptionRequestCreate(subRouteAction SubRouteInfo) error {
	subID := int32(subRouteAction.SubID)
	xapp.Logger.Debug("CREATE %s ongoing", subRouteAction.String())
	createData := rtmgr_models.XappSubscriptionData{&subRouteAction.EpList.Endpoints[0].Addr, &subRouteAction.EpList.Endpoints[0].Port, &subID}
	createHandle := rtmgrhandle.NewProvideXappSubscriptionHandleParamsWithTimeout(2 * time.Second)
	createHandle.WithXappSubscriptionData(&createData)
	_, err := rc.rtClient.Handle.ProvideXappSubscriptionHandle(createHandle)
	if err != nil && !(strings.Contains(err.Error(), "status 200")) {
		return fmt.Errorf("CREATE %s failed with error: %s", subRouteAction.String(), err.Error())
	}
	xapp.Logger.Debug("CREATE %s successful", subRouteAction.String())
	return nil
}

func (rc *RtmgrClient) SubscriptionRequestUpdate(subRouteAction SubRouteInfo) error {
	xapp.Logger.Debug("UPDATE %s ongoing", subRouteAction.String())
	var updateData rtmgr_models.XappList
	for i := range subRouteAction.EpList.Endpoints {
		updateData = append(updateData, &rtmgr_models.XappElement{Address: &subRouteAction.EpList.Endpoints[i].Addr, Port: &subRouteAction.EpList.Endpoints[i].Port})
	}
	updateHandle := rtmgrhandle.NewUpdateXappSubscriptionHandleParamsWithTimeout(2 * time.Second)
	updateHandle.WithSubscriptionID(subRouteAction.SubID)
	updateHandle.WithXappList(updateData)
	_, err := rc.rtClient.Handle.UpdateXappSubscriptionHandle(updateHandle)
	if err != nil && !(strings.Contains(err.Error(), "status 200")) {
		return fmt.Errorf("UPDATE %s failed with error: %s", subRouteAction.String(), err.Error())
	}
	xapp.Logger.Debug("UPDATE %s successful", subRouteAction.String())
	return nil

}

func (rc *RtmgrClient) SubscriptionRequestDelete(subRouteAction SubRouteInfo) error {
	subID := int32(subRouteAction.SubID)
	xapp.Logger.Debug("DELETE %s ongoing", subRouteAction.String())
	deleteData := rtmgr_models.XappSubscriptionData{&subRouteAction.EpList.Endpoints[0].Addr, &subRouteAction.EpList.Endpoints[0].Port, &subID}
	deleteHandle := rtmgrhandle.NewDeleteXappSubscriptionHandleParamsWithTimeout(2 * time.Second)
	deleteHandle.WithXappSubscriptionData(&deleteData)
	_, _, err := rc.rtClient.Handle.DeleteXappSubscriptionHandle(deleteHandle)
	if err != nil && !(strings.Contains(err.Error(), "status 200")) {
		return fmt.Errorf("DELETE %s failed with error: %s", subRouteAction.String(), err.Error())
	}
	xapp.Logger.Debug("DELETE %s successful", subRouteAction.String())
	return nil
}
