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
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststube2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/assert"
)

var sdlShouldReturnError bool = false

const sdlTestErrorString string = "Test sdl returned error on purpose"

const (
	subsResponse = 1
	subsFailure  = 2
	noResponse   = 3
)

type Mock struct {
	e2SubsDb           map[string]string // Store information as a string like real db does.
	register           map[uint32]*Subscription
	subIds             []uint32
	lastAllocatedSubId uint32
	marshalLock        sync.Mutex
}

var mock *Mock

func CreateMock() *Mock {
	fmt.Println("Test CreateMock()")
	mock = new(Mock)
	mock.ResetTestSettings()
	return mock
}

func (m *Mock) ResetTestSettings() {
	m.e2SubsDb = make(map[string]string)
	m.register = make(map[uint32]*Subscription)
	var i uint32
	for i = 1; i < 65535; i++ {
		m.subIds = append(m.subIds, i)
	}
}

func (m *Mock) AllocNextSubId() uint32 {
	m.lastAllocatedSubId = m.subIds[0]
	return m.lastAllocatedSubId
}

func TestWait(t *testing.T) {
	// Wait to test settings to complete
	<-time.After(1 * time.Second)
}

func GetSubscription(t *testing.T, e2SubId uint32, responseType int, srcEndPoint, ranName string, xId string) *Subscription {
	t.Log("TEST: Getting subscription")

	subs := &Subscription{}

	// Create unpacked e2SubReqMsg
	subReqParams := &teststube2ap.E2StubSubsReqParams{}
	subReqParams.Init()

	meid := xapp.RMRMeid{}
	meid.RanName = ranName

	params := &xapp.RMRParams{}
	params.Src = srcEndPoint
	params.Xid = xId
	params.Meid = &meid

	// Create xApp transaction
	trans := mainCtrl.c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subReqParams.Req.RequestId, params.Meid)
	if trans == nil {
		t.Errorf("TEST: %s", idstring(fmt.Errorf("transaction not created"), params))
		return nil
	}

	// Allocate E2 instanceId/subId
	subReqParams.Req.RequestId.InstanceId = e2SubId

	subs.ReqId.Id = 123
	subs.ReqId.InstanceId = subReqParams.Req.RequestId.InstanceId
	subs.Meid = &meid
	subs.EpList.AddEndpoint(trans.GetEndpoint())
	subs.SubReqMsg = subReqParams.Req
	// subs.SubRFMsg contains received/cached SubscriptionResponse or SubscriptionFailure, nil in no response received
	if responseType == subsResponse {
		subs.SubRFMsg = GetSubsResponse(t, subReqParams.Req)
		subs.valid = true
	} else if responseType == subsFailure {
		subs.SubRFMsg = GetSubsFailure(t, subReqParams.Req)
		subs.valid = false
	} else if responseType == noResponse {
		subs.SubRFMsg = nil
		subs.valid = false
	}
	return subs
}

func GetSubsResponse(t *testing.T, req *e2ap.E2APSubscriptionRequest) *e2ap.E2APSubscriptionResponse {
	t.Log("TEST: Getting ricSubscriptionResponse")

	// Create e2SubRespMsg
	resp := &e2ap.E2APSubscriptionResponse{}
	resp.RequestId.Id = 123
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId

	resp.ActionAdmittedList.Items = make([]e2ap.ActionAdmittedItem, len(req.ActionSetups))
	for index := int(0); index < len(req.ActionSetups); index++ {
		resp.ActionAdmittedList.Items[index].ActionId = req.ActionSetups[index].ActionId
	}

	for index := uint64(0); index < 1; index++ {
		item := e2ap.ActionNotAdmittedItem{}
		item.ActionId = index
		item.Cause.Content = 1
		item.Cause.Value = 1
		resp.ActionNotAdmittedList.Items = append(resp.ActionNotAdmittedList.Items, item)
	}
	return resp
}

func GetSubsFailure(t *testing.T, req *e2ap.E2APSubscriptionRequest) *e2ap.E2APSubscriptionFailure {
	t.Log("TEST: Getting ricSubscriptionFailure")

	fail := &e2ap.E2APSubscriptionFailure{}
	fail.RequestId.Id = req.RequestId.Id
	fail.RequestId.InstanceId = req.RequestId.InstanceId
	fail.FunctionId = req.FunctionId
	fail.Cause.Content = e2ap.E2AP_CauseContent_RICrequest
	fail.Cause.Value = e2ap.E2AP_CauseValue_RICrequest_control_message_invalid
	return fail
}

func PrintSubscriptionData(t *testing.T, subs *Subscription) {
	t.Log("TEST: subscription data")
	t.Logf("TEST: subs.mutex = %v", subs.mutex)
	t.Logf("TEST: subs.ReqId.InstanceId = %v", subs.ReqId.InstanceId)
	t.Logf("TEST: subs.ReqId.Id = %v", subs.ReqId.Id)
	t.Logf("TEST: subs.EpList = %v", subs.EpList)
	t.Logf("TEST: subs.Meid.RanName = %v", subs.Meid.RanName)
	t.Logf("TEST: subs.SubReqMsg = %v", subs.SubReqMsg.String())
	t.Logf("TEST: subs.valid = %v", subs.valid)

	if subs.SubRFMsg != nil {
		switch typeofSubsMessage(subs.SubRFMsg) {
		case "SubResp":
			t.Logf("TEST: subs.SubRFMsg == SubResp")
			subResp := subs.SubRFMsg.(*e2ap.E2APSubscriptionResponse)
			t.Logf("TEST: subResp = %+v", subResp)
		case "SubFail":
			t.Logf("TEST: subs.SubRFMsg == SubFail")
			subFail := subs.SubRFMsg.(*e2ap.E2APSubscriptionFailure)
			t.Logf("TEST: subFail = %+v", subFail)
		}
	} else {
		t.Logf("TEST: subs.SubRFMsg == nil")
	}
}

func TestWriteSubscriptionToSdl(t *testing.T) {

	// Write one subscription
	subId := mock.AllocNextSubId()
	subs := GetSubscription(t, subId, subsResponse, "localhost:13560", "RAN_NAME_1", "123456")
	PrintSubscriptionData(t, subs)
	t.Logf("TEST: Writing subId = %v\n", subId)
	err := mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
	}
	verifyE2KeyCount(t, 1)
}

func verifyE2KeyCount(t *testing.T, expectedCount int) {

	count, err := mainCtrl.c.GetE2KeyCount()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
	} else {
		assert.Equal(t, expectedCount, count)
	}
}

func TestReadSubscriptionFromSdl(t *testing.T) {

	subId := mock.lastAllocatedSubId
	t.Logf("Reading subId = %v\n", subId)
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	PrintSubscriptionData(t, subs)
	assert.Equal(t, mock.register[subId].SubReqMsg, subs.SubReqMsg)
}

func TestRemoveSubscriptionFromSdl(t *testing.T) {

	subId := mock.lastAllocatedSubId
	err := mainCtrl.c.RemoveSubscriptionFromSdl(subId)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	delete(mock.register, subId)
	mock.subIds = append(mock.subIds, subId)
	t.Logf("TEST: subscription removed from db. subId = %v", subId)
}

func TestReadNotExistingSubscriptionFromSdl(t *testing.T) {

	var subId uint32 = 0
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		t.Logf("TEST: subscription not found from db. subId = %v", subId)
		return
	}
	t.Errorf("TEST: subscription read from db. %v", subs.String())
	PrintSubscriptionData(t, subs)
}

func TestReadNotExistingSubscriptionFromSdl2(t *testing.T) {

	var subId uint32 = 7
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		t.Logf("TEST: subscription not found from db. subId = %v", subId)
		return
	}
	t.Errorf("TEST: subscription read from db. %v", subs.String())
	PrintSubscriptionData(t, subs)
}

func TestRemoveNotExistingSubscriptionFromSdl(t *testing.T) {

	var subId uint32 = 0
	err := mainCtrl.c.RemoveSubscriptionFromSdl(subId)
	if err != nil {
		t.Logf("TEST: %s", err.Error())
		return
	}
	t.Logf("TEST: subscription removed from db. subId = %v", subId)
}

func TestWriteSubscriptionsToSdl(t *testing.T) {

	// Write 1st subscription
	subId := mock.AllocNextSubId()
	t.Logf("TEST: Writing subId = %v\n", subId)
	subs := GetSubscription(t, subId, subsResponse, "localhost:13560", "RAN_NAME_1", "123456")
	PrintSubscriptionData(t, subs)
	err := mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	t.Logf("TEST: subscription written in db = %v", subs.String())

	// Write 2nd subscription
	subId = mock.AllocNextSubId()
	t.Logf("TEST:Writing subId = %v\n", subId)
	subs = GetSubscription(t, subId, subsFailure, "localhost:13560", "RAN_NAME_2", "123457")
	PrintSubscriptionData(t, subs)
	err = mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	t.Logf("TEST: subscription written in db = %v", subs.String())

	// Write 3rd subscription
	subId = mock.AllocNextSubId()
	t.Logf("TEST:Writing subId = %v\n", subId)
	subs = GetSubscription(t, subId, noResponse, "localhost:13560", "RAN_NAME_3", "123458")
	PrintSubscriptionData(t, subs)
	err = mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	t.Logf("TEST: subscription written in db = %v", subs.String())
}

func TestReadSubscriptionsFromSdl(t *testing.T) {

	// Subscription with subId 1 was added and and removed above. Then subscriptions with subIds 2, 3 and 4 was added
	// Db subscriptions should now contain subIDs 2, 3 and 4
	var subId uint32
	for subId = 2; subId <= 4; subId++ {
		subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
		if err != nil {
			t.Errorf("TEST: %s", err.Error())
			return
		}
		PrintSubscriptionData(t, subs)
	}
}

func TestReadAllSubscriptionsFromSdl(t *testing.T) {

	// This test cases simulates submgr restart. SubIds and subscriptions are restored from db
	// after initializing mock.subIds and mock.register
	//	var err error
	subIds, register, err := mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	//	for _, subs := range mock.register {
	for _, subs := range register {
		PrintSubscriptionData(t, subs)
	}
	// SubIds slices before and after restart can't be directly compared as original slice is not stored
	// in the db. SubId values 1, 2, 3, 4 are already removed from the beginning of subIds slice above
	// so far. Next free subId is 5 in the beginning of mock.subIds slice. The db contains now however only
	// 3 subscriptions with subIds 2, 3 and 4, so only subId values 2, 3, 4 are removed from the returned
	// subIds slice and there next free value is 1
	assert.Equal(t, uint32(0x1), subIds[0])
}

func TestRemoveAllSubscriptionsFromSdl(t *testing.T) {

	err := mainCtrl.c.RemoveAllSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	t.Log("TEST: All subscription removed from db")
}

func TestReadAllSubscriptionsFromSdl2(t *testing.T) {

	// This test cases simulates submgr startup. SubIds and subscriptions are restored from empty db
	// after initializing mock.subIds and mock.register
	subIds, register, err := mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	for _, subs := range mock.register {
		PrintSubscriptionData(t, subs)
	}
	assert.Equal(t, len(subIds), 65534)
	assert.Equal(t, len(register), 0)
}

func TestWriteSubscriptionToSdlFail(t *testing.T) {

	// Try to write one subscription. Test db should return test error string
	MakeNextSdlCallFail()
	subId := mock.AllocNextSubId()
	subs := GetSubscription(t, subId, subsResponse, "localhost:13560", "RAN_NAME_1", "123456")
	PrintSubscriptionData(t, subs)
	t.Logf("TEST: Writing subId = %v\n", subId)
	err := mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
	} else {
		t.Errorf("TEST: This test case should return error")
	}
}

func TestReadSubscriptionFromSdlFail(t *testing.T) {

	// Try to read one subscription. Test db should return test error string
	MakeNextSdlCallFail()
	subId := mock.lastAllocatedSubId
	t.Logf("Reading subId = %v\n", subId)
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	PrintSubscriptionData(t, subs)
	assert.Equal(t, mock.register[subId].SubReqMsg, subs.SubReqMsg)
}

func TestRemoveSubscriptionFromSdlFail(t *testing.T) {

	// Try to remove one subscription. Test db should return test error string
	MakeNextSdlCallFail()
	subId := mock.lastAllocatedSubId
	err := mainCtrl.c.RemoveSubscriptionFromSdl(subId)
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	delete(mock.register, subId)
	mock.subIds = append(mock.subIds, subId)
	t.Logf("TEST: subscription removed from db. subId = %v", subId)
}

func TestReadAllSubscriptionsFromSdlFail(t *testing.T) {

	// Try to read all subscriptions. Test db should return test error string
	MakeNextSdlCallFail()
	// This test cases simulates submgr restart. SubIds and subscriptions are restored from db
	// after initializing mock.subIds and mock.register
	//	var err error
	subIds, register, err := mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	//	for _, subs := range mock.register {
	for _, subs := range register {
		PrintSubscriptionData(t, subs)
	}
	// SubIds slices before and after restart can't be directly compared as original slice is not stored
	// in the db. SubId values 1, 2, 3, 4 are already removed from the beginning of subIds slice above
	// so far. Next free subId is 5 in the beginning of mock.subIds slice. The db contains now however only
	// 3 subscriptions with subIds 2, 3 and 4, so only subId values 2, 3, 4 are removed from the returned
	// subIds slice and there next free value is 1
	assert.Equal(t, uint32(0x1), subIds[0])
}

func TestRemoveAllSubscriptionsFromSdlFail(t *testing.T) {

	// Try to remove all subscriptions. Test db should return test error string
	MakeNextSdlCallFail()
	err := mainCtrl.c.RemoveAllSubscriptionsFromSdl()
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	t.Log("TEST: All subscription removed from db")
}

func (m *Mock) Set(ns string, pairs ...interface{}) error {
	var key string
	var val string

	m.marshalLock.Lock()
	defer m.marshalLock.Unlock()

	if ns != e2SubSdlNs {
		return fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if sdlShouldReturnError == true {
		return GetSdlError()
	}

	for _, v := range pairs {
		reflectType := reflect.TypeOf(v)
		switch reflectType.Kind() {
		case reflect.Slice:
			val = fmt.Sprintf("%s", v.([]uint8))
		default:
			switch v.(type) {
			case string:
				key = v.(string)
			default:
				return fmt.Errorf("Set() error: Unexpected type\n")
			}
		}
	}

	if key != "" {
		m.e2SubsDb[key] = val
		subId := m.subIds[0]
		subscriptionInfo := &SubscriptionInfo{}
		err := json.Unmarshal([]byte(val), subscriptionInfo)
		if err != nil {
			return fmt.Errorf("Set() json.unmarshal error: %s\n", err.Error())
		}

		subs := mainCtrl.c.CreateSubscription(subscriptionInfo, &val)
		m.register[subId] = subs
		m.subIds = m.subIds[1:]
	} else {
		return fmt.Errorf("Set() error: key == ''\n")
	}
	return nil
}

func (m *Mock) Get(ns string, keys []string) (map[string]interface{}, error) {
	retMap := make(map[string]interface{})

	if ns != e2SubSdlNs {
		return nil, fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("Get() error: len(key) == 0\n")
	}

	if sdlShouldReturnError == true {
		return nil, GetSdlError()
	}

	for _, key := range keys {
		if key != "" {
			retMap[key] = m.e2SubsDb[key]
		} else {
			return nil, fmt.Errorf("Get() error: key == ''\n")
		}
	}
	return retMap, nil
}

func (m *Mock) GetAll(ns string) ([]string, error) {

	if ns != e2SubSdlNs {
		return nil, fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if sdlShouldReturnError == true {
		return nil, GetSdlError()
	}

	keys := []string{}
	for key, _ := range m.e2SubsDb {
		keys = append(keys, key)
	}
	return keys, nil
}

func (m *Mock) Remove(ns string, keys []string) error {

	if ns != e2SubSdlNs {
		return fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if len(keys) == 0 {
		return fmt.Errorf("Remove() error: len(key) == 0\n")
	}
	subId64, err := strconv.ParseUint(keys[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Remove() ParseUint() error: %s\n", err.Error())
	}

	if sdlShouldReturnError == true {
		return GetSdlError()
	}

	subId := uint32(subId64)
	delete(m.e2SubsDb, keys[0])
	delete(m.register, subId)
	m.subIds = append(m.subIds, subId)
	return nil
}

func (m *Mock) RemoveAll(ns string) error {

	if ns != e2SubSdlNs {
		return fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	for key := range m.e2SubsDb {
		subId64, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return fmt.Errorf("RemoveAll() ParseUint() error: %s\n", err.Error())
		}

		subId := uint32(subId64)
		delete(m.e2SubsDb, key)
		delete(m.register, subId)
		m.subIds = append(m.subIds, subId)
	}

	if sdlShouldReturnError == true {
		return GetSdlError()
	}

	return nil
}

func MakeNextSdlCallFail() {
	sdlShouldReturnError = true
}

func GetSdlError() error {
	sdlShouldReturnError = false
	return fmt.Errorf(sdlTestErrorString)
}
