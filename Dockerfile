FROM golang:1.12.7
WORKDIR /go/src/github.com/moshloop/platform-cli
COPY . .
ENV GO111MODULE=on
RUN go mod download
RUN go build main.go -o platform-cli
