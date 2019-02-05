FROM golang:1.11

WORKDIR /hello-web
COPY main.go .
RUN go build --ldflags '-linkmode "external" -extldflags "-static"' -o hello main.go

FROM scratch
COPY --from=0 /hello-web/hello .
ENV LISTEN_ADDR :8080
EXPOSE 8080/tcp
CMD ["./hello"]

