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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func (c *Control) TestRestHandler(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Debug("RESTTestRestHandler() called")

	pathParams := mux.Vars(r)
	s := pathParams["testId"]

	// This can be used to delete single subscription from db
	if contains := strings.Contains(s, "deletesubid="); contains == true {
		var splits = strings.Split(s, "=")
		if subId, err := strconv.ParseInt(splits[1], 10, 64); err == nil {
			xapp.Logger.Debug("RemoveSubscriptionFromSdl() called. subId = %v", subId)
			err := c.RemoveSubscriptionFromSdl(uint32(subId))
			if err != nil {
				xapp.Logger.Error("c.RemoveSubscriptionFromSdl failure: %s", err.Error())
			}
			return
		}
	}

	// This can be used to remove all subscriptions db from
	if s == "emptydb" {
		xapp.Logger.Debug("RemoveAllSubscriptionsFromSdl() called")
		err := c.RemoveAllSubscriptionsFromSdl()
		if err != nil {
			xapp.Logger.Error("RemoveAllSubscriptionsFromSdl() RemoveAllSubscriptionsFromSdl() failure: %s", err.Error())
		}
		err = c.RemoveAllRESTSubscriptionsFromSdl()
		if err != nil {
			xapp.Logger.Error("RemoveAllRESTSubscriptionsFromSdl() RemoveAllSubscriptionsFromSdl() failure: %s", err.Error())
		}
		return
	}

	// This is meant to cause submgr's restart in testing
	if s == "restart" {
		xapp.Logger.Debug("os.Exit(1) called")
		os.Exit(1)
	}

	xapp.Logger.Debug("Unsupported rest command received %s", s)
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) GetAllRestSubscriptions(w http.ResponseWriter, r *http.Request) {

	// Get all REST Subscriptions in subscription manager
	xapp.Logger.Debug("GetAllRestSubscriptions() called")
	_, err := w.Write(c.registry.GetAllRestSubscriptionsJson())
	if err != nil {
		xapp.Logger.Error("GetAllRestSubscriptions() w.Write failure: %s", err.Error())
	}
}

func (c *Control) GetAllE2Nodes(w http.ResponseWriter, r *http.Request) {

	// Get all E2Nodes in subscription manager
	xapp.Logger.Debug("GetAllE2Nodes() called")
	_, err := w.Write(c.e2IfState.GetE2NodesJson())
	if err != nil {
		xapp.Logger.Error("w.Write failure: %s", err.Error())
	}
}

func (c *Control) GetAllE2NodeRestSubscriptions(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Debug("GetAllE2NodeRestSubscriptions() called: Req= %v", r.URL.Path)

	// Get all REST Subscriptions of a E2Node
	pathParams := mux.Vars(r)
	ranName := pathParams["ranName"]
	xapp.Logger.Debug("GetAllE2NodeRestSubscriptions() ranName=%s", ranName)
	if ranName != "" {
		_, err := w.Write(c.registry.GetAllE2NodeRestSubscriptionsJson(ranName))
		if err != nil {
			xapp.Logger.Error("GetAllE2NodeRestSubscriptions() w.Write failure: %s", err.Error())
		}
	} else {
		xapp.Logger.Debug("GetAllE2NodeRestSubscriptions() Invalid path %s", ranName)
		w.WriteHeader(400) // Bad request
	}
}

func (c *Control) GetAllXapps(w http.ResponseWriter, r *http.Request) {

	// Get all xApps in subscription manager
	xapp.Logger.Debug("GetAllXapps() called: Req= %v", r.URL.Path)
	_, err := w.Write(c.registry.GetAllXappsJson())
	if err != nil {
		xapp.Logger.Error("GetAllXapps() w.Write failure: %s", err.Error())
	}
}

func (c *Control) GetAllXappRestSubscriptions(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Debug("GetAllXappRestSubscriptions() called")

	// Get all REST Subscriptions of a xApp
	pathParams := mux.Vars(r)
	xappServiceName := pathParams["xappServiceName"]
	xapp.Logger.Debug("GetAllXappRestSubscriptions() xappServiceName=%s", xappServiceName)
	if xappServiceName != "" {
		_, err := w.Write(c.registry.GetAllXappRestSubscriptionsJson(xappServiceName))
		if err != nil {
			xapp.Logger.Error("GetAllXappRestSubscriptions() w.Write failure: %s", err.Error())
		}
	} else {
		xapp.Logger.Debug("GetAllXappRestSubscriptions() Invalid path %s", xappServiceName)
		w.WriteHeader(400) // Bad request
	}
}

func (c *Control) DeleteAllE2nodeSubscriptions(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Debug("DeleteAllE2nodeSubscriptions() called: Req= %v", r.URL.Path)

	// Delete all REST Subscriptions of a E2Node
	pathParams := mux.Vars(r)
	ranName := pathParams["ranName"]
	xapp.Logger.Debug("DeleteE2nodeSubscriptions() ranName=%s", ranName)
	if ranName == "" {
		w.WriteHeader(400) // Bad request
	}
	nbIds := c.e2IfState.GetAllE2Nodes()
	ranName, ok := nbIds[ranName]
	if ok {
		restSubscriptions := c.registry.GetAllE2NodeRestSubscriptions(ranName)
		for restSubsId, _ := range restSubscriptions {
			c.RESTSubscriptionDeleteHandler(restSubsId)
		}
		return
	} else {
		w.WriteHeader(404) // Not found
	}
}

func (c *Control) DeleteAllXappSubscriptions(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Debug("DeleteAllXappSubscriptions() called: Req= %v", r.URL.Path)

	// Delete all REST Subscriptions of a xApp
	pathParams := mux.Vars(r)
	xappServiceName := pathParams["xappServiceName"]
	xapp.Logger.Debug("DeleteAllXappSubscriptions() ranName=%s", xappServiceName)
	if xappServiceName == "" {
		w.WriteHeader(400) // Bad request
	}
	xapps := c.registry.GetAllXapps()
	_, ok := xapps[xappServiceName]
	if ok {
		XappRestSubscriptions := c.registry.GetAllXappRestSubscriptions(xappServiceName)
		for restSubsId, _ := range XappRestSubscriptions {
			c.RESTSubscriptionDeleteHandler(restSubsId)
		}
		return
	} else {
		w.WriteHeader(404) // Not found
	}
	w.WriteHeader(200) // OK
}

func (c *Control) GetE2Subscriptions(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Debug("GetE2Subscriptions() called: Req= %v", r.URL.Path)

	// Get all E2 subscriptions of a REST Subscription
	pathParams := mux.Vars(r)
	restId := pathParams["restId"]
	xapp.Logger.Debug("GetE2Subscriptions(): restId=%s", restId)
	if restId == "" {
		w.WriteHeader(400) // Bad request
	}

	e2Subscriptions, err := c.registry.GetE2SubscriptionsJson(restId)
	if err != nil {
		w.WriteHeader(404) // Not found
	} else {
		_, err := w.Write(e2Subscriptions)
		if err != nil {
			xapp.Logger.Error("GetE2Subscriptions() w.Write failure: %s", err.Error())
		}
	}
}
