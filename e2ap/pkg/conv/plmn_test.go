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

package conv

import (
	"bytes"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestPlmnId(t *testing.T) {
	var ident PlmnIdentity
	ident.Set("12345")
	if ident.Mcc != "123" {
		t.Errorf("TestPlmnId: mcc expect 123 was %s", ident.Mcc)
	}
	if ident.Mnc != "45" {
		t.Errorf("TestPlmnId: mnc expect 45 was %s", ident.Mnc)
	}

	ident.Set("123456")
	if ident.Mcc != "123" {
		t.Errorf("TestPlmnId: mcc expect 123 was %s", ident.Mcc)
	}
	if ident.Mnc != "456" {
		t.Errorf("TestPlmnId: mnc expect 456 was %s", ident.Mnc)
	}

}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestPlmnIdBcd1(t *testing.T) {

	var encident PlmnIdentityBcd
	encident.Mcc = "233"
	encident.Mnc = "50"
	encbuf := new(bytes.Buffer)
	retlen, err := encident.EncodeTo(encbuf)

	if err != nil {
		t.Errorf("TestPlmnIdBcd1: EncodeTo err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdBcd1: EncodeTo expected len 3 got %d", retlen)
	}

	encdata := encbuf.Bytes()

	if encdata[0] != 0x23 {
		t.Errorf("TestPlmnIdBcd1: encident.val[0] expected 0x23 got 0x%x", encdata[0])
	}

	if encdata[1] != 0x3f {
		t.Errorf("TestPlmnIdBcd1: encident.val[1] expected 0x3f got 0x%x", encdata[1])
	}

	if encdata[2] != 0x50 {
		t.Errorf("TestPlmnIdBcd1: encident.val[2] expected 0x50 got 0x%x", encdata[2])
	}

	var decident PlmnIdentityBcd
	decbuf := []byte{0x23, 0x3f, 0x50}

	reader := bytes.NewReader(decbuf)
	retlen, err = decident.DecodeFrom(reader)

	if err != nil {
		t.Errorf("TestPlmnIdBcd1: DecodeFrom err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdBcd1: DecodeFrom expected len 3 got %d", retlen)
	}

	if decident.Mcc != "233" {
		t.Errorf("TestPlmnIdBcd1: mcc expected 233 got %s", decident.Mcc)
	}
	if decident.Mnc != "50" {
		t.Errorf("TestPlmnIdBcd1: mnc expected 50 got %s", decident.Mnc)
	}
}

func TestPlmnIdBcd2(t *testing.T) {

	var encident PlmnIdentityBcd

	encident.Mcc = "233"
	encident.Mnc = "550"
	encbuf := new(bytes.Buffer)
	retlen, err := encident.EncodeTo(encbuf)

	if err != nil {
		t.Errorf("TestPlmnIdBcd2: EncodeTo err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdBcd2: EncodeTo expected len 3 got %d", retlen)
	}

	encdata := encbuf.Bytes()

	if encdata[0] != 0x23 {
		t.Errorf("TestPlmnIdBcd2: encident.val[0] expected 0x23 got 0x%x", encdata[0])
	}

	if encdata[1] != 0x35 {
		t.Errorf("TestPlmnIdBcd1: encident.val[1] expected 0x35 got 0x%x", encdata[1])
	}

	if encdata[2] != 0x50 {
		t.Errorf("TestPlmnIdBcd2: encident.val[2] expected 0x50 got 0x%x", encdata[2])
	}

	var decident PlmnIdentityBcd
	decbuf := []byte{0x23, 0x35, 0x50}

	reader := bytes.NewReader(decbuf)
	retlen, err = decident.DecodeFrom(reader)

	if err != nil {
		t.Errorf("TestPlmnIdBcd2: DecodeFrom err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdBcd2: DecodeFrom expected len 3 got %d", retlen)
	}

	if decident.Mcc != "233" {
		t.Errorf("TestPlmnIdBcd2: mcc expected 233 got %s", decident.Mcc)
	}
	if decident.Mnc != "550" {
		t.Errorf("TestPlmnIdBcd2: mnc expected 550 got %s", decident.Mnc)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestPlmnIdTbcd1(t *testing.T) {

	var encident PlmnIdentityTbcd
	encident.Mcc = "233"
	encident.Mnc = "50"
	encbuf := new(bytes.Buffer)
	retlen, err := encident.EncodeTo(encbuf)

	if err != nil {
		t.Errorf("TestPlmnIdTbcd1: EncodeTo err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdTbcd1: EncodeTo expected len 3 got %d", retlen)
	}

	encdata := encbuf.Bytes()

	if encdata[0] != 0x32 {
		t.Errorf("TestPlmnIdTbcd1: encident.val[0] expected 0x32 got 0x%x", encdata[0])
	}

	if encdata[1] != 0xf3 {
		t.Errorf("TestPlmnIdTbcd1: encident.val[1] expected 0xf3 got 0x%x", encdata[1])
	}

	if encdata[2] != 0x05 {
		t.Errorf("TestPlmnIdTbcd1: encident.val[2] expected 0x05 got 0x%x", encdata[2])
	}

	var decident PlmnIdentityTbcd
	decbuf := []byte{0x32, 0xf3, 0x05}

	reader := bytes.NewReader(decbuf)
	retlen, err = decident.DecodeFrom(reader)

	if err != nil {
		t.Errorf("TestPlmnIdTbcd1: DecodeFrom err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdTbcd1: DecodeFrom expected len 3 got %d", retlen)
	}

	if decident.Mcc != "233" {
		t.Errorf("TestPlmnIdTbcd1: mcc expected 233 got %s", decident.Mcc)
	}
	if decident.Mnc != "50" {
		t.Errorf("TestPlmnIdTbcd1: mnc expected 50 got %s", decident.Mnc)
	}
}

func TestPlmnIdTbcd2(t *testing.T) {

	var encident PlmnIdentityTbcd

	encident.Mcc = "233"
	encident.Mnc = "550"
	encbuf := new(bytes.Buffer)
	retlen, err := encident.EncodeTo(encbuf)

	if err != nil {
		t.Errorf("TestPlmnIdTbcd2: EncodeTo err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdTbcd2: EncodeTo expected len 3 got %d", retlen)
	}

	encdata := encbuf.Bytes()

	if encdata[0] != 0x32 {
		t.Errorf("TestPlmnIdTbcd2: encident.val[0] expected 0x32 got 0x%x", encdata[0])
	}

	if encdata[1] != 0x03 {
		t.Errorf("TestPlmnIdTbcd1: encident.val[1] expected 0x03 got 0x%x", encdata[1])
	}

	if encdata[2] != 0x55 {
		t.Errorf("TestPlmnIdTbcd2: encident.val[2] expected 0x55 got 0x%x", encdata[2])
	}

	var decident PlmnIdentityTbcd
	decbuf := []byte{0x32, 0x03, 0x55}

	reader := bytes.NewReader(decbuf)
	retlen, err = decident.DecodeFrom(reader)

	if err != nil {
		t.Errorf("TestPlmnIdTbcd2: DecodeFrom err %s", err.Error())
	}
	if retlen != 3 {
		t.Errorf("TestPlmnIdTbcd2: DecodeFrom expected len 3 got %d", retlen)
	}

	if decident.Mcc != "233" {
		t.Errorf("TestPlmnIdTbcd2: mcc expected 233 got %s", decident.Mcc)
	}
	if decident.Mnc != "550" {
		t.Errorf("TestPlmnIdTbcd2: mnc expected 550 got %s", decident.Mnc)
	}
}
