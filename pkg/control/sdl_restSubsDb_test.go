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
	"strings"
	"sync"
	"testing"

	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

var sdlRestShouldReturnError bool = false

const sdlRestTestErrorString string = "Test sdl REST returned error on purpose"

type RestSubsDbMock struct {
	restSubsDb             map[string]string // Store information as a string like real db does.
	restSubscriptions      map[string]*RESTSubscription
	lastAllocatedRestSubId string
	restSubIdsInDb         []string
	marshalLock            sync.Mutex
}

var restSubsDbMock *RestSubsDbMock

func CreateRestSubsDbMock() *RestSubsDbMock {
	fmt.Println("Test CreateRestSubsDbMock()")
	restSubsDbMock = new(RestSubsDbMock)
	restSubsDbMock.ResetTestSettings()
	restSubsDbMock.lastAllocatedRestSubId = ""
	return restSubsDbMock
}

func (m *RestSubsDbMock) ResetTestSettings() {
	m.restSubsDb = make(map[string]string)
	m.restSubscriptions = make(map[string]*RESTSubscription)
}

func (m *RestSubsDbMock) AllocNextRestSubId() string {
	m.lastAllocatedRestSubId = ksuid.New().String()
	return m.lastAllocatedRestSubId
}

func (m *RestSubsDbMock) GetLastAllocatedRestSubId() string {
	return m.lastAllocatedRestSubId
}

func (m *RestSubsDbMock) AddRestSubIdsInDb(restSubId string) {
	m.restSubIdsInDb = append(m.restSubIdsInDb, restSubId)
}

func (m *RestSubsDbMock) DeleteRestSubIdsFromDb(restSubId string) {
	newrestSubIdsInDb := []string{}
	for _, i := range m.restSubIdsInDb {
		if i != restSubId {
			newrestSubIdsInDb = append(newrestSubIdsInDb, i)
		}
	}
	m.restSubIdsInDb = newrestSubIdsInDb
}

func (m *RestSubsDbMock) EmptyRestSubIdsFromDb() {
	m.restSubIdsInDb = nil
}

func CreateRESTSubscription(t *testing.T) *RESTSubscription {
	t.Log("TEST: Creating REST subscription")

	restSubscription := &RESTSubscription{}
	restSubscription.xAppServiceName = "localhost"
	restSubscription.xAppRmrEndPoint = "localhost:13560"
	restSubscription.Meid = "RAN_NAME_1"
	restSubscription.SubReqOngoing = true
	restSubscription.SubDelReqOngoing = false
	restSubscription.xAppIdToE2Id = make(map[int64]int64)
	restSubscription.lastReqMd5sum = "856e9546f6f7b65b13a86956f2e16f6a"
	return restSubscription
}

func PrintRESTSubscriptionData(t *testing.T, restSubs *RESTSubscription) {
	t.Log("TEST: RESTSubscription data")
	t.Logf("TEST: restSubs.xAppServiceName = %v", restSubs.xAppServiceName)
	t.Logf("TEST: restSubs.xAppRmrEndPoint = %v", restSubs.xAppRmrEndPoint)
	t.Logf("TEST: restSubs.Meid = %v", restSubs.Meid)
	t.Logf("TEST: restSubs.InstanceIds = %v", restSubs.InstanceIds)
	t.Logf("TEST: restSubs.xAppIdToE2Id = %v", restSubs.xAppIdToE2Id)
	t.Logf("TEST: restSubs.SubReqOngoing = %v", restSubs.SubReqOngoing)
	t.Logf("TEST: restSubs.SubDelReqOngoing = %v", restSubs.SubDelReqOngoing)
}

func TestWriteRESTSubscriptionToSdl(t *testing.T) {

	// Write one subscription
	restSubId := restSubsDbMock.AllocNextRestSubId()
	restSubs := CreateRESTSubscription(t)
	PrintRESTSubscriptionData(t, restSubs)
	t.Logf("TEST: Writing subId = %v\n", restSubId)
	err := mainCtrl.c.WriteRESTSubscriptionToSdl(restSubId, restSubs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
	}
	restSubsDbMock.AddRestSubIdsInDb(restSubId)
	verifyRESTKeyCount(t, 1)
}

func verifyRESTKeyCount(t *testing.T, expectedCount int) {

	count, err := mainCtrl.c.GetRESTKeyCount()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
	} else {
		assert.Equal(t, expectedCount, count)
	}
}

func TestReadRESTSubscriptionFromSdl(t *testing.T) {

	restSubId := restSubsDbMock.GetLastAllocatedRestSubId()
	t.Logf("Reading restSubId = %v\n", restSubId)
	restSubs, err := mainCtrl.c.ReadRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	PrintRESTSubscriptionData(t, restSubs)
	assert.Equal(t, restSubsDbMock.restSubscriptions[restSubId], restSubs)
}

func TestRemoveRESTSubscriptionFromSdl(t *testing.T) {

	restSubId := restSubsDbMock.GetLastAllocatedRestSubId()
	err := mainCtrl.c.RemoveRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	delete(restSubsDbMock.restSubscriptions, restSubId)
	t.Logf("TEST: REST subscription removed from db. subId = %v", restSubId)
	restSubsDbMock.DeleteRestSubIdsFromDb(restSubId)
}

func TestReadNotExistingRESTSubscriptionFromSdl(t *testing.T) {

	restSubId := ""
	restSubs, err := mainCtrl.c.ReadRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		t.Logf("TEST: REST subscription not found from db. restSubId = %v", restSubId)
		return
	}
	t.Errorf("TEST: REST subscription read from db. %v", restSubs)
	PrintRESTSubscriptionData(t, restSubs)
}

func TestReadNotExistingRESTSubscriptionFromSdl2(t *testing.T) {

	restSubId := "NotExistingSubsId"
	restSubs, err := mainCtrl.c.ReadRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		t.Logf("TEST: REST subscription not found from db. restSubId = %v", restSubId)
		return
	}
	t.Errorf("TEST: REST subscription read from db. %v", restSubs)
	PrintRESTSubscriptionData(t, restSubs)
}

func TestRemoveNotExistingRESTSubscriptionFromSdl(t *testing.T) {

	restSubId := ""
	err := mainCtrl.c.RemoveRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		t.Logf("TEST: %s", err.Error())
		return
	}
	t.Logf("TEST: REST subscription removed from db. subId = %v", restSubId)
}

func TestWriteRESTSubscriptionsToSdl(t *testing.T) {

	// Write 1st subscription
	restSubId := restSubsDbMock.AllocNextRestSubId()
	t.Logf("TEST: Writing restSubId = %v\n", restSubId)
	restSubs := CreateRESTSubscription(t)
	PrintRESTSubscriptionData(t, restSubs)
	err := mainCtrl.c.WriteRESTSubscriptionToSdl(restSubId, restSubs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	restSubsDbMock.AddRestSubIdsInDb(restSubId)
	t.Logf("TEST: REST subscription written in db = %v", restSubs)

	// Write 2nd subscription
	restSubId = restSubsDbMock.AllocNextRestSubId()
	t.Logf("TEST:Writing restSubId = %v\n", restSubId)
	restSubs = CreateRESTSubscription(t)
	PrintRESTSubscriptionData(t, restSubs)
	err = mainCtrl.c.WriteRESTSubscriptionToSdl(restSubId, restSubs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	restSubsDbMock.AddRestSubIdsInDb(restSubId)
	t.Logf("TEST: REST subscription written in db = %v", restSubs)

	// Write 3rd subscription
	restSubId = restSubsDbMock.AllocNextRestSubId()
	t.Logf("TEST: Writing restSubId = %v\n", restSubId)
	restSubs = CreateRESTSubscription(t)
	PrintRESTSubscriptionData(t, restSubs)
	err = mainCtrl.c.WriteRESTSubscriptionToSdl(restSubId, restSubs)
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	restSubsDbMock.AddRestSubIdsInDb(restSubId)
	t.Logf("TEST: REST subscription written in db = %v", restSubs)
}

func TestReadRESTSubscriptionsFromSdl(t *testing.T) {

	for _, restSubId := range restSubsDbMock.restSubIdsInDb {
		restSubs, err := mainCtrl.c.ReadRESTSubscriptionFromSdl(restSubId)
		if err != nil {
			t.Errorf("TEST: %s", err.Error())
			return
		}
		PrintRESTSubscriptionData(t, restSubs)
	}
}

func TestReadAllRESTSubscriptionsFromSdl(t *testing.T) {

	register, err := mainCtrl.c.ReadAllRESTSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}

	for _, restSubs := range register {
		PrintRESTSubscriptionData(t, restSubs)
	}

	assert.Equal(t, len(register), 3)
}

func TestRemoveAllRESTSubscriptionsFromSdl(t *testing.T) {

	err := mainCtrl.c.RemoveAllRESTSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	t.Log("TEST: All subscription removed from db")
	restSubsDbMock.EmptyRestSubIdsFromDb()
}

func TestReadAllRESTSubscriptionsFromSdl2(t *testing.T) {

	register, err := mainCtrl.c.ReadAllRESTSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %s", err.Error())
		return
	}
	for _, restSubs := range restSubsDbMock.restSubscriptions {
		PrintRESTSubscriptionData(t, restSubs)
	}
	assert.Equal(t, len(register), 0)
}

func TestWriteRESTSubscriptionToSdlFail(t *testing.T) {

	// Try to write one subscription.
	// Test db should return test error string
	MakeNextSdlRestCallFail()
	restsubId := restSubsDbMock.AllocNextRestSubId()
	restSubs := CreateRESTSubscription(t)
	PrintRESTSubscriptionData(t, restSubs)
	t.Logf("TEST: Writing subId = %v\n", restsubId)
	err := mainCtrl.c.WriteRESTSubscriptionToSdl(restsubId, restSubs)
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlRestTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
	} else {
		t.Errorf("TEST: This test case should return error")
	}
}

func TestReadRESTSubscriptionFromSdlFail(t *testing.T) {

	// Try to read one subscription.
	// Test db should return test error string
	MakeNextSdlRestCallFail()
	restSubId := restSubsDbMock.GetLastAllocatedRestSubId()
	t.Logf("Reading restSubId = %v\n", restSubId)
	restSubs, err := mainCtrl.c.ReadRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlRestTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	PrintRESTSubscriptionData(t, restSubs)
}

func TestRemoveRESTSubscriptionFromSdlFail(t *testing.T) {

	// Try to remove one subscription.
	// Test db should return test error string
	MakeNextSdlRestCallFail()
	restSubId := restSubsDbMock.GetLastAllocatedRestSubId()
	err := mainCtrl.c.RemoveRESTSubscriptionFromSdl(restSubId)
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlRestTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	t.Logf("TEST: subscription removed from db. subId = %v", restSubId)
}

func TestReadAllRESTSubscriptionsFromSdlFail(t *testing.T) {

	// Try to read all subscriptions.
	// Test db should return test error string
	MakeNextSdlRestCallFail()
	register, err := mainCtrl.c.ReadAllRESTSubscriptionsFromSdl()
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlRestTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}

	for _, restSubs := range register {
		PrintRESTSubscriptionData(t, restSubs)
	}
}

func TestRemoveAllRESTSubscriptionsFromSdlFail(t *testing.T) {

	// Try to remove all subscriptions.
	// Test db should return test error string
	MakeNextSdlRestCallFail()
	err := mainCtrl.c.RemoveAllRESTSubscriptionsFromSdl()
	if err != nil {
		if !strings.Contains(fmt.Sprintf("%s", err), sdlRestTestErrorString) {
			t.Errorf("TEST: %s", err.Error())
		}
		return
	} else {
		t.Errorf("TEST: This test case should return error")
	}
	t.Log("TEST: All subscription removed from db")
}

func (m *RestSubsDbMock) Set(ns string, pairs ...interface{}) error {
	var key string
	var val string

	m.marshalLock.Lock()
	defer m.marshalLock.Unlock()

	if ns != restSubSdlNs {
		return fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if sdlRestShouldReturnError == true {
		return GetSdlRestError()
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
		m.restSubsDb[key] = val
		restSubId := key
		restSubscriptionInfo := &RESTSubscriptionInfo{}
		err := json.Unmarshal([]byte(val), restSubscriptionInfo)
		if err != nil {
			return fmt.Errorf("Set() json.unmarshal error: %s\n", err.Error())
		}

		restSubs := mainCtrl.c.CreateRESTSubscription(restSubscriptionInfo, &val)
		m.restSubscriptions[restSubId] = restSubs
	} else {
		return fmt.Errorf("Set() error: key == ''\n")
	}
	return nil
}

func (m *RestSubsDbMock) Get(ns string, keys []string) (map[string]interface{}, error) {
	retMap := make(map[string]interface{})

	if ns != restSubSdlNs {
		return nil, fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("Get() error: len(key) == 0\n")
	}

	if sdlRestShouldReturnError == true {
		return nil, GetSdlRestError()
	}

	for _, key := range keys {
		if key != "" {
			retMap[key] = m.restSubsDb[key]
		} else {
			return nil, fmt.Errorf("Get() error: key == ''\n")
		}
	}
	return retMap, nil
}

func (m *RestSubsDbMock) GetAll(ns string) ([]string, error) {

	if ns != restSubSdlNs {
		return nil, fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if sdlRestShouldReturnError == true {
		return nil, GetSdlRestError()
	}

	keys := []string{}
	for key, _ := range m.restSubsDb {
		keys = append(keys, key)
	}
	return keys, nil
}

func (m *RestSubsDbMock) Remove(ns string, keys []string) error {

	if ns != restSubSdlNs {
		return fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	if len(keys) == 0 {
		return fmt.Errorf("Remove() error: len(key) == 0\n")
	}

	if sdlRestShouldReturnError == true {
		return GetSdlRestError()
	}

	restSubId := keys[0]
	delete(m.restSubsDb, restSubId)
	delete(m.restSubscriptions, restSubId)
	return nil
}

func (m *RestSubsDbMock) RemoveAll(ns string) error {

	if ns != restSubSdlNs {
		return fmt.Errorf("Unexpected namespace '%s' error\n", ns)
	}

	for key := range m.restSubsDb {

		restSubId := key
		delete(m.restSubsDb, restSubId)
		delete(m.restSubscriptions, restSubId)
	}

	if sdlRestShouldReturnError == true {
		return GetSdlRestError()
	}

	return nil
}

func MakeNextSdlRestCallFail() {
	sdlRestShouldReturnError = true
}

func GetSdlRestError() error {
	sdlRestShouldReturnError = false
	return fmt.Errorf(sdlRestTestErrorString)
}
