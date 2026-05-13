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
	"gerrit.o-ran-sc.org/r/ric-plt/nodeb-rnib.git/entities"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strings"
	"sync"
	"testing"
)

var xappRnibMock *XappRnibMock

type XappRnibMock struct {
	Mutex            sync.Mutex
	nbIdentityMap    map[string]xapp.RNIBNbIdentity
	RNIBNodebInfoMap map[string]xapp.RNIBNodebInfo
	RnibSubscription RnibSubscription // Submgr can have only one subscription
}

type RnibSubscription struct {
	Channel string                            // Subscribed channel/topic "RAN_CONNECTION_STATUS_CHANGE"
	cb      func(ch string, events ...string) // Submgr's call back function
}

func CreateXappRnibIfMock() *XappRnibMock {
	fmt.Println("XappRnibMock: CreateXappRnibIfMock()")
	xappRnibMock = new(XappRnibMock)
	xappRnibMock.Init()
	return xappRnibMock
}

func (x *XappRnibMock) Init() {
	x.nbIdentityMap = make(map[string]xapp.RNIBNbIdentity, 0)
	x.RNIBNodebInfoMap = make(map[string]xapp.RNIBNodebInfo, 0)
}

func TestMock(t *testing.T) {

	// Current UT test cases use these ran names
	xappRnibMock.CreateGnb("RAN_NAME_1", entities.ConnectionStatus_CONNECTED)
	xappRnibMock.CreateGnb("RAN_NAME_11", entities.ConnectionStatus_CONNECTED)
	xappRnibMock.CreateGnb("RAN_NAME_2", entities.ConnectionStatus_CONNECTED)

	xappRnibMock.CreateGnb("gnb_208_092_303030", entities.ConnectionStatus_CONNECTED) // This same value is used in gnb simulator!
	xappRnibMock.CreateGnb("gnb_208_092_303030", entities.ConnectionStatus_DISCONNECTED)

	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_UNKNOWN_CONNECTION_STATUS)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_CONNECTED_SETUP_FAILED)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_CONNECTING)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_CONNECTED)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_SHUTTING_DOWN)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_SHUT_DOWN)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_DISCONNECTED)
	xappRnibMock.CreateGnb("gnb_369_11105_aaaaa3", entities.ConnectionStatus_UNDER_RESET)

	mainCtrl.c.e2IfState.ReadE2ConfigurationFromRnib()
	mainCtrl.c.e2IfState.SubscribeChannels()
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_UNKNOWN_CONNECTION_STATUS", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_CONNECTED_SETUP_FAILED", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_CONNECTING", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_CONNECTED", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_SHUTTING_DOWN", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_DISCONNECTED", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
	if err := xappRnibMock.XappRnibStoreAndPublish("RAN_CONNECTION_STATUS_CHANGE", "gnb_369_11105_aaaaa3_UNDER_RESET", "key1", "data1"); err != nil {
		t.Errorf("XappRnibStoreAndPublish failed: %v", err)
	}
}

func (x *XappRnibMock) CreateGnb(gnbId string, connectionStatus entities.ConnectionStatus) {

	xapp.Logger.Debug("XappRnibMock: CreateGnb() gnbId=%v, ConnectionStatus=%v", gnbId, connectionStatus)
	nb := xapp.RNIBNodebInfo{}
	nb.NodeType = xapp.RNIBNodeGNB
	nb.ConnectionStatus = connectionStatus
	if nb.ConnectionStatus < 0 || nb.ConnectionStatus > 6 {
		xapp.Logger.Error("XappRnibMock: CreateGnb() Incorrect connectionStatus=%v", nb.ConnectionStatus)
		return
	}
	nb.Ip = "localhost"
	nb.Port = 5656
	gnb := xapp.RNIBGnb{}

	gnb.ServedNrCells = nil
	nb.Configuration = &xapp.RNIBNodebInfoGnb{Gnb: &gnb}
	nbIdentity := &xapp.RNIBNbIdentity{
		InventoryName: gnbId,
		GlobalNbId: &xapp.RNIBGlobalNbId{
			PlmnId: "001EF5",
			NbId:   "0045FE50",
		},
	}

	err := xappRnibMock.XappRnibSaveNodeb(nbIdentity, &nb)
	if err != nil {
		xapp.Logger.Error("XappRnibMock: XappRnibSaveNodeb() failed. Error: %v", err)
	}
}

func (x *XappRnibMock) XappRnibSaveNodeb(nbIdentity *xapp.RNIBNbIdentity, nodeb *xapp.RNIBNodebInfo) xapp.RNIBIRNibError {

	xapp.Logger.Debug("XappRnibMock: XappRnibSaveNodeb() inventoryName=%v, ConnectionStatus=%v", nbIdentity.InventoryName, nodeb.ConnectionStatus)
	x.Mutex.Lock()
	defer x.Mutex.Unlock()
	x.nbIdentityMap[nbIdentity.InventoryName] = *nbIdentity
	x.RNIBNodebInfoMap[nbIdentity.InventoryName] = *nodeb
	return nil
}

func (x *XappRnibMock) XappRnibGetNodeb(inventoryName string) (*xapp.RNIBNodebInfo, xapp.RNIBIRNibError) {

	x.Mutex.Lock()
	defer x.Mutex.Unlock()
	xapp.Logger.Debug("XappRnibMock: XappRnibGetNodeb() inventoryName=%v", inventoryName)
	nodebInfo, ok := x.RNIBNodebInfoMap[inventoryName]
	if ok {
		return &nodebInfo, nil
	} else {
		return nil, fmt.Errorf("XappRnibMock: XappRnibGetNodeb() failed: inventoryName=%s:", inventoryName)
	}
}

func (x *XappRnibMock) XappRnibSubscribe(cb func(string, ...string), channel string) error {

	if x.RnibSubscription.Channel == "RAN_CONNECTION_STATUS_CHANGE" {
		xapp.Logger.Debug("XappRnibMock: RAN_CONNECTION_STATUS_CHANGE channel already subscribed")
		return nil
	}
	if x.RnibSubscription.Channel == "" {
		x.RnibSubscription.cb = cb
		x.RnibSubscription.Channel = channel
		xapp.Logger.Debug("XappRnibMock: RAN_CONNECTION_STATUS_CHANGE subscribed")
		return nil
	} else {
		return fmt.Errorf("XappRnibMock: Invalid channel/topic to subscribe: channel = %s", channel)
	}
}

func (x *XappRnibMock) XappRnibGetListGnbIds() ([]*xapp.RNIBNbIdentity, xapp.RNIBIRNibError) {

	xapp.Logger.Debug("XappRnibMock: XappRnibGetListGnbIds()")
	x.Mutex.Lock()
	defer x.Mutex.Unlock()
	var nbIdentities []*xapp.RNIBNbIdentity
	for _, nbIdentity := range x.nbIdentityMap {
		newNbIdentity := entities.NbIdentity{}
		newNbIdentity = nbIdentity
		nbIdentities = append(nbIdentities, &newNbIdentity)
	}
	xapp.Logger.Debug("XappRnibMock: XappRnibGetListGnbIds(). len(nbIdentities) = %v", len(nbIdentities))
	return nbIdentities, nil
}

func (x *XappRnibMock) XappRnibStoreAndPublish(channel string, event string, pairs ...interface{}) error {

	x.Mutex.Lock()
	defer x.Mutex.Unlock()
	xapp.Logger.Debug("XappRnibMock: Change published. channel=%s, event=%s", channel, event)
	if channel != "RAN_CONNECTION_STATUS_CHANGE" || channel == "" || event == "" {
		xapp.Logger.Debug("XappRnibMock: Invalid change published. channel=%s, event=%s", channel, event)
	}

	nbId, connectionStatus, err := ExtratNbIdAndConnectionStatus(event)
	if err != nil {
		xapp.Logger.Error("XappRnibMock: ExtratNbIdAndConnectionStatus. Err=%s", err)
	}

	nbIdentity, ok := x.nbIdentityMap[nbId]
	if ok {
		nbIdentity.ConnectionStatus = connectionStatus
	}

	if x.RnibSubscription.cb != nil {
		x.RnibSubscription.cb(channel, event)
	} else {
		xapp.Logger.Error("XappRnibMock: x.RnibSubscription.cb == nil")
	}
	return nil
}

func ExtratNbIdAndConnectionStatus(s string) (string, entities.ConnectionStatus, error) {

	var connectionStatus entities.ConnectionStatus
	var nbId string
	if strings.Contains(s, "_UNKNOWN_CONNECTION_STATUS") {
		connectionStatus = entities.ConnectionStatus_UNKNOWN_CONNECTION_STATUS
		splitStringTbl := strings.Split(s, "_UNKNOWN_CONNECTION_STATUS")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_CONNECTED") {
		connectionStatus = entities.ConnectionStatus_CONNECTED
		splitStringTbl := strings.Split(s, "_CONNECTED")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_DISCONNECTED") {
		connectionStatus = entities.ConnectionStatus_DISCONNECTED
		splitStringTbl := strings.Split(s, "_DISCONNECTED")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_CONNECTED_SETUP_FAILED") {
		connectionStatus = entities.ConnectionStatus_CONNECTED_SETUP_FAILED
		splitStringTbl := strings.Split(s, "_CONNECTED_SETUP_FAILED")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_CONNECTING") {
		connectionStatus = entities.ConnectionStatus_CONNECTING
		splitStringTbl := strings.Split(s, "_CONNECTING")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_SHUTTING_DOWN") {
		connectionStatus = entities.ConnectionStatus_SHUTTING_DOWN
		splitStringTbl := strings.Split(s, "_SHUTTING_DOWN")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_SHUT_DOWN") {
		connectionStatus = entities.ConnectionStatus_SHUT_DOWN
		splitStringTbl := strings.Split(s, "_SHUT_DOWN")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_UNDER_RESET") {
		connectionStatus = entities.ConnectionStatus_UNDER_RESET
		splitStringTbl := strings.Split(s, "_UNDER_RESET")
		nbId = splitStringTbl[0]
	} else {
		return "", 0, fmt.Errorf("XappRnibMock: Invalid connection status. %s", s)
	}
	if len(nbId) == 0 {
		return "", 0, fmt.Errorf("ExtractNbiIdFromString(): len(nbId) == 0 ")
	}
	return nbId, connectionStatus, nil
}
