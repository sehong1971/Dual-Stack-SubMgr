package control

import (
	"context"
	"encoding/hex"
	"net"

	// gRPC 표준 에러 처리를 위해 아래 두 패키지 추가
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/ricapi" // 생성된 패키지 경로
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"

	"fmt"
	"time"
)

// gRPCServer 구조체: 확인된 Unimplemented 인터페이스를 임베딩합니다.
type gRPCServer struct {
	ricapi.UnimplementedE2_Subscription_APIServer // [cite: 557, 606, 614]
	c                                             *Control
}

// RunGRPCServer: NewControl() 또는 main에서 호출될 서버 시작 함수
func (c *Control) RunGRPCServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		xapp.Logger.Error("Failed to listen on gRPC port %s: %v", port, err)
		return
	}

	// [이 부분 추가] 서버가 성공적으로 리슨을 시작했음을 알림
	xapp.Logger.Info("gRPC NBI Server for subscription is running on port %s", port)

	s := grpc.NewServer()
	// 확인된 Register 함수명을 사용합니다.
	ricapi.RegisterE2_Subscription_APIServer(s, &gRPCServer{c: c}) // [cite: 615]

	// grpcurl/Postman 검증을 위해 리플렉션 등록 [cite: 361, 375, 417]
	reflection.Register(s)

	xapp.Logger.Info("gRPC Server is running on port %s", port)
	if err := s.Serve(lis); err != nil {
		xapp.Logger.Error("Failed to serve gRPC: %v", err)
	}
}

// 설계 예시: gRPC 메시지를 내부 구조체로 변환
func (s *gRPCServer) mapGRPCtoInternalParams(req *ricapi.E2SubscriptionInitMsg) (*models.SubscriptionParams, error) {
	params := &models.SubscriptionParams{}
	subReq := req.GetE2SubscriptionRequest()
	if subReq == nil {
		return nil, fmt.Errorf("E2SubscriptionRequest is nil")
	}

	// 1-1. MEID 추출 (GlobalE2NodeId -> string)
	// 기지국 식별을 위해 gnb_MCC_MNC_ID 형식의 문자열로 변환 필요
	// 1-1-1. gRPC 메시지에서 바이너리 추출

	// type GlobalE2NodeId_GlobalE2NodeGnbId struct {
	// 	state         protoimpl.MessageState `protogen:"open.v1"`
	// 	GlobalGnbId   *GlobalGnbId           `protobuf:"bytes,1,opt,name=global_gnb_id,json=globalGnbId,proto3,oneof" json:"global_gnb_id,omitempty"`
	// 	GlobalEnGnbId *GlobalEnGnbId         `protobuf:"bytes,2,opt,name=global_en_gnb_id,json=globalEnGnbId,proto3,oneof" json:"global_en_gnb_id,omitempty"`
	// 	GnbCuUpId     *GnbCuUpId             `protobuf:"bytes,3,opt,name=gnb_cu_up_id,json=gnbCuUpId,proto3,oneof" json:"gnb_cu_up_id,omitempty"`
	// 	GnbDuId       *GnbDuId               `protobuf:"bytes,4,opt,name=gnb_du_id,json=gnbDuId,proto3,oneof" json:"gnb_du_id,omitempty"`
	// 	unknownFields protoimpl.UnknownFields
	// 	sizeCache     protoimpl.SizeCache
	// }

	nodeId := subReq.GetGlobalE2NodeId()
	// nodeId2 := subReq.GlobalE2NodeId
	// if nodeId == nodeId2 {
	// 	fmt.Print("구조체의 필드에 직접 접근(subReq.GlobalE2NodeId)하는 것이 가능")
	// 	fmt.Print("하지만 gRPC/Protobuf 환경에서는 직접 접근보다 Get...() 메서드를 사용하는 것이 권장")
	// 	fmt.Print(("1. Nil-Safety (런타임 에러 방지)"))
	// 	fmt.Print("기본값(Default Value) 처리")
	// }
	gnbNode := nodeId.GetGnb()
	if gnbNode == nil {
		return nil, fmt.Errorf("gNB node information is missing")
	}
	globalGnb := gnbNode.GetGlobalGnbId()
	// 바이너리 데이터 추출
	plmn := globalGnb.GetPlmnIdentity().GetValue()
	gnb := globalGnb.GetGnbId().GetValue()
	// DU ID는 선택 사항이므로 필드 존재 여부 확인 후 GetValue()
	// DU ID 처리 (uint64 타입에 맞게 포인터 사용)
	var duPtr *uint64
	if gnbNode.GetGnbDuId() != nil {
		val := gnbNode.GetGnbDuId().GetValue() // 여기서 uint64 반환
		duPtr = &val
	}

	ranName := GetRanName(plmn, gnb, duPtr)
	params.Meid = &ranName

	// 1-2. RAN Function ID
	ranFuncId := int64(subReq.GetRanFunctionId().GetValue())
	params.RANFunctionID = &ranFuncId

	// 1-3. Client Endpoint (xApp 주소)
	dest := subReq.GetE2IndicationDestination()
	// - dest.GetPort()는 일반 uint32 값을 반환
	// - models.SubscriptionParamsClientEndpoint 구조체의 HTTPPort 필드 타입은 int64가 아니라 ***int64 (int64 타입의 포인터)로 정의
	// - Go에서는 함수 리턴값이나 상수 리터럴에 대해 즉시 주소값(&)을 취하는 것이 허용되지 않기 때문에(예: &int64(dest.GetPort())는 문법 오류),
	//   다음과 같이 임시 변수를 사용하거나 도우미 함수를 사용해 해결
	port := int64(dest.GetPort())
	rmrPort := int64(4560)
	params.ClientEndpoint = &models.SubscriptionParamsClientEndpoint{
		Host:     dest.GetDomainName(),
		HTTPPort: &port,    // 내부 로직상 포트 필드 매핑
		RMRPort:  &rmrPort, // temporarily assigned due to pannic in restendpoint.go:38
	}

	// 1-4. Subscription Details (Trigger & Actions)
	// 1-4-1. SubscriptionDetail 객체 생성 (REST 모델의 핵심 단위)
	// gRPC의 "1개 트리거 + 여러 액션"을 REST의 "1개 Detail"로 매핑.
	grpcDetails := subReq.GetE2SubscriptionDetails()
	detail := &models.SubscriptionDetail{}

	// 1-4-2. field XappEventInstanceID *int64 `json:"XappEventInstanceId"`
	valueXappEventInstanceID := int64(1)
	detail.XappEventInstanceID = &valueXappEventInstanceID

	// 1-4-3. Event Trigger 변환 ([]byte -> []int64)
	triggerBytes := grpcDetails.GetE2EventTriggerDefinition().GetValue()
	detail.EventTriggers = s.byteToInt64Slice(triggerBytes)

	// 1-4-4. Action List 변환 및 추가
	for _, grpcAction := range grpcDetails.GetSequenceOfActions() {
		// 1. ActionID 값을 임시 변수에 저장
		actionID := int64(grpcAction.GetRequestedE2ActionId().GetValue())
		actionType := grpcAction.GetE2ActionType().String()
		action := &models.ActionToBeSetup{
			ActionID:   &actionID,
			ActionType: &actionType,
			// Action Definition 변환 ([]byte -> []int64)
			ActionDefinition: s.byteToInt64Slice(grpcAction.GetE2ActionDefinition().GetValue()),
		}

		if grpcAction.GetE2SubsequentAction() != nil {
			subsequentActionType := grpcAction.GetE2SubsequentAction().GetSubsequentActionType().String()
			timeToWait := grpcAction.GetE2SubsequentAction().GetE2TimeToWait().String()
			action.SubsequentAction = &models.SubsequentAction{
				SubsequentActionType: &subsequentActionType,
				TimeToWait:           &timeToWait,
			}
		}
		// Detail 내부의 ActionList에 추가
		detail.ActionToBeSetupList = append(detail.ActionToBeSetupList, action)
	}

	// 2. 완성된 Detail을 전체 Params에 추가
	params.SubscriptionDetails = append(params.SubscriptionDetails, detail)

	return params, nil
}

// E2SubscriptionProc는 xApp으로부터 gRPC 구독 요청(Subscription Request)을 수신하여 처리.
// 이 함수는 수신된 gRPC 메시지를 SubMgr 내부 데이터 모델(SubscriptionParams)로 매핑한 후,
// 기존 REST 인터페이스와 공유하는 내부 SubscriptionHandler를 호출하여 듀얼 스택 처리를 수행.
//
// [입력 파라미터]
// 1. ctx (context.Context): gRPC 호출의 생명주기, 타임아웃, 취소 신호를 관리.
// 2. req (*ricapi.E2SubscriptionInitMsg): 구독 요청 정보를 담고 있는 최상위 메시지.
//   - req.E2SubscriptionRequest: 실제 구독 상세 정보
//   - ProcedureTransactionId: 요청을 식별하기 위한 트랜잭션 ID
//   - GlobalE2NodeId: 대상 기지국(gNB/eNB) 식별 정보 (PLMN ID 및 gNB ID)
//   - RanFunctionId: 구독할 RAN Function의 ID (예: KPM=2)
//   - E2SubscriptionDetails: 이벤트 트리거 정의 및 액션(Action) 리스트
//   - E2IndicationDestination: xApp이 Indication(데이터)을 수신할 gRPC 엔드포인트(FQDN 및 Port)
//
// [출력 리턴]
// 1. *ricapi.E2SubscriptionOutMsg: 처리 결과를 담은 최상위 응답 메시지(Oneof 구조).
//   - E2SubscriptionSuccess: 구독 성공 시 반환. SubMgr이 할당한 RIC Request ID 및 수락된 액션 리스트 포함
//   - E2SubscriptionReject: 논리적 거부(중복 구독, 권한 부족 등) 시 반환
//   - E2SubscriptionFailure: 물리적 실패(기지국 응답 없음, 네트워크 오류 등) 시 반환
//
// 2. error: gRPC 프레임워크 수준의 통신 에러나 서버 내부의 심각한 런타임 오류를 반환.
func (s *gRPCServer) E2SubscriptionProc(ctx context.Context, req *ricapi.E2SubscriptionInitMsg) (*ricapi.E2SubscriptionOutMsg, error) {
	xapp.Logger.Info("gRPC Received E2SubscriptionProc request from xApp")
	s.c.CntRecvMsg++
	s.c.UpdateCounter(cRestSubReqFromXapp) // 기존 REST 카운터 활용 또는 신규 생성 [cite: 151, 152]

	// 1. gRPC 메시지를 내부 models.SubscriptionParams 구조체로 매핑 (통역 단계) [cite: 194, 350, 425]
	params, err := s.mapGRPCtoInternalParams(req)
	if err != nil {
		xapp.Logger.Error("gRPC Mapping failed: %v", err)
		s.c.UpdateCounter(cRestSubFailToXapp)
		return nil, err
	}

	// subReq := req.GetE2SubscriptionRequest()
	// if subReq == nil {
	// 	return nil, fmt.Errorf("E2SubscriptionRequest is nil")
	// }

	// // 1. 내부 로직 처리를 위한 Params 객체 생성 (REST 모델 재활용)
	// params := &models.SubscriptionParams{}

	// // 1-1. MEID 추출 (GlobalE2NodeId -> string)
	// // 기지국 식별을 위해 gnb_MCC_MNC_ID 형식의 문자열로 변환 필요
	// params.Meid = s.getMeidFromGrpc(subReq.GetGlobalE2NodeId())

	// // 1-2. RAN Function ID
	// ranFuncId := int64(subReq.GetRanFunctionId().GetValue())
	// params.RANFunctionID = &ranFuncId

	// // 1-3. Client Endpoint (xApp 주소)
	// dest := subReq.GetE2IndicationDestination()
	// // - dest.GetPort()는 일반 uint32 값을 반환
	// // - models.SubscriptionParamsClientEndpoint 구조체의 HTTPPort 필드 타입은 int64가 아니라 ***int64 (int64 타입의 포인터)로 정의
	// // - Go에서는 함수 리턴값이나 상수 리터럴에 대해 즉시 주소값(&)을 취하는 것이 허용되지 않기 때문에(예: &int64(dest.GetPort())는 문법 오류),
	// //   다음과 같이 임시 변수를 사용하거나 도우미 함수를 사용해 해결
	// port := int64(dest.GetPort())
	// params.ClientEndpoint = &models.SubscriptionParamsClientEndpoint{
	// 	Host:     dest.GetDomainName(),
	// 	HTTPPort: &port, // 내부 로직상 포트 필드 매핑
	// }

	// // 1-4. Subscription Details (Trigger & Actions)
	// // 1-4-1. SubscriptionDetail 객체 생성 (REST 모델의 핵심 단위)
	// // gRPC의 "1개 트리거 + 여러 액션"을 REST의 "1개 Detail"로 매핑.
	// grpcDetails := subReq.GetE2SubscriptionDetails()
	// detail := &models.SubscriptionDetail{}

	// // 1-4-2. Event Trigger 변환 ([]byte -> []int64)
	// triggerBytes := grpcDetails.GetE2EventTriggerDefinition().GetValue()
	// detail.EventTriggers = s.byteToInt64Slice(triggerBytes)

	// // 1-4-3. Action List 변환 및 추가
	// for _, grpcAction := range grpcDetails.GetSequenceOfActions() {
	// 	// 1. ActionID 값을 임시 변수에 저장
	// 	actionID := int64(grpcAction.GetRequestedE2ActionId().GetValue())
	// 	actionType := grpcAction.GetE2ActionType().String()
	//     action := &models.ActionToBeSetup{
	//         ActionID:   &actionID,
	//         ActionType: &actionType,
	//         // Action Definition 변환 ([]byte -> []int64)
	//         ActionDefinition: s.byteToInt64Slice(grpcAction.GetE2ActionDefinition().GetValue()),
	//     }

	//     if grpcAction.GetE2SubsequentAction() != nil {
	// 		subsequentActionType := grpcAction.GetE2SubsequentAction().GetSubsequentActionType().String()
	// 		timeToWait := grpcAction.GetE2SubsequentAction().GetE2TimeToWait().String()
	//         action.SubsequentAction = &models.SubsequentAction{
	//             SubsequentActionType: &subsequentActionType,
	//             TimeToWait:           &timeToWait,
	//         }
	//     }
	//     // Detail 내부의 ActionList에 추가
	//     detail.ActionToBeSetupList = append(detail.ActionToBeSetupList, action)
	// }

	// // 2. 완성된 Detail을 전체 Params에 추가
	// params.SubscriptionDetails = append(params.SubscriptionDetails, detail)
	if s.c.LoggerLevel > 2 {
		s.c.PrintRESTSubscriptionRequest(params)
	}

	// 2. E2 연결 상태 확인 (기존 REST 로직과 동일)
	if s.c.e2IfState.IsE2ConnectionUp(params.Meid) == false || s.c.e2IfState.IsE2ConnectionUnderReset(params.Meid) == true {
		xapp.Logger.Error("No E2 connection or Under Reset for Meid: %v", *params.Meid)
		s.c.UpdateCounter(cRestReqRejDueE2Down)
		return nil, fmt.Errorf("E2 Service Unavailable")
	}

	// 3. Endpoint 유효성 검사 및 주소 구성
	if params.ClientEndpoint == nil {
		xapp.Logger.Error("ClientEndpoint == nil in gRPC request")
		return nil, fmt.Errorf("ClientEndpoint is nil")
	}

	e2SubscriptionDirectives, err := s.c.GetE2SubscriptionDirectives(params)
	if err != nil {
		return nil, err
	}

	_, xAppRmrEndpoint, err := ConstructEndpointAddresses(*params.ClientEndpoint)
	if err != nil {
		return nil, err
	}

	// 4. 중복 체크를 위한 MD5 계산 (동일 요청 재전송 방지)
	md5sum, err := CalculateRequestMd5sum(params)
	if err != nil {
		xapp.Logger.Error("Failed to generate md5sum: %s", err.Error())
	}

	// 5. 트랜잭션 관리: 기존 Subscription ID 검색 또는 생성
	// gRPC 요청도 내부적으로는 REST 트랜잭션 구조를 빌려 상태를 관리함 [cite: 21, 29]
	restSubscription, grpcSubId, err := s.c.GetOrCreateRestSubscription(params, md5sum, xAppRmrEndpoint, params.ClientEndpoint.Host)
	if err != nil {
		return nil, err
	}

	// gRPC 요청인 경우에만 ID를 보관
	restSubscription.PtransID = req.E2SubscriptionRequest.GetProcedureTransactionId()
	xapp.Logger.Info("restSubscription.PtransID: %d", restSubscription.PtransID.GetValue())

	// 6. E2AP 메시지 생성 준비
	subReqList := e2ap.SubscriptionRequestList{}
	err = s.c.e2ap.FillSubscriptionReqMsgs(params, &subReqList, restSubscription)
	if err != nil {
		s.c.restDuplicateCtrl.DeleteLastKnownRestSubsIdBasedOnMd5sum(md5sum)
		s.c.registry.DeleteRESTSubscription(&grpcSubId)
		return nil, err
	}
	xapp.Logger.Info("passed FillSubscriptionReqMsgs for grpcSubId: %s", grpcSubId)

	// 7. 진행 중인 트랜잭션 중복 여부 최종 확인
	if s.c.restDuplicateCtrl.IsDuplicateToOngoingTransaction(grpcSubId, md5sum) {
		xapp.Logger.Debug("Retransmission detected for grpcSubId %s", grpcSubId)
		return &ricapi.E2SubscriptionOutMsg{
			// 성공 응답 구성
		}, nil
	}
	xapp.Logger.Info("Passed Duplication Check for grpcSubId: %s", grpcSubId)

	// 8. DB 저장 및 비동기 엔진 실행
	s.c.WriteRESTSubscriptionToDb(grpcSubId, restSubscription)
	xapp.Logger.Info("Passed WriteRESTSubscriptionToDb for grpcSubId: %s", grpcSubId)

	// <비차단 엔진과 동기식 핸들러의 조합 Design>
	// 실제 RAN과의 통신을 담당하는 processSubscriptionRequests는 여전히 비동기(고루틴)로 동작.
	// 엔진 파트: RAN으로부터 응답이 오는 순서대로 맵에서 채널을 찾아 결과를 던져줍니다. (요청 순서와 응답 순서가 달라도 상관없음)
	// 핸들러 파트: 각 xApp의 요청은 자신의 채널만 바라보며 기다리다가, 자기 몫의 응답이 오면 즉시 리턴
	// 1). 응답을 전달받을 채널 생성 (결과를 담을 그릇)
	respChan := make(chan *ricapi.E2SubscriptionOutMsg, 1)

	// 2). Control 객체에 이 요청에 대한 '대기자(Waiter)' 등록
	// restSubId를 키로 하여 응답 채널을 저장함
	s.c.RegisterGRPCWaiter(grpcSubId, respChan)
	defer s.c.UnregisterGRPCWaiter(grpcSubId) // 함수 종료 시 자동 삭제

	// 3). 기존 비동기 엔진 실행 (RAN으로 요청 전송)
	// 실제 RAN으로의 구독 절차는 별도 고루틴에서 실행
	// failuer from RTMGR
	go s.c.processSubscriptionRequests(restSubscription, &subReqList, params.ClientEndpoint, params.Meid, &grpcSubId, xAppRmrEndpoint, md5sum, e2SubscriptionDirectives)

	// 4). [핵심] 채널에서 결과가 올 때까지 대기 (Blocking)
	xapp.Logger.Info("gRPC Handler waiting for E2 response for grpcSubId: %s", grpcSubId)

	select {
	case finalResp, ok := <-respChan:
		// 성공적으로 RAN 응답(Success/Failure)을 받은 경우
		// 1. 채널이 정상적으로 데이터를 보내고 닫혔는지 확인 (ok 체크)
		if !ok {
			xapp.Logger.Error("gRPC respChan closed unexpectedly")
			return nil, status.Error(codes.Internal, "Internal communication channel closed")
		}

		// 2. 받은 데이터가 nil인지 확인 (nil 체크)
		if finalResp == nil {
			xapp.Logger.Error("gRPC received nil response from engine")
			return nil, status.Error(codes.Internal, "Engine returned empty response")
		}

		// 3. (선택사항) 응답 내용이 비어있지는 않은지 검증
		if finalResp.TypeOfMessage == nil {
			xapp.Logger.Error("gRPC response has no message type (oneof is nil)")
			return nil, status.Error(codes.Internal, "Invalid response message structure")
		}

		xapp.Logger.Info("gRPC successfully sending response for restSubId")
		return finalResp, nil

	case <-ctx.Done():
		// gRPC 클라이언트(xApp)가 연결을 끊은 경우
		xapp.Logger.Error("gRPC Client cancelled request for %s", grpcSubId)
		return nil, ctx.Err()

	case <-time.After(30 * time.Second):
		// RAN 응답이 너무 오래 걸릴 경우 (Timeout 처리)
		xapp.Logger.Error("gRPC Timeout waiting for E2 response for %s", grpcSubId)
		return nil, fmt.Errorf("E2 Response Timeout")
	}

}

// E2SubscriptionDeleteProc: 기존의 Unsubscribe 역할을 수행하는 핸들러 [cite: 348, 642]
func (s *gRPCServer) E2SubscriptionDeleteProc(ctx context.Context, req *ricapi.E2SubscriptionDeleteInitMsg) (*ricapi.E2SubscriptionDeleteOutMsg, error) {
	xapp.Logger.Info("Received gRPC E2SubscriptionDeleteProc (Unsubscribe) request")

	return &ricapi.E2SubscriptionDeleteOutMsg{}, nil
}

// E2SubscriptionModificationProc: 수정 요청 처리 핸들러
func (s *gRPCServer) E2SubscriptionModificationProc(ctx context.Context, req *ricapi.E2SubscriptionModificationInitMsg) (*ricapi.E2SubscriptionModificationOutMsg, error) {
	xapp.Logger.Info("Received gRPC E2SubscriptionModificationProc request")

	return &ricapi.E2SubscriptionModificationOutMsg{}, nil
}

// Helper Functions
// GlobalE2NodeId를 gnb_001_001_123456 형식의 문자열로 변환
func (s *gRPCServer) getMeidFromGrpc(nodeId *ricapi.GlobalE2NodeId) *string {
	if gnbNode := nodeId.GetGnb(); gnbNode != nil {
		globalGnb := gnbNode.GetGlobalGnbId()
		plmn := hex.EncodeToString(globalGnb.GetPlmnIdentity().GetValue())
		gnbId := hex.EncodeToString(globalGnb.GetGnbId().GetValue())
		meid := fmt.Sprintf("gnb_%s_%s", plmn, gnbId)
		return &meid
	}
	return nil
}

// 성공 응답 메시지 생성
func (s *gRPCServer) createSuccessResponse(req *ricapi.E2SubscriptionRequest, internalResp *models.SubscriptionResponse) *ricapi.E2SubscriptionOutMsg {
	successMsg := &ricapi.E2SubscriptionSuccess{}

	// Transaction ID 복사
	successMsg.ProcedureTransactionId = req.GetProcedureTransactionId()

	// 할당된 Subscription ID 설정 (internalResp에서 가져옴)
	// gRPC 필드 구조에 맞춰 E2RequestId 등을 채워 넣습니다.
	// ... (상세 필드 설정 생략) ...

	return &ricapi.E2SubscriptionOutMsg{
		TypeOfMessage: &ricapi.E2SubscriptionOutMsg_E2SubscriptionSuccess{
			E2SubscriptionSuccess: successMsg,
		},
	}
}

// []byte를 models에서 사용하는 []int64 타입으로 변환하는 함수
// Go 언어는 []byte를 []int64로 직접 캐스팅(Type Conversion)하는 것을 허용하지 않기 때문에, 바이트 하나하나를 정수로 변환하여 복사하는 과정이 필요.
func (s *gRPCServer) byteToInt64Slice(b []byte) []int64 {
	res := make([]int64, len(b))
	for i, v := range b {
		res[i] = int64(v)
	}
	return res
}

// DecodePLMN은 3바이트 PLMN 바이너리를 MCC, MNC 문자열로 변환합니다.
// VIAVI 환경에 맞춰 MNC가 2자리인 경우 앞에 '0'을 붙여 3자리로 만듭니다.
func DecodePLMN(plmn []byte) (string, string) {
	if len(plmn) < 3 {
		return "000", "000"
	}

	// Octet 1: d2 d1
	d1 := plmn[0] & 0x0F
	d2 := (plmn[0] & 0xF0) >> 4

	// Octet 2: m3 d3 (m3는 2자리 MNC일 경우 0xF)
	d3 := plmn[1] & 0x0F
	m3 := (plmn[1] & 0xF0) >> 4

	// Octet 3: m2 m1
	m1 := plmn[2] & 0x0F
	m2 := (plmn[2] & 0xF0) >> 4

	mcc := fmt.Sprintf("%d%d%d", d1, d2, d3)

	var mnc string
	if m3 == 0xF {
		// 2자리 MNC (예: 01) -> VIAVI 관습에 따라 001로 변환
		mnc = fmt.Sprintf("0%d%d", m1, m2)
	} else {
		mnc = fmt.Sprintf("%d%d%d", m1, m2, m3)
	}

	return mcc, mnc
}

// GetRanName은 gRPC로 받은 바이너리 데이터들을 조합하여 RanName을 생성합니다.
func GetRanName(plmnID []byte, gnbID []byte, duID *uint64) string {
	// 1. PLMN 디코딩 (00f110 -> "001", "001")
	mcc, mnc := DecodePLMN(plmnID)

	// 2. gNB ID 변환 (바이너리 -> hex 문자열)
	gnbStr := hex.EncodeToString(gnbID)

	// duID가 nil이 아니면 gnbd 형식으로 조립
	if duID != nil {
		// duID가 uint64이므로 %d를 사용하여 직접 출력
		return fmt.Sprintf("gnbd_%s_%s_%s_%d", mcc, mnc, gnbStr, *duID)
	}

	// 결과 예시: gnb_208_099_00000e00
	return fmt.Sprintf("gnb_%s_%s_%s", mcc, mnc, gnbStr)
}
