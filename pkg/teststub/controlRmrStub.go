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
package teststub

import (
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strings"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrStubControl struct {
	RmrControl
	*xapp.RMRClient
	RecvChan chan *xapp.RMRParams
	Active   bool
	InitMsg  int
	CheckXid bool
}

func (tc *RmrStubControl) SetActive() {
	tc.Active = true
}

func (tc *RmrStubControl) IsActive() bool {
	return tc.Active
}

func (tc *RmrStubControl) SetCheckXid(val bool) {
	tc.CheckXid = val
}

func (tc *RmrStubControl) IsCheckXid() bool {
	return tc.CheckXid
}

func (tc *RmrStubControl) IsChanEmpty() bool {
	if len(tc.RecvChan) > 0 {
		return false
	}
	return true
}

func (tc *RmrStubControl) TestMsgChanEmpty(t *testing.T) {
	if tc.IsChanEmpty() == false {
		tc.TestError(t, "message channel not empty")
	}
}

func (tc *RmrStubControl) Init(desc string, srcId RmrSrcId, rtgSvc RmrRtgSvc, statDesc string, initMsg int) {
	tc.InitMsg = initMsg
	tc.Active = false
	tc.RecvChan = make(chan *xapp.RMRParams)
	tc.RmrControl.Init(desc, srcId, rtgSvc)

	tc.RMRClient = xapp.NewRMRClientWithParams(&xapp.RMRClientParams{
		StatDesc: statDesc,
		RmrData: xapp.PortData{
			Name:       "", // Not used currently
			Port:       int(srcId.Port),
			MaxSize:    65534,
			ThreadType: 0,
			LowLatency: false,
			FastAck:    false,
			//Policies:				// Not used currently
			MaxRetryOnFailure: 1,
		},
	})

	tc.RMRClient.SetReadyCB(tc.ReadyCB, nil)
	go tc.RMRClient.Start(tc)

	tc.WaitCB()
	allRmrStubs = append(allRmrStubs, tc)
}

func (tc *RmrStubControl) Consume(msg *xapp.RMRParams) (err error) {
	defer tc.RMRClient.Free(msg.Mbuf)

	cPay := append(msg.Payload[:0:0], msg.Payload...)
	msg.Payload = cPay
	msg.PayloadLen = len(cPay)

	if msg.Mtype == tc.InitMsg {
		tc.Debug("Testing message ignore %s", msg.String())
		tc.SetActive()
		return
	}

	if tc.IsCheckXid() == true && strings.Contains(msg.Xid, tc.GetDesc()) == false {
		tc.Debug("Ignore %s", msg.String())
		return
	}

	tc.Debug("Consume %s", msg.String())
	tc.PushMsg(msg)
	return
}

func (tc *RmrStubControl) PushMsg(msg *xapp.RMRParams) {
	tc.Debug("RmrStubControl PushMsg ... msg(%d) waiting", msg.Mtype)
	tc.RecvChan <- msg
	tc.Debug("RmrStubControl PushMsg ... done")
}

func (tc *RmrStubControl) WaitMsg(secs time.Duration) *xapp.RMRParams {
	tc.Debug("RmrStubControl WaitMsg ... waiting")
	if secs == 0 {
		msg := <-tc.RecvChan
		tc.Debug("RmrStubControl WaitMsg ... msg(%d) done", msg.Mtype)
		return msg
	}
	select {
	case msg := <-tc.RecvChan:
		tc.Debug("RmrStubControl WaitMsg ... msg(%d) done", msg.Mtype)
		return msg
	case <-time.After(secs * time.Second):
		tc.Debug("RmrStubControl WaitMsg ... timeout")
		return nil
	}
	tc.Debug("RmrStubControl WaitMsg ... error")
	return nil
}

var allRmrStubs []*RmrStubControl

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func RmrStubControlWaitAlive(seconds int, mtype int, rmr *xapp.RMRClient, tent *TestWrapper) bool {

	var dummyBuf []byte = make([]byte, 100)

	params := &xapp.RMRParams{}
	params.Mtype = mtype
	params.SubId = -1
	params.Payload = dummyBuf
	params.PayloadLen = 100
	params.Meid = &xapp.RMRMeid{RanName: "TESTPING"}
	params.Xid = "TESTPING"
	params.Mbuf = nil

	if len(allRmrStubs) == 0 {
		tent.Debug("No rmr stubs so no need to wait those to be alive")
		return true
	}
	status := false
	i := 1
	for ; i <= seconds*2 && status == false; i++ {

		tent.Debug("SEND TESTPING: %s", params.String())
		rmr.SendWithRetry(params, false, 0)

		status = true
		for _, val := range allRmrStubs {
			if val.IsActive() == false {
				status = false
				break
			}
		}
		if status == true {
			break
		}
		tent.Debug("Sleep 0.5 secs and try routes again")
		time.Sleep(500 * time.Millisecond)
	}

	if status == false {
		tent.Error("Could not initialize routes")
	}
	return status
}
