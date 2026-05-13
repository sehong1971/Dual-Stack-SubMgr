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
	"encoding/json"
	"fmt"

	sdl "gerrit.o-ran-sc.org/r/ric-plt/sdlgo"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

const restSubSdlNs = "submgr_restSubsDb"

type RESTSubscriptionInfo struct {
	Created          string
	XAppServiceName  string
	XAppRmrEndPoint  string
	Meid             string
	InstanceIds      []uint32
	XAppIdToE2Id     map[int64]int64
	SubReqOngoing    bool
	SubDelReqOngoing bool
	Md5sum           string
}

func CreateRESTSdl() Sdlnterface {
	return sdl.NewSyncStorage()
}

func (c *Control) WriteRESTSubscriptionToSdl(restSubId string, restSubs *RESTSubscription) error {

	var restSubscriptionInfo RESTSubscriptionInfo
	restSubscriptionInfo.Created = restSubs.Created
	restSubscriptionInfo.XAppServiceName = restSubs.xAppServiceName
	restSubscriptionInfo.XAppRmrEndPoint = restSubs.xAppRmrEndPoint
	restSubscriptionInfo.Meid = restSubs.Meid
	restSubscriptionInfo.InstanceIds = restSubs.InstanceIds
	restSubscriptionInfo.XAppIdToE2Id = restSubs.xAppIdToE2Id
	restSubscriptionInfo.SubReqOngoing = restSubs.SubReqOngoing
	restSubscriptionInfo.SubDelReqOngoing = restSubs.SubDelReqOngoing
	restSubscriptionInfo.Md5sum = restSubs.lastReqMd5sum

	jsonData, err := json.Marshal(restSubscriptionInfo)
	if err != nil {
		return fmt.Errorf("SDL: WriteSubscriptionToSdl() json.Marshal error: %s", err.Error())
	}

	if err = c.restSubsDb.Set(restSubSdlNs, restSubId, jsonData); err != nil {
		c.UpdateCounter(cSDLWriteFailure)
		return fmt.Errorf("SDL: WriteSubscriptionToSdl(): %s", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription written in restSubsDb. restSubId = %v", restSubId)
	}
	return nil
}

func (c *Control) ReadRESTSubscriptionFromSdl(restSubId string) (*RESTSubscription, error) {

	// This function is now just for testing purpose
	key := restSubId
	retMap, err := c.restSubsDb.Get(restSubSdlNs, []string{key})
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl(): %s", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription read from restSubsDb.  restSubId = %v", restSubId)
	}

	restSubs := &RESTSubscription{}
	for _, iRESTSubscriptionInfo := range retMap {

		if iRESTSubscriptionInfo == nil {
			return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl() REST subscription not found. restSubId = %v\n", restSubId)
		}

		restSubscriptionInfo := &RESTSubscriptionInfo{}
		jsonSubscriptionInfo := iRESTSubscriptionInfo.(string)

		if err := json.Unmarshal([]byte(jsonSubscriptionInfo), restSubscriptionInfo); err != nil {
			return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl() json.unmarshal error: %s\n", err.Error())
		}

		restSubs = c.CreateRESTSubscription(restSubscriptionInfo, &jsonSubscriptionInfo)

		c.restDuplicateCtrl.SetMd5sumFromLastOkRequest(restSubId, restSubs.lastReqMd5sum)
	}
	return restSubs, nil
}

func (c *Control) CreateRESTSubscription(restSubscriptionInfo *RESTSubscriptionInfo, jsonSubscriptionInfo *string) *RESTSubscription {

	restSubs := &RESTSubscription{}
	restSubs.Created = restSubscriptionInfo.Created
	restSubs.xAppServiceName = restSubscriptionInfo.XAppServiceName
	restSubs.xAppRmrEndPoint = restSubscriptionInfo.XAppRmrEndPoint
	restSubs.Meid = restSubscriptionInfo.Meid
	restSubs.InstanceIds = restSubscriptionInfo.InstanceIds
	restSubs.xAppIdToE2Id = restSubscriptionInfo.XAppIdToE2Id
	restSubs.SubReqOngoing = restSubscriptionInfo.SubReqOngoing
	restSubs.SubDelReqOngoing = restSubscriptionInfo.SubDelReqOngoing
	restSubs.lastReqMd5sum = restSubscriptionInfo.Md5sum

	return restSubs
}

func (c *Control) RemoveRESTSubscriptionFromSdl(restSubId string) error {

	key := restSubId
	if err := c.restSubsDb.Remove(restSubSdlNs, []string{key}); err != nil {
		return fmt.Errorf("SDL: RemoveSubscriptionfromSdl(): %s\n", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription removed from restSubsDb. restSubId = %v", restSubId)
	}
	return nil
}

func (c *Control) ReadAllRESTSubscriptionsFromSdl() (map[string]*RESTSubscription, error) {

	retMap := make(map[string]*RESTSubscription)
	// Get all keys
	keys, err := c.restSubsDb.GetAll(restSubSdlNs)
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl(), GetAll(). Error while reading REST subscriptions keys from DBAAS %s\n", err.Error())
	}

	if len(keys) == 0 {
		return retMap, nil
	}

	// Get all subscriptionInfos
	iRESTSubscriptionMap, err := c.restSubsDb.Get(restSubSdlNs, keys)
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl(), Get():  Error while reading REST subscriptions from DBAAS %s\n", err.Error())
	}

	for iRESTSubId, iRESTSubscriptionInfo := range iRESTSubscriptionMap {

		if iRESTSubscriptionInfo == nil {
			return nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() iRESTSubscriptionInfo = nil\n")
		}

		restSubscriptionInfo := &RESTSubscriptionInfo{}
		jsonSubscriptionInfo := iRESTSubscriptionInfo.(string)

		if err := json.Unmarshal([]byte(jsonSubscriptionInfo), restSubscriptionInfo); err != nil {
			return nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() json.unmarshal error: %s\n", err.Error())
		}

		restSubs := c.CreateRESTSubscription(restSubscriptionInfo, &jsonSubscriptionInfo)
		retMap[iRESTSubId] = restSubs
	}
	return retMap, nil
}

func (c *Control) RemoveAllRESTSubscriptionsFromSdl() error {

	if err := c.restSubsDb.RemoveAll(restSubSdlNs); err != nil {
		c.UpdateCounter(cSDLRemoveFailure)
		return fmt.Errorf("SDL: RemoveAllSubscriptionsFromSdl(): %s\n", err.Error())
	} else {
		xapp.Logger.Debug("SDL: All subscriptions removed from restSubsDb")
	}
	return nil
}

func (c *Control) GetRESTKeyCount() (int, error) {

	// Get all keys
	keys, err := c.restSubsDb.GetAll(restSubSdlNs)
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return 0, fmt.Errorf("SDL: GetRESTKeyCount(), GetAll(). Error while reading E2 subscriptions  keys from DBAAS %s\n", err.Error())
	}
	return len(keys), nil
}
