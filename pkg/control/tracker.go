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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"sync"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Tracker struct {
	mutex                sync.Mutex
	transactionXappTable map[TransactionXappKey]*TransactionXapp
	transSeq             uint64
}

func (t *Tracker) Init() {
	t.transactionXappTable = make(map[TransactionXappKey]*TransactionXapp)
}

func (t *Tracker) initTransaction(transBase *Transaction) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	transBase.EventChan = make(chan interface{})
	transBase.tracker = t
	transBase.Seq = t.transSeq
	t.transSeq++
}

func (t *Tracker) NewSubsTransaction(subs *Subscription) *TransactionSubs {
	trans := &TransactionSubs{}
	trans.Meid = subs.GetMeid()
	t.initTransaction(&trans.Transaction)
	xapp.Logger.Debug("CREATE %s", trans.String())
	return trans
}

func (t *Tracker) NewXappTransaction(
	endpoint *xapp.RmrEndpoint,
	xid string,
	requestId e2ap.RequestId,
	meid *xapp.RMRMeid) *TransactionXapp {

	trans := &TransactionXapp{}
	trans.XappKey = &TransactionXappKey{requestId.Id, *endpoint, xid}
	trans.Meid = meid
	trans.RequestId = requestId
	t.initTransaction(&trans.Transaction)
	xapp.Logger.Debug("CREATE %s", trans.String())
	return trans
}

func (t *Tracker) Track(trans *TransactionXapp) error {

	if trans.GetEndpoint() == nil {
		err := fmt.Errorf("Tracker: No valid endpoint given in %s", trans.String())
		return err
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	theKey := *trans.XappKey

	if othtrans, ok := t.transactionXappTable[theKey]; ok {
		err := fmt.Errorf("Tracker: %s is ongoing, not tracking %s", othtrans, trans)
		return err
	}

	trans.tracker = t
	t.transactionXappTable[theKey] = trans
	xapp.Logger.Debug("Tracker: Append %s", trans.String())
	//xapp.Logger.Debug("Tracker: transtable=%v", t.transactionXappTable)
	return nil
}

func (t *Tracker) UnTrackTransaction(xappKey TransactionXappKey) (*TransactionXapp, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if trans, ok2 := t.transactionXappTable[xappKey]; ok2 {
		xapp.Logger.Debug("Tracker: Remove %s", trans.String())
		delete(t.transactionXappTable, xappKey)
		//xapp.Logger.Debug("Tracker: transtable=%v", t.transactionXappTable)
		return trans, nil
	}
	return nil, fmt.Errorf("Tracker: No record %v", xappKey)
}
