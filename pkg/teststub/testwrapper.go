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
package teststub

import (
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strings"
	"testing"
)

type TestWrapper struct {
	*xapp.Log
	desc string
}

func (tw *TestWrapper) Init(desc string) {
	tw.desc = strings.ToUpper(desc)
	tw.Log = xapp.NewLogger(tw.desc)
	//tw.SetLevel(defaultLogLevel)
}

func (tw *TestWrapper) GetDesc() string {
	return tw.desc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tw *TestWrapper) TestError(t *testing.T, pattern string, args ...interface{}) {
	tw.Error(fmt.Sprintf(pattern, args...))
	t.Errorf(fmt.Sprintf(pattern, args...))
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tw *TestWrapper) TestLog(t *testing.T, pattern string, args ...interface{}) {
	tw.Debug(fmt.Sprintf(pattern, args...))
	t.Logf(fmt.Sprintf(pattern, args...))
}

func NewTestWrapper(desc string) *TestWrapper {
	tent := &TestWrapper{}
	tent.Init(desc)
	return tent
}
