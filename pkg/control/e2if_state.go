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

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

type XappRnibIf struct {
	XappRnibInterface
}

func (x *XappRnibIf) XappRnibSubscribe(NotificationCb func(string, ...string), channel string) error {
	return xapp.Rnib.Subscribe(NotificationCb, channel)
}

func (x *XappRnibIf) XappRnibGetListGnbIds() ([]*xapp.RNIBNbIdentity, xapp.RNIBIRNibError) {
	return xapp.Rnib.GetListGnbIds()
}
func (x *XappRnibIf) XappRnibGetNodeb(inventoryName string) (*xapp.RNIBNodebInfo, xapp.RNIBIRNibError) {
	nodeInfo, err := xapp.Rnib.GetNodeb(inventoryName)
	return nodeInfo, err
}

func CreateXappRnibIfInstance() XappRnibInterface {
	return new(XappRnibIf)
}

type E2IfState struct {
	mutex         sync.Mutex
	control       *Control
	NbIdMap       map[string]string
	NbIdStatusMap map[string]string
}

func (e *E2IfState) Init(c *Control) {
	e.control = c
	e.NbIdMap = make(map[string]string, 0)
	e.NbIdStatusMap = make(map[string]string, 0)
	e.ReadE2ConfigurationFromRnib()
	err := e.SubscribeChannels()
	if err != nil {
		xapp.Logger.Error("Init(): SubscribeChannels() failed: %v", err)
	}
}

func (e *E2IfState) GetE2NodesJson() []byte {

	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Map contains something like this{"RAN_NAME_1":"RAN_NAME_1","RAN_NAME_11":"RAN_NAME_11","RAN_NAME_2":"RAN_NAME_2"}
	var ranNameList []string
	for _, ranName := range e.NbIdMap {
		ranNameList = append(ranNameList, ranName)
	}

	e2NodesJson, err := json.Marshal(ranNameList)
	if err != nil {
		xapp.Logger.Error("GetE2Node() json.Marshal error: %v", err)
	}
	return e2NodesJson
}

func (e *E2IfState) GetAllE2Nodes() map[string]string {

	e.mutex.Lock()
	defer e.mutex.Unlock()
	return e.NbIdMap
}

func (e *E2IfState) NotificationCb(ch string, events ...string) {

	xapp.Logger.Debug("SDL notification received from channel=%s, event=%v", ch, events[0])
	if len(events) == 0 {
		xapp.Logger.Error("Invalid SDL notification received: %d", len(events))
		return
	}

	if strings.Contains(events[0], "_CONNECTED") && !strings.Contains(events[0], "_CONNECTED_SETUP_FAILED") {
		e.control.UpdateCounter(cE2StateChangedToUp)
		nbId, err := ExtractNbiIdFromString(events[0])
		if err != nil {
			xapp.Logger.Error("NotificationCb CONNECTED len(nbId) == 0 ")
			return
		}
		xapp.Logger.Debug("E2 CONNECTED. NbId=%s", nbId)
		e.NbIdMap[nbId] = nbId
		e.NbIdStatusMap[nbId] = "CONNECTED"
	} else if strings.Contains(events[0], "_DISCONNECTED") {
		e.control.UpdateCounter(cE2StateChangedToDown)
		nbId, err := ExtractNbiIdFromString(events[0])
		if err != nil {
			xapp.Logger.Error("NotificationCb DISCONNECTED len(nbId) == 0 ")
			return
		}
		xapp.Logger.Debug("E2 DISCONNECTED. NbId=%s", nbId)
		if _, ok := e.NbIdMap[nbId]; ok {
			e.NbIdStatusMap[nbId] = "DISCONNECTED"
			delete(e.NbIdMap, nbId)
			e.control.registry.DeleteAllE2Subscriptions(nbId, e.control)
		}
	} else if strings.Contains(events[0], "_UNDER_RESET") {
		xapp.Logger.Debug("NotificationCb UNDER_RESET len(nbId) == 0 ")
		e.control.UpdateCounter(cE2StateUnderReset)
		nbId, err := ExtractNbiIdFromString(events[0])
		if err != nil {
			xapp.Logger.Error("NotificationCb _UNDER_RESET %v ", err)
			return
		}
		e.NbIdStatusMap[nbId] = "UNDER_RESET"
		xapp.Logger.Debug("E2 Under Reset. NbId=%s", nbId)
		if _, ok := e.NbIdMap[nbId]; ok {
			e.control.registry.DeleteAllE2Subscriptions(nbId, e.control)
		}
	}
}

func (e *E2IfState) SubscribeChannels() error {

	if err := e.control.e2IfStateDb.XappRnibSubscribe(e.NotificationCb, "RAN_CONNECTION_STATUS_CHANGE"); err != nil {
		xapp.Logger.Error("Sdl.SubscribeChannel failed: %v", err)
		return err
	}
	xapp.Logger.Debug("Subscription to RAN state changes done!")
	return nil
}

func (e *E2IfState) ReadE2ConfigurationFromRnib() {

	xapp.Logger.Debug("ReadE2ConfigurationFromRnib()")
	nbIdentities, err := e.control.e2IfStateDb.XappRnibGetListGnbIds()
	if err != nil || len(nbIdentities) == 0 {
		xapp.Logger.Debug("There are no active NodeBs available: %v", err)
		e.NbIdMap = make(map[string]string, 0)
		return
	}

	for _, nbIdentity := range nbIdentities {
		if e.isNodeBActive(nbIdentity.InventoryName) == false {
			if _, ok := e.NbIdMap[nbIdentity.InventoryName]; ok {
				delete(e.NbIdMap, nbIdentity.InventoryName)
				xapp.Logger.Debug("E2 connection DISCONNETED: %v", nbIdentity.InventoryName)

				// Delete all subscriptions related to InventoryName/nbId
				e.control.registry.DeleteAllE2Subscriptions(nbIdentity.InventoryName, e.control)
			}
			continue
		}

		if _, ok := e.NbIdMap[nbIdentity.InventoryName]; !ok {
			e.NbIdMap[nbIdentity.InventoryName] = nbIdentity.InventoryName
			xapp.Logger.Debug("E2 connection CONNECTED: %v", nbIdentity.InventoryName)
		}
	}
}

func (e *E2IfState) isNodeBActive(inventoryName string) bool {
	nodeInfo, err := e.control.e2IfStateDb.XappRnibGetNodeb(inventoryName)
	if err != nil {
		xapp.Logger.Error("GetNodeb() failed for inventoryName=%s: %v", inventoryName, err)
		return false
	}
	xapp.Logger.Debug("NodeB['%s'] connection status = %d", inventoryName, nodeInfo.ConnectionStatus)
	return nodeInfo.ConnectionStatus == 1
}

func (e *E2IfState) IsE2ConnectionUp(nbId *string) bool {

	if checkE2State == "false" {
		return true
	}

	if _, ok := e.NbIdMap[*nbId]; ok {
		return true
	} else {
		return false
	}
}

func (e *E2IfState) IsE2ConnectionUnderReset(nbId *string) bool {

	if status := e.NbIdStatusMap[*nbId]; status == "UNDER_RESET" {
		return true
	} else {
		return false
	}
}

func ExtractNbiIdFromString(s string) (string, error) {

	// Expected string formats are below
	// gnb_208_092_303030_CONNECTED
	// gnb_208_092_303030_DISCONNECTED
	// ...

	var nbId string
	var err error
	if strings.Contains(s, "_CONNECTED") {
		splitStringTbl := strings.Split(s, "_CONNECTED")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_DISCONNECTED") {
		splitStringTbl := strings.Split(s, "_DISCONNECTED")
		nbId = splitStringTbl[0]
	} else if strings.Contains(s, "_UNDER_RESET") {
		splitStringTbl := strings.Split(s, "_UNDER_RESET")
		nbId = splitStringTbl[0]
	}
	if len(nbId) == 0 {
		return "", fmt.Errorf("ExtractNbiIdFromString(): len(nbId) == 0 ")
	}
	return nbId, err
}
