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
	"io"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type PlmnIdentityIf interface {
	String() string
	MccString() string
	MncString() string
	EncodeTo(writer io.Writer) (int, error)
	DecodeFrom(reader io.Reader) (int, error)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type PlmnIdentity struct {
	Mcc string
	Mnc string
}

func (plmnid *PlmnIdentity) String() string {
	return plmnid.MccString() + plmnid.MncString()
}

func (plmnid *PlmnIdentity) MccString() string {
	return plmnid.Mcc
}

func (plmnid *PlmnIdentity) MncString() string {
	return plmnid.Mnc
}

func (plmnid *PlmnIdentity) Set(str string) {
	plmnid.Mcc = str[0:3]
	plmnid.Mnc = str[3:]
}

//-----------------------------------------------------------------------------
//
// MCC 3 digits MNC 2 digits
// String format      : C1C2C3N1N2
// Pre encode format  : C1C2C3FN1N2
// TBCD Coded format  : 0xC2C1 0xfC3 0xN2N1
// Post decode format : C1C2C3FN1N2
// String format      : C1C2C3N1N2
//
// MCC 3 digits MNC 3 digits
// String format      : C1C2C3N1N2N3
// Pre encode format  : C1C2C3N3N1N2
// TBCD Coded format  : 0xC2C1 0xN3C3 0xN2N1
// Post decode format : C1C2C3N3N1N2
// String format      : C1C2C3N1N2N3
//
//-----------------------------------------------------------------------------

type PlmnIdentityTbcd struct {
	PlmnIdentity
}

func (plmnid *PlmnIdentityTbcd) EncodeTo(writer io.Writer) (int, error) {

	var tmpStr string
	switch {
	case len(plmnid.Mnc) == 2:
		tmpStr = plmnid.Mcc + string("f") + plmnid.Mnc
	case len(plmnid.Mnc) == 3:
		tmpStr = plmnid.Mcc + string(plmnid.Mnc[2]) + string(plmnid.Mnc[0:2])
	default:
		return 0, nil
	}

	buf := TBCD.Encode(tmpStr)
	return writer.Write(buf)
}

func (plmnid *PlmnIdentityTbcd) DecodeFrom(reader io.Reader) (int, error) {
	tmpBytes := make([]byte, 3)
	n, err := reader.Read(tmpBytes)
	if err != nil {
		return n, err
	}
	str := TBCD.Decode(tmpBytes)

	if str[3] == 'f' {
		plmnid.Mcc = string(str[0:3])
		plmnid.Mnc = string(str[4:])
	} else {
		plmnid.Mcc = string(str[0:3])
		plmnid.Mnc = string(str[4:]) + string(str[3])
	}
	return n, nil
}

//-----------------------------------------------------------------------------
//
// MCC 3 digits MNC 2 digits
// String format      : C1C2C3N1N2
// Pre encode format  : C1C2C3FN1N2
// BCD Coded format   : 0xC2C1 0xC3f 0xN1N2
// Post decode format : C1C2C3FN1N2
// String format      : C1C2C3N1N2
//
// MCC 3 digits MNC 3 digits
// String format      : C1C2C3N1N2N3
// Pre encode format  : C1C2C3N1N2N3
// BCD Coded format   : 0xC2C1 0xC3N1 0xN2N3
// Post decode format : C1C2C3N1N2N3
// String format      : C1C2C3N1N2N3
//
//-----------------------------------------------------------------------------

type PlmnIdentityBcd struct {
	PlmnIdentity
}

func (plmnid *PlmnIdentityBcd) EncodeTo(writer io.Writer) (int, error) {

	var tmpStr string
	switch {
	case len(plmnid.Mnc) == 2:
		tmpStr = plmnid.Mcc + string("f") + plmnid.Mnc
	case len(plmnid.Mnc) == 3:
		tmpStr = plmnid.Mcc + plmnid.Mnc
	default:
		return 0, nil
	}

	buf := BCD.Encode(tmpStr)
	return writer.Write(buf)
}

func (plmnid *PlmnIdentityBcd) DecodeFrom(reader io.Reader) (int, error) {
	tmpBytes := make([]byte, 3)
	n, err := reader.Read(tmpBytes)
	if err != nil {
		return n, err
	}
	str := BCD.Decode(tmpBytes)

	if str[3] == 'f' {
		plmnid.Mcc = string(str[0:3])
		plmnid.Mnc = string(str[4:])
	} else {
		plmnid.Mcc = string(str[0:3])
		plmnid.Mnc = string(str[3:])
	}
	return n, nil
}
