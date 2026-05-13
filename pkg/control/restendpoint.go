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
	"strconv"
	"strings"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func ConstructEndpointAddresses(clientEndpoint models.SubscriptionParamsClientEndpoint) (string, string, error) {

	var HTTP_port int64 = *clientEndpoint.HTTPPort
	var RMR_port int64 = *clientEndpoint.RMRPort
	var host string = clientEndpoint.Host
	var xAppHTTPEndPoint string
	var xAppRMREndPoint string

	if host == "" {
		err := fmt.Errorf("ClientEndpoint provided without HOST name")
		return "", "", err
	}

	if HTTP_port == 0 && RMR_port == 0 {
		err := fmt.Errorf("ClientEndpoint provided without PORT numbers")
		return "INVALID_HTTP_ADDRESS:" + host + (string)(*clientEndpoint.HTTPPort),
			"INVALID_RMR_ADDRESS:" + host + (string)(*clientEndpoint.RMRPort),
			err
	}

	if *clientEndpoint.HTTPPort > 0 {
		xAppHTTPEndPoint = host + ":" + strconv.FormatInt(*clientEndpoint.HTTPPort, 10)
	}
	if *clientEndpoint.RMRPort > 0 {
		if i := strings.Index(host, "http"); i != -1 {
			host = strings.Replace(host, "http", "rmr", -1)
		}
		// START: grpc update
		// [추가] grpc -> rmr 변환 로직 추가
		if i := strings.Index(host, "grpc"); i != -1 {
			host = strings.Replace(host, "grpc", "rmr", -1)
		}
		// END: grpc update
		xAppRMREndPoint = host + ":" + strconv.FormatInt(*clientEndpoint.RMRPort, 10)
	}

	xapp.Logger.Debug("xAppHttpEndPoint=%v, xAppRrmEndPoint=%v", xAppHTTPEndPoint, xAppRMREndPoint)

	return xAppHTTPEndPoint, xAppRMREndPoint, nil
}
