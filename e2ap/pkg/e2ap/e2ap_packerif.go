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

package e2ap

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionRequestIf interface {
	Pack(*E2APSubscriptionRequest) (error, *PackedData)
	UnPack(msg *PackedData) (error, *E2APSubscriptionRequest)
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionResponseIf interface {
	Pack(*E2APSubscriptionResponse) (error, *PackedData)
	UnPack(msg *PackedData) (error, *E2APSubscriptionResponse)
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionFailureIf interface {
	Pack(*E2APSubscriptionFailure) (error, *PackedData)
	UnPack(msg *PackedData) (error, *E2APSubscriptionFailure)
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteRequestIf interface {
	Pack(*E2APSubscriptionDeleteRequest) (error, *PackedData)
	UnPack(msg *PackedData) (error, *E2APSubscriptionDeleteRequest)
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteResponseIf interface {
	Pack(*E2APSubscriptionDeleteResponse) (error, *PackedData)
	UnPack(msg *PackedData) (error, *E2APSubscriptionDeleteResponse)
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteFailureIf interface {
	Pack(*E2APSubscriptionDeleteFailure) (error, *PackedData)
	UnPack(msg *PackedData) (error, *E2APSubscriptionDeleteFailure)
	String() string
}

//-----------------------------------------------------------------------------
// Changes to support "RIC_SUB_DEL_REQUIRED"
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteRequiredIf interface {
	Pack(list *SubscriptionDeleteRequiredList) (error, *PackedData)
	UnPack(msg *PackedData) (error, *SubscriptionDeleteRequiredList)
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APPackerIf interface {
	NewPackerSubscriptionRequest() E2APMsgPackerSubscriptionRequestIf
	NewPackerSubscriptionResponse() E2APMsgPackerSubscriptionResponseIf
	NewPackerSubscriptionFailure() E2APMsgPackerSubscriptionFailureIf
	NewPackerSubscriptionDeleteRequest() E2APMsgPackerSubscriptionDeleteRequestIf
	NewPackerSubscriptionDeleteResponse() E2APMsgPackerSubscriptionDeleteResponseIf
	NewPackerSubscriptionDeleteFailure() E2APMsgPackerSubscriptionDeleteFailureIf
	NewPackerSubscriptionDeleteRequired() E2APMsgPackerSubscriptionDeleteRequiredIf
	//UnPack(*PackedData) (error, interface{})
	//Pack(interface{}, *PackedData) (error, *PackedData)
}
