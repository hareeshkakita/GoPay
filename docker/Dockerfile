FROM golang:1.26-alpine

WORKDIR /app

RUN go install github.com/go-delve/delve/cmd/dlv@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -gcflags="all=-N -l" -o /app/server ./cmd/api

EXPOSE 8060 40000

CMD ["/go/bin/dlv", "exec", "/app/server", "--headless", "--listen=:40000", "--api-version=2", "--accept-multiclient","--continue"]