FROM golang:latest as BUILDER
# build binary
COPY . /go/src/github.com/opensourceways/xihe-inference-evaluate
RUN cd /go/src/github.com/opensourceways/xihe-inference-evaluate && GO111MODULE=on CGO_ENABLED=0 go build

# copy binary config and utils
FROM alpine:latest
WORKDIR /opt/app/

COPY ./template ./template
COPY  --from=BUILDER /go/src/github.com/opensourceways/xihe-inference-evaluate/xihe-inference-evaluate /opt/app

ENTRYPOINT ["/opt/app/xihe-inference-evaluate"]
