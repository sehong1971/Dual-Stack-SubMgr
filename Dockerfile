###########################################################
# 1. Base 스테이지: Go 1.23 및 기본 도구 설정
###########################################################
FROM nexus3.o-ran-sc.org:10002/o-ran-sc/bldr-ubuntu20-c-go:1.0.0 as submgrcore

# Go 버전을 gRPC 호환성을 위해 1.23.2로 설정 (사용자 요청 반영)
ARG GOVERSION="1.23.2"
RUN wget -nv https://dl.google.com/go/go${GOVERSION}.linux-amd64.tar.gz \
     && tar -xf go${GOVERSION}.linux-amd64.tar.gz \
     && mv go /opt/go/${GOVERSION} \
     && rm -f go*.gz

ENV DEFAULTPATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
ENV PATH=$DEFAULTPATH:/usr/local/go/bin:/opt/go/${GOVERSION}/bin:/root/go/bin

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump gdb valgrind

WORKDIR /tmp

# Swagger, RMR 설치 로직 (기존과 동일)
ARG SWAGGERVERSION=v0.23.0
ARG SWAGGERURL=https://github.com/go-swagger/go-swagger/releases/download/${SWAGGERVERSION}/swagger_linux_amd64
RUN wget --quiet ${SWAGGERURL} \
    && mv swagger_linux_amd64 swagger \
    && chmod +x swagger \
    && mv swagger /usr/local/bin/

RUN export GOBIN=/usr/local/bin/ ; \
  go install github.com/go-delve/delve/cmd/dlv@latest

ARG RMRVERSION=4.9.4
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/rmr_${RMRVERSION}_amd64.deb/download.deb && dpkg -i rmr_${RMRVERSION}_amd64.deb
RUN wget --content-disposition https://packagecloud.io/o-ran-sc/release/packages/debian/stretch/rmr-dev_${RMRVERSION}_amd64.deb/download.deb && dpkg -i rmr-dev_${RMRVERSION}_amd64.deb
RUN rm -f rmr*.deb

WORKDIR /opt/submgr

###########################################################
# 2. E2AP Build 스테이지: C-Library 빌드 (기존 동일)
###########################################################
FROM submgrcore as submgre2apbuild
ENV CFLAGS="-DASN_DISABLE_OER_SUPPORT"
ENV CGO_CFLAGS="-DASN_DISABLE_OER_SUPPORT"

COPY 3rdparty 3rdparty
RUN cd 3rdparty/E2AP-v02.00.00 && \
    gcc -c ${CFLAGS} -I. -g -fPIC *.c && \
    gcc *.o -g -shared -o libe2ap.so && \
    cp libe2ap.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

COPY e2ap e2ap
RUN cd e2ap/libe2ap_wrapper && \
    gcc -c ${CFLAGS} -g -fPIC *.c && \
    gcc *.o -g -shared -o libe2ap_wrapper.so && \
    cp libe2ap_wrapper.so /usr/local/lib/ && \
    cp *.h /usr/local/include/ && \
    ldconfig

###########################################################
# 3. SubMgr Build 스테이지: gRPC 소스 포함 및 빌드
###########################################################
FROM submgre2apbuild as submgrbuild

COPY go.mod go.mod
RUN go mod download

RUN mkdir pkg
COPY api api

# RTMgr Client 생성 (기존 동일)
ARG RTMGRVERSION=8becf0c4e06bc89b13d217a102eb7a50470cddc5
RUN git clone "https://gerrit.o-ran-sc.org/r/ric-plt/rtmgr" \
    && git -C "rtmgr" checkout $RTMGRVERSION \
    && cp rtmgr/api/routing_manager.yaml api/ \
    && rm -rf rtmgr
RUN swagger generate client -f api/routing_manager.yaml -t pkg/ -m rtmgr_models -c rtmgr_client

# [중요] 호스트에서 미리 컴파일한 pb.go 파일들을 포함한 pkg 폴더 복사
# 이 단계에서 pkg/ricapi/*.pb.go 파일들이 컨테이너 내부로 들어옵니다.
COPY pkg pkg
COPY cmd cmd
COPY go.sum go.sum

# gRPC 라이브러리 의존성 및 버전 정합성 정리
RUN go mod tidy

# SubMgr 바이너리 빌드
RUN mkdir -p /opt/bin && \
    go build -o /opt/bin/submgr cmd/submgr.go

###########################################################
# 4. Final 스테이지: 실행 이미지 생성
###########################################################
FROM ubuntu:20.04

RUN apt update && apt install -y iputils-ping net-tools curl tcpdump

COPY --from=submgrbuild /opt/bin/submgr /
COPY --from=submgrbuild /usr/local/lib/librmr* /usr/local/lib/
COPY --from=submgrbuild /usr/local/lib/libe2ap* /usr/local/lib/
RUN ldconfig

# 환경 변수 및 설정 파일 복사 (기존 동일)
COPY config /opt/config
ENV CFG_FILE=/opt/config/submgr-config.yaml
ENV RMR_SEED_RT=/opt/config/submgr-uta-rtg.rt

# [추가] gRPC NBI용 50051 포트 개방
EXPOSE 8080 4560 50051

ENTRYPOINT ["/submgr"]