ARG GO_VERSION=1.17.2
ARG ALPINE_VERSION=3.14

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build

ARG BUF_VERSION=v1.0.0-rc6
ARG ACTION

RUN apk add --no-cache make
RUN go install github.com/bufbuild/buf/cmd/buf@${BUF_VERSION}
RUN go install github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking@${BUF_VERSION}
RUN go install github.com/bufbuild/buf/cmd/protoc-gen-buf-lint@${BUF_VERSION}

WORKDIR /opt/hello-web
COPY . .
RUN make $ACTION

FROM alpine:${ALPINE_VERSION}
WORKDIR /opt/hello-web
COPY --from=build /opt/hello-web/bin/hello-web .
CMD ["./hello-web"]
