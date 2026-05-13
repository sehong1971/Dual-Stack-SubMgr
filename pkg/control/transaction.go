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
	"strconv"
	"sync"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type TransactionIf interface {
	String() string
	Release()
	SendEvent(interface{}, time.Duration) (bool, bool)
	WaitEvent(time.Duration) (interface{}, bool)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type Transaction struct {
	mutex     sync.Mutex       //
	Seq       uint64           //transaction sequence
	tracker   *Tracker         //tracker instance
	Meid      *xapp.RMRMeid    //meid transaction related
	Mtype     int              //Encoded message type to be send
	Payload   *e2ap.PackedData //Encoded message to be send
	EventChan chan interface{}
}

func (t *Transaction) String() string {
	meidstr := "N/A"
	if t.Meid != nil {
		meidstr = t.Meid.String()
	}
	return "trans(" + strconv.FormatUint(uint64(t.Seq), 10) + "/" + meidstr + ")"
}

func (t *Transaction) SendEvent(event interface{}, waittime time.Duration) (bool, bool) {
	if waittime > 0 {
		select {
		case t.EventChan <- event:
			return true, false
		case <-time.After(waittime):
			return false, true
		}
		return false, false
	}
	t.EventChan <- event
	return true, false
}

func (t *Transaction) WaitEvent(waittime time.Duration) (interface{}, bool) {
	if waittime > 0 {
		select {
		case event := <-t.EventChan:
			return event, false
		case <-time.After(waittime):
			return nil, true
		}
	}
	event := <-t.EventChan
	return event, false
}

func (t *Transaction) GetMtype() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Mtype
}

func (t *Transaction) GetMeid() *xapp.RMRMeid {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.Meid != nil {
		return t.Meid
	}
	return nil
}

/*  // This function is not used. Commented out to get better test coverage result
func (t *Transaction) GetPayload() *e2ap.PackedData {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Payload
}
*/
//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type TransactionSubs struct {
	Transaction //
}

func (t *TransactionSubs) String() string {
	return "transsubs(" + t.Transaction.String() + ")"
}

func (t *TransactionSubs) Release() {
	t.mutex.Lock()
	xapp.Logger.Debug("RELEASE %s", t.String())
	t.tracker = nil
	t.mutex.Unlock()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type TransactionXappKey struct {
	InstanceID uint32
	xapp.RmrEndpoint
	Xid string // xapp xid in req
}

func (key *TransactionXappKey) String() string {
	return "transkey(" + key.RmrEndpoint.String() + "/" + key.Xid + ")"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type TransactionXapp struct {
	Transaction
	XappKey   *TransactionXappKey
	RequestId e2ap.RequestId
}

func (t *TransactionXapp) String() string {
	var transkey string = "transkey(N/A)"
	if t.XappKey != nil {
		transkey = t.XappKey.String()
	}
	return "transxapp(" + t.Transaction.String() + "/" + transkey + "/" + strconv.FormatUint(uint64(t.RequestId.InstanceId), 10) + ")"
}

func (t *TransactionXapp) GetEndpoint() *xapp.RmrEndpoint {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.XappKey != nil {
		return &t.XappKey.RmrEndpoint
	}
	return nil
}

func (t *TransactionXapp) GetXid() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.XappKey != nil {
		return t.XappKey.Xid
	}
	return ""
}

/*  // This function is not used. Commented out to get better test coverage result
func (t *TransactionXapp) GetSrc() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.XappKey != nil {
		return t.XappKey.RmrEndpoint.String()
	}
	return ""
}
*/
func (t *TransactionXapp) GetSubId() uint32 {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.RequestId.InstanceId
}

func (t *TransactionXapp) Release() {
	t.mutex.Lock()
	xapp.Logger.Debug("RELEASE %s", t.String())
	tracker := t.tracker
	xappkey := t.XappKey
	t.tracker = nil
	t.mutex.Unlock()

	if tracker != nil && xappkey != nil {
		_, err := tracker.UnTrackTransaction(*xappkey)
		if err != nil {
			xapp.Logger.Error("tracker.UnTrackTransaction() failed:%s", err.Error())
		}
	}
}
