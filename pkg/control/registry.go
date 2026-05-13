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
	"strings"
	"sync"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"

	// grpc update
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/ricapi"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type RESTSubscription struct {
	Created          string
	xAppServiceName  string
	xAppRmrEndPoint  string
	Meid             string
	InstanceIds      []uint32
	xAppIdToE2Id     map[int64]int64
	SubReqOngoing    bool
	SubDelReqOngoing bool
	lastReqMd5sum    string
	// START: grpc update
	// [추가] gRPC 응답을 위해 xApp이 보낸 Transaction ID를 보관합니다.
	PtransID *ricapi.ProcedureTransactionId
	// END: grpc update
}

func (r *RESTSubscription) AddE2InstanceId(instanceId uint32) {

	for _, v := range r.InstanceIds {
		if v == instanceId {
			return
		}

	}

	r.InstanceIds = append(r.InstanceIds, instanceId)
}

func (r *RESTSubscription) AddMd5Sum(md5sum string) {
	if md5sum != "" {
		r.lastReqMd5sum = md5sum
	} else {
		xapp.Logger.Error("EMPTY md5sum attempted to be add to subscrition")
	}
}

func (r *RESTSubscription) DeleteE2InstanceId(instanceId uint32) {
	r.InstanceIds = r.InstanceIds[1:]
}

func (r *RESTSubscription) AddXappIdToE2Id(xAppEventInstanceID int64, e2EventInstanceID int64) {
	r.xAppIdToE2Id[xAppEventInstanceID] = e2EventInstanceID
}

func (r *RESTSubscription) GetE2IdFromXappIdToE2Id(xAppEventInstanceID int64) int64 {
	return r.xAppIdToE2Id[xAppEventInstanceID]
}

func (r *RESTSubscription) DeleteXappIdToE2Id(xAppEventInstanceID int64) {
	delete(r.xAppIdToE2Id, xAppEventInstanceID)
}

func (r *RESTSubscription) SetProcessed(err error) {
	r.SubReqOngoing = false
	if err != nil {
		r.lastReqMd5sum = ""
	}
}

type Registry struct {
	mutex             *sync.Mutex
	register          map[uint32]*Subscription
	subIds            []uint32
	rtmgrClient       *RtmgrClient
	restSubscriptions map[string]*RESTSubscription
}

func (r *Registry) Initialize() {
	r.mutex = new(sync.Mutex)
	r.register = make(map[uint32]*Subscription)
	r.restSubscriptions = make(map[string]*RESTSubscription)

	var i uint32
	for i = 1; i < 65535; i++ {
		r.subIds = append(r.subIds, i)
	}
}

func (r *Registry) GetAllRestSubscriptionsJson() []byte {

	r.mutex.Lock()
	defer r.mutex.Unlock()
	restSubscriptionsJson, err := json.Marshal(r.restSubscriptions)
	if err != nil {
		xapp.Logger.Error("GetAllRestSubscriptions() json.Marshal error: %v", err)
	}
	return restSubscriptionsJson
}

func (r *Registry) GetAllE2NodeRestSubscriptionsJson(ranName string) []byte {

	restSubscriptions := r.GetAllE2NodeRestSubscriptions(ranName)
	e2NodeRestSubscriptionsJson, err := json.Marshal(restSubscriptions)
	if err != nil {
		xapp.Logger.Error("GetE2NodeRestSubscriptions() json.Marshal error: %v", err)
	}
	return e2NodeRestSubscriptionsJson
}

func (r *Registry) GetAllE2NodeRestSubscriptions(ranName string) map[string]RESTSubscription {

	r.mutex.Lock()
	defer r.mutex.Unlock()
	var restSubscriptions map[string]RESTSubscription
	restSubscriptions = make(map[string]RESTSubscription)
	for restSubsId, restSubscription := range r.restSubscriptions {
		if restSubscription.Meid == ranName {
			restSubscriptions[restSubsId] = *restSubscription
		}
	}
	return restSubscriptions
}

func (r *Registry) GetAllXappsJson() []byte {

	r.mutex.Lock()
	var xappList []string
	var xappsMap map[string]string
	xappsMap = make(map[string]string)
	for _, restSubscription := range r.restSubscriptions {
		_, ok := xappsMap[restSubscription.xAppServiceName]
		if !ok {
			xappsMap[restSubscription.xAppServiceName] = restSubscription.xAppServiceName
			xappList = append(xappList, restSubscription.xAppServiceName)
		}
	}
	r.mutex.Unlock()

	xappsJson, err := json.Marshal(xappList)
	if err != nil {
		xapp.Logger.Error("GetXapps() json.Marshal error: %v", err)
	}
	return xappsJson
}

func (r *Registry) GetAllXapps() map[string]string {

	r.mutex.Lock()
	defer r.mutex.Unlock()
	var xappsMap map[string]string
	xappsMap = make(map[string]string)
	for _, restSubscription := range r.restSubscriptions {
		_, ok := xappsMap[restSubscription.xAppServiceName]
		if !ok {
			xappsMap[restSubscription.xAppServiceName] = restSubscription.xAppServiceName
		}
	}
	return xappsMap
}

func (r *Registry) GetAllXappRestSubscriptionsJson(xAppServiceName string) []byte {

	xappRestSubscriptions := r.GetAllXappRestSubscriptions(xAppServiceName)
	xappRestSubscriptionsJson, err := json.Marshal(xappRestSubscriptions)
	if err != nil {
		xapp.Logger.Error("GetXappRestSubscriptions() json.Marshal error: %v", err)
	}
	return xappRestSubscriptionsJson
}

func (r *Registry) GetAllXappRestSubscriptions(xAppServiceName string) map[string]RESTSubscription {

	r.mutex.Lock()
	defer r.mutex.Unlock()
	var xappRestSubscriptions map[string]RESTSubscription
	xappRestSubscriptions = make(map[string]RESTSubscription)
	for restSubsId, xappRestSubscription := range r.restSubscriptions {
		if xappRestSubscription.xAppServiceName == xAppServiceName {
			xappRestSubscriptions[restSubsId] = *xappRestSubscription
		}
	}
	return xappRestSubscriptions
}

func (r *Registry) GetE2SubscriptionsJson(restSubsId string) ([]byte, error) {

	// Get all E2 subscriptions of a REST subscription
	restSubs, err := r.GetRESTSubscription(restSubsId, false)
	if err != nil {
		return nil, err
	}

	r.mutex.Lock()
	var e2Subscriptions []Subscription
	for _, e2SubId := range restSubs.InstanceIds {
		e2Subscription, ok := r.register[e2SubId]
		if ok {
			e2Subscriptions = append(e2Subscriptions, *e2Subscription)
		}
	}
	r.mutex.Unlock()
	e2SubscriptionsJson, err := json.Marshal(e2Subscriptions)
	if err != nil {
		xapp.Logger.Error("GetE2Subscriptions() json.Marshal error: %v", err)
	}
	return e2SubscriptionsJson, nil
}

func (r *Registry) CreateRESTSubscription(restSubId *string, xappServiceName *string, xAppRmrEndPoint *string, maid *string) *RESTSubscription {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	newRestSubscription := RESTSubscription{}
	newRestSubscription.Created = time.Now().Format("2006-01-02 15:04:05.000")
	newRestSubscription.xAppServiceName = *xappServiceName
	newRestSubscription.xAppRmrEndPoint = *xAppRmrEndPoint
	newRestSubscription.Meid = *maid
	newRestSubscription.SubReqOngoing = true
	newRestSubscription.SubDelReqOngoing = false
	r.restSubscriptions[*restSubId] = &newRestSubscription
	newRestSubscription.xAppIdToE2Id = make(map[int64]int64)
	xapp.Logger.Debug("Registry: Created REST subscription successfully. restSubId=%v, subscriptionCount=%v, e2apSubscriptionCount=%v", *restSubId, len(r.restSubscriptions), len(r.register))
	return &newRestSubscription
}

func (r *Registry) DeleteRESTSubscription(restSubId *string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.restSubscriptions, *restSubId)
	xapp.Logger.Debug("Registry: Deleted REST subscription successfully. restSubId=%v, subscriptionCount=%v", *restSubId, len(r.restSubscriptions))
}

func (r *Registry) GetRESTSubscription(restSubId string, IsDelReqOngoing bool) (*RESTSubscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if restSubscription, ok := r.restSubscriptions[restSubId]; ok {
		// Subscription deletion is not allowed if prosessing subscription request in not ready
		if restSubscription.SubDelReqOngoing == false && restSubscription.SubReqOngoing == false {
			if IsDelReqOngoing == true {
				restSubscription.SubDelReqOngoing = true
			}
			r.restSubscriptions[restSubId] = restSubscription
			return restSubscription, nil
		} else {
			return restSubscription, fmt.Errorf("Registry: REST request is still ongoing for the endpoint=%v, restSubId=%v, SubDelReqOngoing=%v, SubReqOngoing=%v", restSubscription, restSubId, restSubscription.SubDelReqOngoing, restSubscription.SubReqOngoing)
		}
	}
	return nil, fmt.Errorf("Registry: No valid subscription found with restSubId=%v", restSubId)
}

func (r *Registry) QueryHandler() (models.SubscriptionList, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	resp := models.SubscriptionList{}
	for _, subs := range r.register {
		subs.mutex.Lock()
		resp = append(resp, &models.SubscriptionData{SubscriptionID: int64(subs.ReqId.InstanceId), Meid: subs.Meid.RanName, ClientEndpoint: subs.EpList.StringList()})
		subs.mutex.Unlock()
	}
	return resp, nil
}

func (r *Registry) allocateSubs(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest, resetTestFlag bool, rmrRoutecreated bool) (*Subscription, error) {
	if len(r.subIds) > 0 {
		subId := r.subIds[0]
		r.subIds = r.subIds[1:]
		if _, ok := r.register[subId]; ok == true {
			r.subIds = append(r.subIds, subId)
			return nil, fmt.Errorf("Registry: Failed to reserve subscription exists")
		}
		subs := &Subscription{
			registry:         r,
			Meid:             trans.Meid,
			RMRRouteCreated:  rmrRoutecreated,
			SubReqMsg:        subReqMsg,
			OngoingReqCount:  0,
			OngoingDelCount:  0,
			valid:            true,
			PolicyUpdate:     false,
			RetryFromXapp:    false,
			SubRespRcvd:      false,
			DeleteFromDb:     false,
			NoRespToXapp:     false,
			DoNotWaitSubResp: false,
		}
		subs.ReqId.Id = subReqMsg.RequestId.Id
		subs.ReqId.InstanceId = subId
		r.SetResetTestFlag(resetTestFlag, subs)

		if subs.EpList.AddEndpoint(trans.GetEndpoint()) == false {
			r.subIds = append(r.subIds, subs.ReqId.InstanceId)
			return nil, fmt.Errorf("Registry: Endpoint existing already in subscription")
		}
		return subs, nil
	}
	return nil, fmt.Errorf("Registry: Failed to reserve subscription no free ids")
}

func (r *Registry) findExistingSubs(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest) (*Subscription, bool) {

	for _, subs := range r.register {
		if subs.IsMergeable(trans, subReqMsg) {

			//
			// check if there has been race conditions
			//
			subs.mutex.Lock()
			//subs has been set to invalid
			if subs.valid == false {
				subs.mutex.Unlock()
				continue
			}
			// If size is zero, entry is to be deleted
			if subs.EpList.Size() == 0 {
				subs.mutex.Unlock()
				continue
			}
			// Try to add to endpointlist. Adding fails if endpoint is already in the list
			if subs.EpList.AddEndpoint(trans.GetEndpoint()) == false {
				subs.mutex.Unlock()
				xapp.Logger.Debug("Registry: Subs with requesting endpoint found. %s for %s", subs.String(), trans.String())
				return subs, true
			}
			subs.mutex.Unlock()

			xapp.Logger.Debug("Registry: Mergeable subs found. %s for %s", subs.String(), trans.String())
			return subs, false
		}
	}
	return nil, false
}

func (r *Registry) AssignToSubscription(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest, resetTestFlag bool, c *Control, createRMRRoute bool) (*Subscription, ErrorInfo, error) {
	var err error
	var newAlloc bool
	errorInfo := ErrorInfo{}
	r.mutex.Lock()
	defer r.mutex.Unlock()

	//
	// Check validity of subscription action types
	//
	actionType, err := r.CheckActionTypes(subReqMsg)
	if err != nil {
		xapp.Logger.Debug("CREATE %s", err)
		err = fmt.Errorf("E2 content validation failed")
		return nil, errorInfo, err
	}

	//
	// Find possible existing Policy subscription
	//
	if actionType == e2ap.E2AP_ActionTypePolicy {
		if subs, ok := r.register[trans.GetSubId()]; ok {
			xapp.Logger.Debug("CREATE %s. Existing subscription for Policy found.", subs.String())
			// Update message data to subscription
			subs.SubReqMsg = subReqMsg
			subs.PolicyUpdate = true
			subs.SetCachedResponse(nil, true)
			r.SetResetTestFlag(resetTestFlag, subs)
			return subs, errorInfo, nil
		}
	}

	subs, endPointFound := r.findExistingSubs(trans, subReqMsg)
	if subs == nil {
		if subs, err = r.allocateSubs(trans, subReqMsg, resetTestFlag, createRMRRoute); err != nil {
			xapp.Logger.Error("%s", err.Error())
			err = fmt.Errorf("subscription not allocated")
			return nil, errorInfo, err
		}
		newAlloc = true
	} else if endPointFound == true {
		// Requesting endpoint is already present in existing subscription. This can happen if xApp is restarted.
		subs.RetryFromXapp = true
		xapp.Logger.Debug("CREATE subReqMsg.InstanceId=%v. Same subscription %s already exists.", subReqMsg.InstanceId, subs.String())
		c.UpdateCounter(cDuplicateE2SubReq)
		return subs, errorInfo, nil
	}

	//
	// Add to subscription
	//
	subs.mutex.Lock()
	defer subs.mutex.Unlock()

	epamount := subs.EpList.Size()
	xapp.Logger.Debug("AssignToSubscription subs.EpList.Size()=%v", subs.EpList.Size())

	r.mutex.Unlock()
	//
	// Subscription route updates
	//
	if createRMRRoute == true {
		if epamount == 1 {
			errorInfo, err = r.RouteCreate(subs, c)
		} else {
			errorInfo, err = r.RouteCreateUpdate(subs, c)
		}
	} else {
		xapp.Logger.Debug("RMR route not created: createRMRRoute=%v", createRMRRoute)
	}
	r.mutex.Lock()

	if err != nil {
		if newAlloc {
			r.subIds = append(r.subIds, subs.ReqId.InstanceId)
		}
		// Delete already added endpoint for the request
		subs.EpList.DelEndpoint(trans.GetEndpoint())
		return nil, errorInfo, err
	}

	if newAlloc {
		r.register[subs.ReqId.InstanceId] = subs
	}
	xapp.Logger.Debug("CREATE %s", subs.String())
	xapp.Logger.Debug("Registry: substable=%v", r.register)
	return subs, errorInfo, nil
}

func (r *Registry) RouteCreate(subs *Subscription, c *Control) (ErrorInfo, error) {
	errorInfo := ErrorInfo{}
	subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.InstanceId)}
	err := r.rtmgrClient.SubscriptionRequestCreate(subRouteAction)
	if err != nil {
		if strings.Contains(err.Error(), "status 400") {
			errorInfo.TimeoutType = models.SubscriptionInstanceTimeoutTypeRTMGRTimeout
		} else {
			errorInfo.ErrorSource = models.SubscriptionInstanceErrorSourceRTMGR
		}
		errorInfo.ErrorCause = err.Error()
		c.UpdateCounter(cRouteCreateFail)
		xapp.Logger.Error("%s", err.Error())
		err = fmt.Errorf("RTMGR route create failure")
	}
	return errorInfo, err
}

func (r *Registry) RouteCreateUpdate(subs *Subscription, c *Control) (ErrorInfo, error) {
	errorInfo := ErrorInfo{}
	subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.InstanceId)}
	err := r.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		if strings.Contains(err.Error(), "status 400") {
			errorInfo.TimeoutType = models.SubscriptionInstanceTimeoutTypeRTMGRTimeout
		} else {
			errorInfo.ErrorSource = models.SubscriptionInstanceErrorSourceRTMGR
		}
		errorInfo.ErrorCause = err.Error()
		c.UpdateCounter(cRouteCreateUpdateFail)
		xapp.Logger.Error("%s", err.Error())
		err = fmt.Errorf("RTMGR route update failure")
		return errorInfo, err
	}
	c.UpdateCounter(cMergedSubscriptions)
	return errorInfo, err
}

func (r *Registry) CheckActionTypes(subReqMsg *e2ap.E2APSubscriptionRequest) (uint64, error) {
	var reportFound bool = false
	var policyFound bool = false
	var insertFound bool = false

	for _, acts := range subReqMsg.ActionSetups {
		if acts.ActionType == e2ap.E2AP_ActionTypeReport {
			reportFound = true
		}
		if acts.ActionType == e2ap.E2AP_ActionTypePolicy {
			policyFound = true
		}
		if acts.ActionType == e2ap.E2AP_ActionTypeInsert {
			insertFound = true
		}
	}
	if reportFound == true && policyFound == true || reportFound == true && insertFound == true || policyFound == true && insertFound == true {
		return e2ap.E2AP_ActionTypeInvalid, fmt.Errorf("Different action types (Report, Policy or Insert) in same RICactions-ToBeSetup-List")
	}
	if reportFound == true {
		return e2ap.E2AP_ActionTypeReport, nil
	}
	if policyFound == true {
		return e2ap.E2AP_ActionTypePolicy, nil
	}
	if insertFound == true {
		return e2ap.E2AP_ActionTypeInsert, nil
	}
	return e2ap.E2AP_ActionTypeInvalid, fmt.Errorf("Invalid action type in RICactions-ToBeSetup-List")
}

func (r *Registry) RemoveFromSubscription(subs *Subscription, trans *TransactionXapp, waitRouteClean time.Duration, c *Control) {

	xapp.Logger.Debug("RemoveFromSubscription %s", idstring(nil, trans, subs, trans))
	r.mutex.Lock()
	defer r.mutex.Unlock()
	subs.mutex.Lock()
	defer subs.mutex.Unlock()

	delStatus := subs.EpList.DelEndpoint(trans.GetEndpoint())
	epamount := subs.EpList.Size()

	subId := subs.ReqId.InstanceId
	if delStatus == false {
		return
	}

	if waitRouteClean > 0 {
		// Wait here that response is delivered to xApp via RMR before route is cleaned
		xapp.Logger.Debug("Pending %v in order to wait route cleanup", waitRouteClean)
		r.mutex.Unlock()
		time.Sleep(waitRouteClean)
		r.mutex.Lock()
	}

	xapp.Logger.Debug("CLEAN %s", subs.String())

	if epamount == 0 {
		//
		// Subscription route delete
		//
		if subs.RMRRouteCreated == true {
			r.RouteDelete(subs, trans, c)
		}

		// Not merged subscription is being deleted
		xapp.Logger.Debug("Subscription route delete RemoveSubscriptionFromDb")
		c.RemoveSubscriptionFromDb(subs)

		//
		// Subscription release
		//

		if _, ok := r.register[subId]; ok {
			xapp.Logger.Debug("RELEASE %s", subs.String())
			delete(r.register, subId)
			xapp.Logger.Debug("Registry: substable=%v", r.register)
		}
		r.subIds = append(r.subIds, subId)
	} else if subs.EpList.Size() > 0 {
		//
		// Subscription route update
		//
		if subs.RMRRouteCreated == true {
			r.RouteDeleteUpdate(subs, c)
		}

		// Endpoint of merged subscription is being deleted
		xapp.Logger.Debug("Subscription route update WriteSubscriptionToDb")
		err := c.WriteSubscriptionToDb(subs)
		if err != nil {
			xapp.Logger.Error("tracker.UnTrackTransaction() failed:%s", err.Error())
		}
		c.UpdateCounter(cUnmergedSubscriptions)
	}
	return
}

func (r *Registry) RouteDelete(subs *Subscription, trans *TransactionXapp, c *Control) {
	tmpList := xapp.RmrEndpointList{}
	tmpList.AddEndpoint(trans.GetEndpoint())
	subRouteAction := SubRouteInfo{tmpList, uint16(subs.ReqId.InstanceId)}
	if err := r.rtmgrClient.SubscriptionRequestDelete(subRouteAction); err != nil {
		c.UpdateCounter(cRouteDeleteFail)
	}
}

func (r *Registry) RouteDeleteUpdate(subs *Subscription, c *Control) {
	subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.InstanceId)}
	if err := r.rtmgrClient.SubscriptionRequestUpdate(subRouteAction); err != nil {
		c.UpdateCounter(cRouteDeleteUpdateFail)
	}
}

func (r *Registry) GetSubscription(subId uint32) *Subscription {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[subId]; ok {
		return r.register[subId]
	}
	return nil
}

func (r *Registry) GetSubscriptionFirstMatch(subIds []uint32) (*Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, subId := range subIds {
		if _, ok := r.register[subId]; ok {
			return r.register[subId], nil
		}
	}
	return nil, fmt.Errorf("No valid subscription found with subIds %v", subIds)
}

func (r *Registry) SetResetTestFlag(resetTestFlag bool, subs *Subscription) {
	if resetTestFlag == true {
		// This is used in submgr restart unit tests
		xapp.Logger.Debug("resetTestFlag == true")
		subs.DoNotWaitSubResp = true
	} else {
		xapp.Logger.Debug("resetTestFlag == false")
	}
}

func (r *Registry) DeleteAllE2Subscriptions(ranName string, c *Control) {

	xapp.Logger.Debug("Registry: DeleteAllE2Subscriptions()")
	for subId, subs := range r.register {
		if subs.Meid.RanName == ranName {
			if subs.OngoingReqCount != 0 || subs.OngoingDelCount != 0 {
				// Subscription creation or deletion processes need to be processed gracefully till the end.
				// Subscription is deleted at end of the process in both cases.
				xapp.Logger.Debug("Registry: E2 subscription under prosessing ongoing cannot delete it yet. subId=%v, OngoingReqCount=%v, OngoingDelCount=%v", subId, subs.OngoingReqCount, subs.OngoingDelCount)
				continue
			} else {
				// Delete route
				if subs.RMRRouteCreated == true {
					for _, ep := range subs.EpList.Endpoints {
						tmpList := xapp.RmrEndpointList{}
						tmpList.AddEndpoint(&ep)
						subRouteAction := SubRouteInfo{tmpList, uint16(subs.ReqId.InstanceId)}
						if err := r.rtmgrClient.SubscriptionRequestDelete(subRouteAction); err != nil {
							c.UpdateCounter(cRouteDeleteFail)
						}
					}
				}
				// Delete E2 subscription from registry and db
				xapp.Logger.Debug("Registry: Subscription delete. subId=%v", subId)
				delete(r.register, subId)
				r.subIds = append(r.subIds, subId)
				c.RemoveSubscriptionFromDb(subs)
			}
		}
	}

	// Delete REST subscription from registry and db
	for restSubId, restSubs := range r.restSubscriptions {
		if restSubs.Meid == ranName {
			if restSubs.SubReqOngoing == true || restSubs.SubDelReqOngoing == true {
				// Subscription creation or deletion processes need to be processed gracefully till the end.
				// Subscription is deleted at end of the process in both cases.
				xapp.Logger.Debug("Registry: REST subscription under prosessing ongoing cannot delete it yet. RestSubId=%v, SubReqOngoing=%v, SubDelReqOngoing=%v", restSubId, restSubs.SubReqOngoing, restSubs.SubDelReqOngoing)
				continue
			} else {
				xapp.Logger.Debug("Registry: REST subscription delete. subId=%v", restSubId)
				delete(r.restSubscriptions, restSubId)
				c.RemoveRESTSubscriptionFromDb(restSubId)
			}
		}
	}
}
