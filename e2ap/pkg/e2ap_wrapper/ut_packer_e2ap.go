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

package e2ap_wrapper

import (
	"fmt"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
)

const (
	SUB_REQ         int = 1
	SUB_RESP        int = 2
	SUB_FAILURE     int = 3
	SUB_DEL_REQ     int = 4
	SUB_DEL_RESP    int = 5
	SUB_DEL_FAILURE int = 6
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var origPackerif e2ap.E2APPackerIf = NewAsn1E2Packer()

var allowAction = map[int]bool{
	SUB_REQ:         true,
	SUB_RESP:        true,
	SUB_FAILURE:     true,
	SUB_DEL_REQ:     true,
	SUB_DEL_RESP:    true,
	SUB_DEL_FAILURE: true,
}

func AllowE2apToProcess(mtype int, actionFail bool) {
	allowAction[mtype] = actionFail
}

type utMsgPackerSubscriptionRequest struct {
	e2apMsgPackerSubscriptionRequest
}

func (e2apMsg *utMsgPackerSubscriptionRequest) init() {
}

func (e2apMsg *utMsgPackerSubscriptionRequest) Pack(data *e2ap.E2APSubscriptionRequest) (error, *e2ap.PackedData) {
	if allowAction[SUB_REQ] {
		e2sub := origPackerif.NewPackerSubscriptionRequest()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT (%v)", allowAction[SUB_REQ]), nil
}

func (e2apMsg *utMsgPackerSubscriptionRequest) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionRequest) {
	if allowAction[SUB_REQ] {
		e2sub := origPackerif.NewPackerSubscriptionRequest()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT (%v)", allowAction[SUB_REQ]), nil
}

func (e2apMsg *utMsgPackerSubscriptionRequest) String() string {
	return "utMsgPackerSubscriptionRequest"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type utMsgPackerSubscriptionResponse struct {
	e2apMsgPackerSubscriptionResponse
}

func (e2apMsg *utMsgPackerSubscriptionResponse) init() {
}

func (e2apMsg *utMsgPackerSubscriptionResponse) Pack(data *e2ap.E2APSubscriptionResponse) (error, *e2ap.PackedData) {
	if allowAction[SUB_RESP] {
		e2sub := origPackerif.NewPackerSubscriptionResponse()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionResponse) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionResponse) {
	if allowAction[SUB_RESP] {
		e2sub := origPackerif.NewPackerSubscriptionResponse()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionResponse) String() string {
	return "utMsgPackerSubscriptionResponse"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type utMsgPackerSubscriptionFailure struct {
	e2apMsgPackerSubscriptionFailure
}

func (e2apMsg *utMsgPackerSubscriptionFailure) init() {
}

func (e2apMsg *utMsgPackerSubscriptionFailure) Pack(data *e2ap.E2APSubscriptionFailure) (error, *e2ap.PackedData) {
	if allowAction[SUB_FAILURE] {
		e2sub := origPackerif.NewPackerSubscriptionFailure()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionFailure) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionFailure) {
	if allowAction[SUB_FAILURE] {
		e2sub := origPackerif.NewPackerSubscriptionFailure()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionFailure) String() string {
	return "utMsgPackerSubscriptionFailure"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type utMsgPackerSubscriptionDeleteRequest struct {
	e2apMsgPackerSubscriptionDeleteRequest
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequest) init() {
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequest) Pack(data *e2ap.E2APSubscriptionDeleteRequest) (error, *e2ap.PackedData) {
	if allowAction[SUB_DEL_REQ] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteRequest()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT (%v)", allowAction[SUB_DEL_REQ]), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequest) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteRequest) {
	if allowAction[SUB_DEL_REQ] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteRequest()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT (%v)", allowAction[SUB_DEL_REQ]), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequest) String() string {
	return "utMsgPackerSubscriptionDeleteRequest"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type utMsgPackerSubscriptionDeleteResponse struct {
	e2apMsgPackerSubscriptionDeleteResponse
}

func (e2apMsg *utMsgPackerSubscriptionDeleteResponse) init() {

}

func (e2apMsg *utMsgPackerSubscriptionDeleteResponse) Pack(data *e2ap.E2APSubscriptionDeleteResponse) (error, *e2ap.PackedData) {
	if allowAction[SUB_DEL_RESP] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteResponse()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteResponse) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteResponse) {
	if allowAction[SUB_DEL_RESP] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteResponse()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteResponse) String() string {
	return "utMsgPackerSubscriptionDeleteResponse"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type utMsgPackerSubscriptionDeleteFailure struct {
	e2apMsgPackerSubscriptionDeleteFailure
}

func (e2apMsg *utMsgPackerSubscriptionDeleteFailure) init() {
}

func (e2apMsg *utMsgPackerSubscriptionDeleteFailure) Pack(data *e2ap.E2APSubscriptionDeleteFailure) (error, *e2ap.PackedData) {
	if allowAction[SUB_DEL_FAILURE] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteFailure()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteFailure) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteFailure) {
	if allowAction[SUB_DEL_FAILURE] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteFailure()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteFailure) String() string {
	return "utMsgPackerSubscriptionDeleteFailure"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type utMsgPackerSubscriptionDeleteRequired struct {
	e2apMsgPackerSubscriptionDeleteFailure
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequired) init() {
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequired) Pack(data *e2ap.SubscriptionDeleteRequiredList) (error, *e2ap.PackedData) {
	if allowAction[SUB_DEL_FAILURE] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteRequired()
		return e2sub.Pack(data)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

func (e2apMsg *utMsgPackerSubscriptionDeleteRequired) UnPack(msg *e2ap.PackedData) (error, *e2ap.SubscriptionDeleteRequiredList) {
	if allowAction[SUB_DEL_FAILURE] {
		e2sub := origPackerif.NewPackerSubscriptionDeleteRequired()
		return e2sub.UnPack(msg)
	}
	return fmt.Errorf("Error: Set to be fail by UT"), nil
}

//-----------------------------------------------------------------------------
// Public E2AP packer creators
//-----------------------------------------------------------------------------

type utAsn1E2APPacker struct{}

func (*utAsn1E2APPacker) NewPackerSubscriptionRequest() e2ap.E2APMsgPackerSubscriptionRequestIf {
	return &utMsgPackerSubscriptionRequest{}
}

func (*utAsn1E2APPacker) NewPackerSubscriptionResponse() e2ap.E2APMsgPackerSubscriptionResponseIf {
	return &utMsgPackerSubscriptionResponse{}
}

func (*utAsn1E2APPacker) NewPackerSubscriptionFailure() e2ap.E2APMsgPackerSubscriptionFailureIf {
	return &utMsgPackerSubscriptionFailure{}
}

func (*utAsn1E2APPacker) NewPackerSubscriptionDeleteRequest() e2ap.E2APMsgPackerSubscriptionDeleteRequestIf {
	return &utMsgPackerSubscriptionDeleteRequest{}
}

func (*utAsn1E2APPacker) NewPackerSubscriptionDeleteResponse() e2ap.E2APMsgPackerSubscriptionDeleteResponseIf {
	return &utMsgPackerSubscriptionDeleteResponse{}
}

func (*utAsn1E2APPacker) NewPackerSubscriptionDeleteFailure() e2ap.E2APMsgPackerSubscriptionDeleteFailureIf {
	return &utMsgPackerSubscriptionDeleteFailure{}
}

func (p *utAsn1E2APPacker) NewPackerSubscriptionDeleteRequired() e2ap.E2APMsgPackerSubscriptionDeleteRequiredIf {
	return &utMsgPackerSubscriptionDeleteRequired{}
}

func NewUtAsn1E2APPacker() e2ap.E2APPackerIf {
	return &utAsn1E2APPacker{}
}
