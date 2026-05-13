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

/*
#cgo LDFLAGS: -le2ap_wrapper -le2ap
*/
import "C"

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

var packerif e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

func GetPackerIf() e2ap.E2APPackerIf {
	return packerif
}

func SetPackerIf(iface e2ap.E2APPackerIf) {
	packerif = iface
}

type E2ap struct {
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) SetASN1DebugPrintStatus(logLevel int) {
	e2ap_wrapper.SetASN1DebugPrintStatus(logLevel)
}

func (e *E2ap) SetE2IEOrderCheck(ieOrderCheck uint8) {
	e2ap_wrapper.SetE2IEOrderCheck(ieOrderCheck)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) FillSubscriptionReqMsgs(params interface{}, subreqList *e2ap.SubscriptionRequestList, restSubscription *RESTSubscription) error {
	xapp.Logger.Debug("FillSubscriptionReqMsgs")

	p := params.(*models.SubscriptionParams)

	for _, subscriptionDetail := range p.SubscriptionDetails {
		subReqMsg := e2ap.E2APSubscriptionRequest{}

		if p.RANFunctionID != nil {
			subReqMsg.FunctionId = (e2ap.FunctionId)(*p.RANFunctionID)
		}
		e2EventInstanceID := restSubscription.GetE2IdFromXappIdToE2Id(*subscriptionDetail.XappEventInstanceID)
		subReqMsg.RequestId = e2ap.RequestId{uint32(*subscriptionDetail.XappEventInstanceID), uint32(e2EventInstanceID)}

		if len(subscriptionDetail.EventTriggers) > 0 {
			for _, val := range subscriptionDetail.EventTriggers {
				subReqMsg.EventTriggerDefinition.Data.Data = append(subReqMsg.EventTriggerDefinition.Data.Data, byte(val))
			}
			subReqMsg.EventTriggerDefinition.Data.Length = uint64(len(subscriptionDetail.EventTriggers))
		}
		for _, actionToBeSetup := range subscriptionDetail.ActionToBeSetupList {
			actionToBeSetupItem := e2ap.ActionToBeSetupItem{}
			actionToBeSetupItem.ActionType = e2ap.E2AP_ActionTypeInvalid
			actionToBeSetupItem.ActionId = uint64(*actionToBeSetup.ActionID)

			actionToBeSetupItem.ActionType = e2ap.E2AP_ActionTypeStrMap[*actionToBeSetup.ActionType]
			actionToBeSetupItem.RicActionDefinitionPresent = true

			if len(actionToBeSetup.ActionDefinition) > 0 {
				for _, val := range actionToBeSetup.ActionDefinition {
					actionToBeSetupItem.ActionDefinitionChoice.Data.Data = append(actionToBeSetupItem.ActionDefinitionChoice.Data.Data, byte(val))
				}
				actionToBeSetupItem.ActionDefinitionChoice.Data.Length = uint64(len(actionToBeSetup.ActionDefinition))

			}
			if actionToBeSetup.SubsequentAction != nil {
				actionToBeSetupItem.SubsequentAction.Present = true
				actionToBeSetupItem.SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeStrMap[*actionToBeSetup.SubsequentAction.SubsequentActionType]
				actionToBeSetupItem.SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitStrMap[*actionToBeSetup.SubsequentAction.TimeToWait]
			}
			subReqMsg.ActionSetups = append(subReqMsg.ActionSetups, actionToBeSetupItem)
		}
		subreqList.E2APSubscriptionRequests = append(subreqList.E2APSubscriptionRequests, subReqMsg)
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) CheckActionNotAdmittedList(msgType int, actionNotAdmittedList e2ap.ActionNotAdmittedList, c *Control) ErrorInfo {

	var prefixString string
	var errorInfo ErrorInfo
	var actionNotAdmittedString string
	if len(actionNotAdmittedList.Items) > 0 {
		if msgType == xapp.RIC_SUB_RESP {
			prefixString = "RICSubscriptionResponse partially accepted:"
			c.UpdateCounter(cPartialSubRespFromE2)
		} else if msgType == xapp.RIC_SUB_FAILURE {
			prefixString = "RICSubscriptionFailure:"
		}
		jsonActionNotAdmittedList, err := json.Marshal(actionNotAdmittedList.Items)
		if err != nil {
			actionNotAdmittedString = "ActionNotAdmittedList > 0. Submgr json.Marshal error"
			xapp.Logger.Error("CheckActionNotAdmittedList() json.Marshal error %s", err.Error())
		} else {
			actionNotAdmittedString = "ActionNotAdmittedList: " + string(jsonActionNotAdmittedList)
		}
	}

	if msgType == xapp.RIC_SUB_FAILURE {
		prefixString = "RICSubscriptionFailure"
		err := fmt.Errorf("%s", prefixString)
		errorInfo.SetInfo(err.Error(), models.SubscriptionInstanceErrorSourceE2Node, "")
	}
	err := fmt.Errorf("%s %s", prefixString, actionNotAdmittedString)
	errorInfo.SetInfo(err.Error(), models.SubscriptionInstanceErrorSourceE2Node, "")
	return errorInfo
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) UnpackSubscriptionRequest(payload []byte) (*e2ap.E2APSubscriptionRequest, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	err, subReq := e2SubReq.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subReq, nil
}

func (c *E2ap) PackSubscriptionRequest(req *e2ap.E2APSubscriptionRequest) (int, *e2ap.PackedData, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	err, packedData := e2SubReq.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_REQ, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) UnpackSubscriptionResponse(payload []byte) (*e2ap.E2APSubscriptionResponse, error) {
	e2SubResp := packerif.NewPackerSubscriptionResponse()
	err, subResp := e2SubResp.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subResp, nil
}

func (e *E2ap) PackSubscriptionResponse(req *e2ap.E2APSubscriptionResponse) (int, *e2ap.PackedData, error) {
	e2SubResp := packerif.NewPackerSubscriptionResponse()
	err, packedData := e2SubResp.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_RESP, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) UnpackSubscriptionFailure(payload []byte) (*e2ap.E2APSubscriptionFailure, error) {
	e2SubFail := packerif.NewPackerSubscriptionFailure()
	err, subFail := e2SubFail.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subFail, nil
}

func (e *E2ap) PackSubscriptionFailure(req *e2ap.E2APSubscriptionFailure) (int, *e2ap.PackedData, error) {
	e2SubFail := packerif.NewPackerSubscriptionFailure()
	err, packedData := e2SubFail.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_FAILURE, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) UnpackSubscriptionDeleteRequest(payload []byte) (*e2ap.E2APSubscriptionDeleteRequest, error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	err, subDelReq := e2SubDelReq.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelReq, nil
}

func (e *E2ap) PackSubscriptionDeleteRequest(req *e2ap.E2APSubscriptionDeleteRequest) (int, *e2ap.PackedData, error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	err, packedData := e2SubDelReq.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_REQ, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) UnpackSubscriptionDeleteResponse(payload []byte) (*e2ap.E2APSubscriptionDeleteResponse, error) {
	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	err, subDelResp := e2SubDelResp.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelResp, nil
}

func (e *E2ap) PackSubscriptionDeleteResponse(req *e2ap.E2APSubscriptionDeleteResponse) (int, *e2ap.PackedData, error) {
	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	err, packedData := e2SubDelResp.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_RESP, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e *E2ap) UnpackSubscriptionDeleteFailure(payload []byte) (*e2ap.E2APSubscriptionDeleteFailure, error) {
	e2SubDelFail := packerif.NewPackerSubscriptionDeleteFailure()
	err, subDelFail := e2SubDelFail.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelFail, nil
}

/*
func (e *E2ap) PackSubscriptionDeleteFailure(req *e2ap.E2APSubscriptionDeleteFailure) (int, *e2ap.PackedData, error) {
	e2SubDelFail := packerif.NewPackerSubscriptionDeleteFailure()
	err, packedData := e2SubDelFail.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_FAILURE, packedData, nil
}
*/

//-----------------------------------------------------------------------------
// Changes to support "RIC_SUB_DEL_REQUIRED"
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionDeleteRequired(payload []byte) (*e2ap.SubscriptionDeleteRequiredList, error) {
	e2SubDelRequ := packerif.NewPackerSubscriptionDeleteRequired()
	err, subsToBeRemove := e2SubDelRequ.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subsToBeRemove, nil
}

func (c *E2ap) PackSubscriptionDeleteRequired(req *e2ap.SubscriptionDeleteRequiredList) (int, *e2ap.PackedData, error) {
	e2SubDelRequ := packerif.NewPackerSubscriptionDeleteRequired()
	err, packedData := e2SubDelRequ.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_REQUIRED, packedData, nil
}
