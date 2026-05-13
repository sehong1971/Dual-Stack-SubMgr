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

//
//

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type bcdbase struct {
	convtbl string
}

func (bcd *bcdbase) index(c byte) int {
	for cpos, cchar := range bcd.convtbl {
		if cchar == rune(c) {
			return cpos
		}
	}
	return -1
}

func (bcd *bcdbase) byte(i int) byte {
	if i < 0 || i > 15 {
		return '?'
	}
	return bcd.convtbl[i]
}

//-----------------------------------------------------------------------------
//
// 5 digit example
// String format   : D1D2D3D4D5
// BCD Coded format: 0xD1D2 0xD3D4 0xD50f
// String format   : D1D2D3D4D5F
//
// 6 digit example
// String format   : D1D2D3D4D5D6
// BCD Coded format: 0xD1D2 0xD3D4 0xD5D6
// String format   : D1D2D3D4D5D6
//
//-----------------------------------------------------------------------------
type Bcd struct {
	bcdbase
}

func (bcd *Bcd) Encode(str string) []byte {
	buf := make([]byte, len(str)/2+len(str)%2)
	for i := 0; i < len(str); i++ {
		var schar int = bcd.index(str[i])
		if schar < 0 {
			return nil
		}
		if i%2 > 0 {
			buf[i/2] &= 0xf0
			buf[i/2] |= ((uint8)(schar) & 0x0f)
		} else {
			buf[i/2] = 0x0f | (uint8)(schar)<<4
		}

	}
	return buf
}

func (bcd *Bcd) Decode(buf []byte) string {
	var strbytes []byte
	for i := 0; i < len(buf); i++ {
		var b byte

		b = bcd.byte(int(buf[i] >> 4))
		//if b == '?' {
		//	return ""
		//}
		strbytes = append(strbytes, b)

		b = bcd.byte(int(buf[i] & 0x0f))
		//if b == '?' {
		//	return ""
		//}
		strbytes = append(strbytes, b)
	}
	return string(strbytes)
}

//-----------------------------------------------------------------------------
//
// 5 digit example
// String format   : D1D2D3D4D5
// TBCD Coded format: 0xD2D1 0xD4D3 0x0fD5
// String format   : D1D2D3D4D5F
//
// 6 digit example
// String format   : D1D2D3D4D5
// TBCD Coded format: 0xD2D1 0xD4D3 0xD6D5
// String format   : D1D2D3D4D5D6
//
//-----------------------------------------------------------------------------
type Tbcd struct {
	bcdbase
}

func (bcd *Tbcd) Encode(str string) []byte {
	buf := make([]byte, len(str)/2+len(str)%2)
	for i := 0; i < len(str); i++ {
		var schar int = bcd.index(str[i])
		if schar < 0 {
			return nil
		}
		if i%2 > 0 {
			buf[i/2] &= 0x0f
			buf[i/2] |= (uint8)(schar) << 4
		} else {
			buf[i/2] = 0xf0 | ((uint8)(schar) & 0x0f)
		}
	}
	return buf
}

func (bcd *Tbcd) Decode(buf []byte) string {
	var strbytes []byte
	for i := 0; i < len(buf); i++ {
		var b byte
		b = bcd.byte(int(buf[i] & 0x0f))
		//if b == '?' {
		//	return ""
		//}
		strbytes = append(strbytes, b)

		b = bcd.byte(int(buf[i] >> 4))
		//if b == '?' {
		//	return ""
		//}
		strbytes = append(strbytes, b)
	}
	return string(strbytes)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func NewBcd(convTbl string) *Bcd {
	b := &Bcd{}
	if len(convTbl) == 16 {
		b.convtbl = convTbl
	} else {
		b.convtbl = "0123456789?????f"
	}
	return b
}

func NewTbcd(convTbl string) *Tbcd {
	b := &Tbcd{}
	if len(convTbl) == 16 {
		b.convtbl = convTbl
	} else {
		b.convtbl = "0123456789*#abcf"
	}
	return b
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var BCD *Bcd
var TBCD *Tbcd

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func init() {
	BCD = NewBcd("")
	TBCD = NewTbcd("")
}
