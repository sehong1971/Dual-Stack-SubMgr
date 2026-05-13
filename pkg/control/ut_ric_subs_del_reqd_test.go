package control

import (
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func initMock() *Control {

	var payload = []byte{0, 8, 0, 50, 0, 0, 3, 0, 29, 0, 5, 0, 0, 3, 0, 159, 0, 5, 0, 2, 0, 2, 0, 30, 0, 28, 0, 7, 8, 0, 0, 5, 0, 0, 0, 0, 0, 19, 64, 14, 64, 122, 32, 10, 0, 1, 4, 64, 0, 0, 0, 0, 0, 23}
	var payload1 = []byte{0, 8, 0, 50, 0, 0, 3, 0, 29, 0, 5, 0, 0, 3, 0, 160, 0, 5, 0, 2, 0, 2, 0, 30, 0, 28, 0, 7, 8, 0, 0, 5, 0, 0, 0, 0, 0, 19, 64, 14, 64, 122, 32, 10, 0, 2, 4, 64, 0, 0, 0, 0, 0, 23}
	meid := &xapp.RMRMeid{}
	meid.RanName = "RAN_NAME_20"
	var params = &xapp.RMRParams{
		Mtype:      12023,
		Payload:    payload,
		PayloadLen: len(payload),
		Meid:       meid,
		Xid:        "457945551669",
		SubId:      -1,
		Src:        "service-ricplt-e2term-rmr-alpha.ricplt:38000",
		Mbuf:       nil,
		Whid:       0,
		Callid:     0,
		Timeout:    0,
	}
	C1 := &Control{
		RMRClient:         mainCtrl.c.RMRClient,
		e2ap:              mainCtrl.c.e2ap,
		registry:          mainCtrl.c.registry,
		tracker:           mainCtrl.c.tracker,
		restDuplicateCtrl: mainCtrl.c.restDuplicateCtrl,
		e2IfState:         mainCtrl.c.e2IfState,
		e2IfStateDb:       mainCtrl.c.e2IfStateDb,
		e2SubsDb:          mainCtrl.c.e2SubsDb,
		restSubsDb:        mainCtrl.c.restSubsDb,
		CntRecvMsg:        mainCtrl.c.CntRecvMsg,
		ResetTestFlag:     mainCtrl.c.ResetTestFlag,
		Counters:          mainCtrl.c.Counters,
		LoggerLevel:       mainCtrl.c.LoggerLevel,
		UTTesting:         mainCtrl.c.UTTesting,
	}

	subReqMsg, _ := C1.e2ap.UnpackSubscriptionRequest(params.Payload)
	subReqMsg1, _ := C1.e2ap.UnpackSubscriptionRequest(payload1)

	trans := C1.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subReqMsg.RequestId, params.Meid)
	trans1 := C1.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subReqMsg1.RequestId, params.Meid)

	for _, acts := range subReqMsg.ActionSetups {
		acts.ActionType = e2ap.E2AP_ActionTypeInsert
	}
	for _, acts := range subReqMsg1.ActionSetups {
		acts.ActionType = e2ap.E2AP_ActionTypeInsert
	}
	_, _, _ = C1.registry.AssignToSubscription(trans, subReqMsg, C1.ResetTestFlag, C1, true)
	_, _, _ = C1.registry.AssignToSubscription(trans1, subReqMsg1, C1.ResetTestFlag, C1, true)

	controlObj := testingSubmgrControl{
		RmrControl: teststub.RmrControl{},
		c:          C1,
	}
	handler := controlObj.c
	return handler
}

func TestControl_handleE2TSubscriptionDeleteRequired(t *testing.T) {
	handler := initMock()
	var payload = []byte{0, 12, 64, 20, 0, 0, 1, 0, 50, 64, 13, 1, 0, 51, 64, 8, 0, 0, 123, 0, 1, 0, 2, 86}
	meid := &xapp.RMRMeid{}
	meid.RanName = "RAN_NAME_20"
	var params = &xapp.RMRParams{
		Mtype:      12023,
		Payload:    payload,
		PayloadLen: len(payload),
		Meid:       meid,
		Xid:        "457945551669",
		SubId:      -1,
		Src:        "service-ricplt-e2term-rmr-alpha.ricplt:38000",
		Mbuf:       nil,
		Whid:       0,
		Callid:     0,
		Timeout:    0,
	}
	type fields struct {
		RMRClient         *xapp.RMRClient
		e2ap              *E2ap
		registry          *Registry
		tracker           *Tracker
		restDuplicateCtrl *DuplicateCtrl
		e2IfState         *E2IfState
		e2IfStateDb       XappRnibInterface
		e2SubsDb          Sdlnterface
		restSubsDb        Sdlnterface
		CntRecvMsg        uint64
		ResetTestFlag     bool
		Counters          map[string]xapp.Counter
		LoggerLevel       int
		UTTesting         bool
	}
	type args struct {
		params *xapp.RMRParams
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "abc",
			fields: fields{
				RMRClient:         handler.RMRClient,
				e2ap:              handler.e2ap,
				registry:          handler.registry,
				tracker:           handler.tracker,
				restDuplicateCtrl: handler.restDuplicateCtrl,
				e2IfState:         handler.e2IfState,
				e2IfStateDb:       handler.e2IfStateDb,
				e2SubsDb:          handler.e2SubsDb,
				restSubsDb:        handler.restSubsDb,
				CntRecvMsg:        handler.CntRecvMsg,
				ResetTestFlag:     handler.ResetTestFlag,
				Counters:          handler.Counters,
				LoggerLevel:       handler.LoggerLevel,
				UTTesting:         handler.UTTesting,
			},

			args: args{params: params},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Control{
				RMRClient:         tt.fields.RMRClient,
				e2ap:              tt.fields.e2ap,
				registry:          tt.fields.registry,
				tracker:           tt.fields.tracker,
				restDuplicateCtrl: tt.fields.restDuplicateCtrl,
				e2IfState:         tt.fields.e2IfState,
				e2IfStateDb:       tt.fields.e2IfStateDb,
				e2SubsDb:          tt.fields.e2SubsDb,
				restSubsDb:        tt.fields.restSubsDb,
				CntRecvMsg:        tt.fields.CntRecvMsg,
				ResetTestFlag:     tt.fields.ResetTestFlag,
				Counters:          tt.fields.Counters,
				LoggerLevel:       tt.fields.LoggerLevel,
				UTTesting:         tt.fields.UTTesting,
			}
			c.e2IfState.NbIdMap[params.Meid.RanName] = "_CONNECTED"
			c.handleE2TSubscriptionDeleteRequired(tt.args.params)
			subs, _ := c.registry.GetSubscriptionFirstMatch([]uint32{uint32(1)})
			subs1, _ := c.registry.GetSubscriptionFirstMatch([]uint32{uint32(2)})
			assert.Nil(t, subs)
			assert.NotNil(t, subs1)
			time.Sleep(1 * time.Second)
		})
	}
}

func TestE2ap_UnpackSubscriptionDeleteRequired(t *testing.T) {
	var payload = []byte{0, 12, 64, 20, 0, 0, 1, 0, 50, 64, 13, 1, 0, 51, 64, 8, 0, 0, 123, 0, 1, 0, 2, 86}
	var e2ap1 = E2ap{}
	list, _ := e2ap1.UnpackSubscriptionDeleteRequired(payload)
	expectedList := &e2ap.SubscriptionDeleteRequiredList{E2APSubscriptionDeleteRequiredRequests: []e2ap.E2APSubscriptionDeleteRequired{{
		RequestId: e2ap.RequestId{
			Id:         123,
			InstanceId: 1,
		},
		FunctionId: 2,
		Cause: e2ap.Cause{
			Content: 6,
			Value:   3,
		},
	}}}
	assert.Equal(t, expectedList, list)
}
func TestE2ap_UnpackSubscriptionDeleteRequiredForWrongPayload(t *testing.T) {
	var payload = []byte{12, 64, 20, 0, 0, 1, 0, 50, 64, 13, 1, 0, 51, 64, 8, 0, 0, 123, 0, 1, 0, 2, 86}
	var e2ap1 = E2ap{}
	_, err := e2ap1.UnpackSubscriptionDeleteRequired(payload)
	assert.NotNil(t, err)

}

func TestE2ap_PackSubscriptionDeleteRequired(t *testing.T) {
	list := &e2ap.SubscriptionDeleteRequiredList{E2APSubscriptionDeleteRequiredRequests: []e2ap.E2APSubscriptionDeleteRequired{{
		RequestId: e2ap.RequestId{
			Id:         123,
			InstanceId: 1,
		},
		FunctionId: 2,
		Cause: e2ap.Cause{
			Content: 6,
			Value:   3,
		},
	}}}
	e2ap1 := E2ap{}
	payload1 := []byte{0, 12, 64, 20, 0, 0, 1, 0, 50, 64, 13, 1, 0, 51, 64, 8, 0, 0, 123, 0, 1, 0, 2, 86}
	payload := &e2ap.PackedData{Buf: payload1}
	_, packedata, _ := e2ap1.PackSubscriptionDeleteRequired(list)
	assert.Equal(t, payload, packedata)
}
