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
	"encoding/hex"
	"fmt"
	"os"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func TestBcdEven(t *testing.T) {

	bcd := NewBcd("0123456789??????")
	bcdbuf := bcd.Encode("123456")
	if len(bcdbuf) == 0 {
		t.Errorf("TestBcdEven: bcd Encode failed")
	}

	if bcdbuf[0] != 0x12 {
		t.Errorf("TestBcdEven: bcdbuf[0] expected 0x12 got 0x%x", bcdbuf[0])
	}

	if bcdbuf[1] != 0x34 {
		t.Errorf("TestBcdEven: bcdbuf[1] expected 0x34 got 0x%x", bcdbuf[1])
	}

	if bcdbuf[2] != 0x56 {
		t.Errorf("TestBcdEven: bcdbuf[2] expected 0x56 got 0x%x", bcdbuf[2])
	}

	hexdata := make([]byte, hex.EncodedLen(len(bcdbuf)))
	hex.Encode(hexdata, bcdbuf)
	fmt.Printf("TestBcdEven: 123456 encoded data [%s]\n", string(hexdata))

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("123456") {
		t.Errorf("TestBcdEven: bcd Decode failed: got %s expect %s", bcdstr, string("123456"))
	}

}

func TestBcdUnEven1(t *testing.T) {

	bcd := NewBcd("0123456789??????")
	bcdbuf := bcd.Encode("12345")
	if len(bcdbuf) == 0 {
		t.Errorf("TestBcdUnEven1: bcd Encode failed")
	}

	if bcdbuf[0] != 0x12 {
		t.Errorf("TestBcdEven: bcdbuf[0] expected 0x12 got 0x%x", bcdbuf[0])
	}

	if bcdbuf[1] != 0x34 {
		t.Errorf("TestBcdEven: bcdbuf[1] expected 0x34 got 0x%x", bcdbuf[1])
	}

	if bcdbuf[2] != 0x5f {
		t.Errorf("TestBcdEven: bcdbuf[2] expected 0x5f got 0x%x", bcdbuf[2])
	}

	hexdata := make([]byte, hex.EncodedLen(len(bcdbuf)))
	hex.Encode(hexdata, bcdbuf)
	fmt.Printf("TestBcdUnEven1: 12345 encoded data [%s]\n", string(hexdata))

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("12345?") {
		t.Errorf("TestBcdUnEven1: bcd Decode failed: got %s expect %s", bcdstr, string("12345?"))
	}
}

func TestBcdUnEven2(t *testing.T) {

	bcd := NewBcd("0123456789?????f")
	bcdbuf := bcd.Encode("12345f")
	if len(bcdbuf) == 0 {
		t.Errorf("TestBcdUnEven2: bcd Encode failed")
	}

	if bcdbuf[0] != 0x12 {
		t.Errorf("TestBcdEven: bcdbuf[0] expected 0x12 got 0x%x", bcdbuf[0])
	}

	if bcdbuf[1] != 0x34 {
		t.Errorf("TestBcdEven: bcdbuf[1] expected 0x34 got 0x%x", bcdbuf[1])
	}

	if bcdbuf[2] != 0x5f {
		t.Errorf("TestBcdEven: bcdbuf[2] expected 0x5f got 0x%x", bcdbuf[2])
	}

	hexdata := make([]byte, hex.EncodedLen(len(bcdbuf)))
	hex.Encode(hexdata, bcdbuf)
	fmt.Printf("TestBcdUnEven2: 12345f encoded data [%s]\n", string(hexdata))

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("12345f") {
		t.Errorf("TestBcdUnEven2: bcd Decode failed: got %s expect %s", bcdstr, string("12345f"))
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func TestTbcdEven(t *testing.T) {

	bcd := NewTbcd("0123456789??????")
	bcdbuf := bcd.Encode("123456")
	if len(bcdbuf) == 0 {
		t.Errorf("TestTbcdEven: bcd Encode failed")
	}

	if bcdbuf[0] != 0x21 {
		t.Errorf("TestBcdEven: bcdbuf[0] expected 0x21 got 0x%x", bcdbuf[0])
	}

	if bcdbuf[1] != 0x43 {
		t.Errorf("TestBcdEven: bcdbuf[1] expected 0x34 got 0x%x", bcdbuf[1])
	}

	if bcdbuf[2] != 0x65 {
		t.Errorf("TestBcdEven: bcdbuf[2] expected 0x65 got 0x%x", bcdbuf[2])
	}

	hexdata := make([]byte, hex.EncodedLen(len(bcdbuf)))
	hex.Encode(hexdata, bcdbuf)
	fmt.Printf("TestTbcdEven: 123456 encoded data [%s]\n", string(hexdata))

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("123456") {
		t.Errorf("TestTbcdEven: bcd Decode failed: got %s expect %s", bcdstr, string("123456"))
	}

}

func TestTbcdUnEven1(t *testing.T) {

	bcd := NewTbcd("0123456789??????")
	bcdbuf := bcd.Encode("12345")
	if len(bcdbuf) == 0 {
		t.Errorf("TestTbcdUnEven1: bcd Encode failed")
	}

	if bcdbuf[0] != 0x21 {
		t.Errorf("TestBcdEven: bcdbuf[0] expected 0x21 got 0x%x", bcdbuf[0])
	}

	if bcdbuf[1] != 0x43 {
		t.Errorf("TestBcdEven: bcdbuf[1] expected 0x43 got 0x%x", bcdbuf[1])
	}

	if bcdbuf[2] != 0xf5 {
		t.Errorf("TestBcdEven: bcdbuf[2] expected 0xf5 got 0x%x", bcdbuf[2])
	}

	hexdata := make([]byte, hex.EncodedLen(len(bcdbuf)))
	hex.Encode(hexdata, bcdbuf)
	fmt.Printf("TestTbcdUnEven1: 12345 encoded data [%s]\n", string(hexdata))

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("12345?") {
		t.Errorf("TestTbcdUnEven1: bcd Decode failed: got %s expect %s", bcdstr, string("12345?"))
	}
}

func TestTbcdUnEven2(t *testing.T) {

	bcd := NewTbcd("0123456789?????f")
	bcdbuf := bcd.Encode("12345f")
	if len(bcdbuf) == 0 {
		t.Errorf("TestTbcdUnEven2: bcd Encode failed")
	}

	if bcdbuf[0] != 0x21 {
		t.Errorf("TestBcdEven: bcdbuf[0] expected 0x21 got 0x%x", bcdbuf[0])
	}

	if bcdbuf[1] != 0x43 {
		t.Errorf("TestBcdEven: bcdbuf[1] expected 0x43 got 0x%x", bcdbuf[1])
	}

	if bcdbuf[2] != 0xf5 {
		t.Errorf("TestBcdEven: bcdbuf[2] expected 0xf5 got 0x%x", bcdbuf[2])
	}

	hexdata := make([]byte, hex.EncodedLen(len(bcdbuf)))
	hex.Encode(hexdata, bcdbuf)
	fmt.Printf("TestTbcdUnEven2: 12345f encoded data [%s]\n", string(hexdata))

	bcdstr := bcd.Decode(bcdbuf)
	if bcdstr != string("12345f") {
		t.Errorf("TestTbcdUnEven2: bcd Decode failed: got %s expect %s", bcdstr, string("12345f"))
	}
}
