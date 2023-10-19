FROM golang:latest AS builder
LABEL stage=intermediate
WORKDIR /
COPY . .
ENV GO111MODULE=on
RUN mkdir -p static && \
    cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" static && \
    cp index.html static
RUN GOOS=js GOARCH=wasm go build -o main.wasm
RUN CGO_ENABLED=0 GOOS=linux go build -o websrv webserver/webserver.go

FROM alpine:latest
LABEL maintainer = "hilli@github.com"
LABEL org.opencontainers.image.source https://github.com/hilli/wasm-test

WORKDIR /app

COPY --from=builder /static static
COPY --from=builder /websrv websrv

EXPOSE 3000
ENTRYPOINT [ "/app/websrv" ]