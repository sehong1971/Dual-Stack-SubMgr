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
	"testing"

	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
)

func TestRestEndpointOk(t *testing.T) {

	var clientEndpoint models.SubscriptionParamsClientEndpoint
	var httpPort int64
	var rmrPort int64
	var host string

	httpPort = 8080
	rmrPort = 4560
	host = "service-ricxapp-xappname-http.ricxapp"

	clientEndpoint.HTTPPort = &httpPort
	clientEndpoint.RMRPort = &rmrPort
	clientEndpoint.Host = host

	expectedHttpEndpoint := "service-ricxapp-xappname-http.ricxapp:8080"
	expectedRmrEndpoint := "service-ricxapp-xappname-rmr.ricxapp:4560"

	httpEndPoint, rmrEndPoint, err := ConstructEndpointAddresses(clientEndpoint)

	if err != nil {
		t.Errorf("Mismatching return value: %s - ecpected NIL", err)
	}
	if httpEndPoint != expectedHttpEndpoint {
		t.Errorf("Mismatching httpEndpoint: %s - expected %s", httpEndPoint, expectedHttpEndpoint)
	}
	if rmrEndPoint != expectedRmrEndpoint {
		t.Errorf("Mismatching httpEndpoint: %s - expected %s", httpEndPoint, expectedHttpEndpoint)
	}
}

func TestRestEndpointNoHttpPort(t *testing.T) {

	var clientEndpoint models.SubscriptionParamsClientEndpoint
	var httpPort int64
	var rmrPort int64
	var host string

	httpPort = 0
	rmrPort = 4561
	host = "service-ricxapp-xappname-http.ricxapp"

	clientEndpoint.HTTPPort = &httpPort
	clientEndpoint.RMRPort = &rmrPort
	clientEndpoint.Host = host

	expectedHttpEndpoint := ""
	expectedRmrEndpoint := "service-ricxapp-xappname-rmr.ricxapp:4561"

	httpEndPoint, rmrEndPoint, err := ConstructEndpointAddresses(clientEndpoint)

	if err != nil {
		t.Errorf("Mismatching return value: %s - ecpected NIL", err)
	}
	if httpEndPoint != expectedHttpEndpoint {
		t.Errorf("Mismatching httpEndpoint: %s - ecpected %s", httpEndPoint, expectedHttpEndpoint)
	}
	if rmrEndPoint != expectedRmrEndpoint {
		t.Errorf("Mismatching httpEndpoint: %s - ecpected %s", httpEndPoint, expectedHttpEndpoint)
	}
}

func TestRestEndpointNok(t *testing.T) {

	var clientEndpoint models.SubscriptionParamsClientEndpoint
	var httpPort int64
	var rmrPort int64
	var host string

	httpPort = 0
	rmrPort = 0
	host = "service-ricxapp-xappname-http.ricxapp"

	clientEndpoint.HTTPPort = &httpPort
	clientEndpoint.RMRPort = &rmrPort
	clientEndpoint.Host = host

	_, _, err := ConstructEndpointAddresses(clientEndpoint)

	if err == nil {
		t.Errorf("Mismatching return value: - expected ERR but got NIL")
	}
}

func TestRestEndpointNoHttpInHost(t *testing.T) {

	var clientEndpoint models.SubscriptionParamsClientEndpoint
	var httpPort int64
	var rmrPort int64
	var host string

	httpPort = 8080
	rmrPort = 4562
	host = "service-ricxapp-xappname.ricxapp"

	clientEndpoint.HTTPPort = &httpPort
	clientEndpoint.RMRPort = &rmrPort
	clientEndpoint.Host = host

	expectedHttpEndpoint := "service-ricxapp-xappname.ricxapp:8080"
	expectedRmrEndpoint := "service-ricxapp-xappname.ricxapp:4562"

	httpEndPoint, rmrEndPoint, err := ConstructEndpointAddresses(clientEndpoint)

	if err != nil {
		t.Errorf("Mismatching return value: %s - ecpected NIL", err)
	}
	if httpEndPoint != expectedHttpEndpoint {
		t.Errorf("Mismatching httpEndpoint: %s - ecpected %s", httpEndPoint, expectedHttpEndpoint)
	}
	if rmrEndPoint != expectedRmrEndpoint {
		t.Errorf("Mismatching httpEndpoint: %s - ecpected %s", httpEndPoint, expectedHttpEndpoint)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
