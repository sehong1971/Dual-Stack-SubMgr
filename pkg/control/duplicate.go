/*
==================================================================================
  Copyright (c) 2021 Nokia

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
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

type RetransEntry struct {
	restSubsId string
	startTime  time.Time
}

type DuplicateCtrl struct {
	mutex              sync.Mutex
	ongoingRequestMap  map[string]RetransEntry
	previousRequestMap map[string]string
	collCount          int
}

func (d *DuplicateCtrl) Init() {
	d.ongoingRequestMap = make(map[string]RetransEntry)
	d.previousRequestMap = make(map[string]string)
}

func (d *DuplicateCtrl) SetMd5sumFromLastOkRequest(restSubsId string, md5sum string) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if md5sum == "" {
		xapp.Logger.Error("Attempt to store empty md5sum for restubsId %s retransmission map skipped", restSubsId)
		return
	}

	err := d.removeOngoingTransaction(md5sum)
	if err != nil {
		xapp.Logger.Error("removeOngoingTransaction() failed:%s", err.Error())
	}

	prevRestSubsId, exists := d.previousRequestMap[md5sum]

	if exists {
		if prevRestSubsId != restSubsId {
			xapp.Logger.Error("Storing md5sum for a processed request for restSubsId %s md5sum %s over a previous restSubsId %s", restSubsId, md5sum, prevRestSubsId)
		} else {
			return
		}
	} else {
		xapp.Logger.Debug("Storing md5sum for a processed request for restSubsId %s md5sum %s", restSubsId, md5sum)
	}

	d.previousRequestMap[md5sum] = restSubsId
}

func (d *DuplicateCtrl) GetLastKnownRestSubsIdBasedOnMd5sum(md5sum string) (string, bool) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	if md5sum == "" {
		return "", false
	}

	m, e := d.previousRequestMap[md5sum]

	return m, e
}

func (d *DuplicateCtrl) DeleteLastKnownRestSubsIdBasedOnMd5sum(md5sum string) {

	d.mutex.Lock()
	defer d.mutex.Unlock()

	restSubsId, exists := d.previousRequestMap[md5sum]

	if !exists {
		if md5sum == "" {
			xapp.Logger.Debug("Attempted to delete a cached md5sum, md5sum not set yet")
		} else {
			xapp.Logger.Error("Attempted to delete a cached md5sum %s, but the value was not found", md5sum)
		}
	} else {
		xapp.Logger.Debug("Deleted a cached md5sum %s for restSubsId %s", md5sum, restSubsId)
		delete(d.previousRequestMap, md5sum)
	}
}

func CalculateRequestMd5sum(payload interface{}) (string, error) {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)

	if err := enc.Encode(payload); err != nil {
		xapp.Logger.Error("%s", err.Error())
		return "", err
	}

	hash := md5.Sum(data.Bytes())

	return hex.EncodeToString(hash[:]), nil
}

func (d *DuplicateCtrl) IsDuplicateToOngoingTransaction(restSubsId string, md5sum string) bool {

	if md5sum == "" {
		return false
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	entry, present := d.ongoingRequestMap[md5sum]

	if present {
		xapp.Logger.Debug("Collision detected. REST subs ID %s has ongoing transaction with md5sum : %s started at %s\n", entry.restSubsId, md5sum, entry.startTime.Format(time.ANSIC))
		d.collCount++
		return true
	}

	entry = RetransEntry{restSubsId: restSubsId, startTime: time.Now()}

	xapp.Logger.Debug("No collision detected against ongoing transaction. Added md5sum %s for restSubsId %s at %s\n", md5sum, entry.restSubsId, entry.startTime)

	d.ongoingRequestMap[md5sum] = entry

	return false
}

func (d *DuplicateCtrl) TransactionComplete(md5sum string) error {

	if md5sum == "" {
		return nil
	}

	d.mutex.Lock()
	defer d.mutex.Unlock()

	return d.removeOngoingTransaction(md5sum)
}

func (d *DuplicateCtrl) removeOngoingTransaction(md5sum string) error {

	entry, present := d.ongoingRequestMap[md5sum]

	if !present {
		xapp.Logger.Error("md5sum : %s NOT found from retransmission table", md5sum)
		return fmt.Errorf("Retransmission entry not found for md5sum %s", md5sum)
	}

	xapp.Logger.Debug("Releasing transaction duplicate blocker for %s, md5sum : %s\n", entry.restSubsId, md5sum)

	delete(d.ongoingRequestMap, md5sum)

	return nil
}
