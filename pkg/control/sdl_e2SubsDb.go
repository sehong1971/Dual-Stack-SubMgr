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
	"strconv"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	sdl "gerrit.o-ran-sc.org/r/ric-plt/sdlgo"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

const e2SubSdlNs = "submgr_e2SubsDb"

type SubscriptionInfo struct {
	Valid        bool
	ReqId        RequestId
	Meid         xapp.RMRMeid
	EpList       xapp.RmrEndpointList
	SubReqMsg    e2ap.E2APSubscriptionRequest
	SubRespMsg   e2ap.E2APSubscriptionResponse
	SubRespRcvd  string
	PolicyUpdate bool
}

func CreateSdl() Sdlnterface {
	return sdl.NewSyncStorage()
}

func (c *Control) WriteSubscriptionToSdl(subId uint32, subs *Subscription) error {

	var subscriptionInfo SubscriptionInfo
	subscriptionInfo.Valid = subs.valid
	subscriptionInfo.ReqId = subs.ReqId
	subscriptionInfo.Meid = *subs.Meid
	subscriptionInfo.EpList = subs.EpList
	subscriptionInfo.SubReqMsg = *subs.SubReqMsg
	subscriptionInfo.PolicyUpdate = subs.PolicyUpdate

	if typeofSubsMessage(subs.SubRFMsg) == "SubResp" {
		subscriptionInfo.SubRespRcvd = "SubResp"
		subscriptionInfo.SubRespMsg = *subs.SubRFMsg.(*e2ap.E2APSubscriptionResponse)
	} else {
		subscriptionInfo.SubRespRcvd = ""
	}

	jsonData, err := json.Marshal(subscriptionInfo)
	if err != nil {
		return fmt.Errorf("SDL: WriteSubscriptionToSdl() json.Marshal error: %s", err.Error())
	}

	if err = c.e2SubsDb.Set(e2SubSdlNs, strconv.FormatUint(uint64(subId), 10), jsonData); err != nil {
		c.UpdateCounter(cSDLWriteFailure)
		return fmt.Errorf("SDL: WriteSubscriptionToSdl(): %s", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription written in e2SubsDb. subId = %v", subId)
	}
	return nil
}

func (c *Control) ReadSubscriptionFromSdl(subId uint32) (*Subscription, error) {

	// This function is now just for testing purpose
	key := strconv.FormatUint(uint64(subId), 10)
	retMap, err := c.e2SubsDb.Get(e2SubSdlNs, []string{key})
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl(): %s", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription read from e2SubsDb.  subId = %v", subId)
	}

	subs := &Subscription{}
	for _, iSubscriptionInfo := range retMap {

		if iSubscriptionInfo == nil {
			return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl() subscription not found. subId = %v\n", subId)
		}

		subscriptionInfo := &SubscriptionInfo{}
		jsonSubscriptionInfo := iSubscriptionInfo.(string)

		if err := json.Unmarshal([]byte(jsonSubscriptionInfo), subscriptionInfo); err != nil {
			return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl() json.unmarshal error: %s\n", err.Error())
		}

		subs = c.CreateSubscription(subscriptionInfo, &jsonSubscriptionInfo)
	}
	return subs, nil
}

func (c *Control) CreateSubscription(subscriptionInfo *SubscriptionInfo, jsonSubscriptionInfo *string) *Subscription {

	subs := &Subscription{}
	subs.registry = c.registry
	subs.valid = subscriptionInfo.Valid
	subs.ReqId = subscriptionInfo.ReqId
	meid := xapp.RMRMeid{}
	meid = subscriptionInfo.Meid
	subs.Meid = &meid
	subs.EpList = subscriptionInfo.EpList
	subs.TheTrans = nil
	subReq := e2ap.E2APSubscriptionRequest{}
	subReq = subscriptionInfo.SubReqMsg
	subs.SubReqMsg = &subReq
	subs.PolicyUpdate = subscriptionInfo.PolicyUpdate

	if subscriptionInfo.SubRespRcvd == "SubResp" {
		subs.SubRespRcvd = true
		subResp := e2ap.E2APSubscriptionResponse{}
		subResp = subscriptionInfo.SubRespMsg
		subs.SubRFMsg = &subResp
	} else {
		subs.SubRespRcvd = false
		subs.SubRFMsg = nil
		xapp.Logger.Debug("SDL: CreateSubscription() subscriptionInfo.SubRespRcvd == '', InstanceId=%v ", subscriptionInfo.ReqId.InstanceId)
	}
	return subs
}

func (c *Control) RemoveSubscriptionFromSdl(subId uint32) error {

	key := strconv.FormatUint(uint64(subId), 10)
	if err := c.e2SubsDb.Remove(e2SubSdlNs, []string{key}); err != nil {
		return fmt.Errorf("SDL: RemoveSubscriptionfromSdl(): %s\n", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription removed from e2SubsDb. subId = %v", subId)
	}
	return nil
}

func (c *Control) ReadAllSubscriptionsFromSdl() ([]uint32, map[uint32]*Subscription, error) {

	// Read all subscriptionInfos
	var subIds []uint32
	var i uint32
	for i = 1; i < 65535; i++ {
		subIds = append(subIds, i)
	}

	retMap := make(map[uint32]*Subscription)
	// Get all keys
	keys, err := c.e2SubsDb.GetAll(e2SubSdlNs)
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl(), GetAll(). Error while reading E2 subscriptions  keys from DBAAS %s\n", err.Error())
	}

	if len(keys) == 0 {
		return subIds, retMap, nil
	}

	// Get all subscriptionInfos
	iSubscriptionMap, err := c.e2SubsDb.Get(e2SubSdlNs, keys)
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl(), Get():  Error while reading E2 subscriptions from DBAAS %s\n", err.Error())
	}

	for _, iSubscriptionInfo := range iSubscriptionMap {

		if iSubscriptionInfo == nil {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() iSubscriptionInfo = nil\n")
		}

		subscriptionInfo := &SubscriptionInfo{}
		jsonSubscriptionInfo := iSubscriptionInfo.(string)

		if err := json.Unmarshal([]byte(jsonSubscriptionInfo), subscriptionInfo); err != nil {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() json.unmarshal error: %s\n", err.Error())
		}

		subs := c.CreateSubscription(subscriptionInfo, &jsonSubscriptionInfo)

		if int(subscriptionInfo.ReqId.InstanceId) >= len(subIds) {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() index is out of range. Index is %d with slice length %d", subscriptionInfo.ReqId.InstanceId, len(subIds))
		}
		retMap[subscriptionInfo.ReqId.InstanceId] = subs

		// Remove subId from free subIds. Original slice is modified here!
		if subIds, err = removeNumber(subIds, subscriptionInfo.ReqId.InstanceId); err != nil {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() error: %s\n", err.Error())
		}
	}
	return subIds, retMap, nil
}

func removeNumber(s []uint32, removedNum uint32) ([]uint32, error) {
	for i, num := range s {
		if removedNum == uint32(num) {
			s = append(s[:i], s[i+1:]...)
			return s[:len(s)], nil
		}
	}
	return nil, fmt.Errorf("SDL: To be removed number not in the slice. removedNum: %v", removedNum)
}
func (c *Control) RemoveAllSubscriptionsFromSdl() error {

	if err := c.e2SubsDb.RemoveAll(e2SubSdlNs); err != nil {
		c.UpdateCounter(cSDLRemoveFailure)
		return fmt.Errorf("SDL: RemoveAllSubscriptionsFromSdl(): %s\n", err.Error())
	} else {
		xapp.Logger.Debug("SDL: All subscriptions removed from e2SubsDb")
	}
	return nil
}

func (c *Control) GetE2KeyCount() (int, error) {

	// Get all keys
	keys, err := c.e2SubsDb.GetAll(e2SubSdlNs)
	if err != nil {
		c.UpdateCounter(cSDLReadFailure)
		return 0, fmt.Errorf("SDL: GetE2KeyCount(), GetAll(). Error while reading E2 subscriptions  keys from DBAAS %s\n", err.Error())
	}
	return len(keys), nil
}
