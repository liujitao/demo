FROM golang:1.17-alpine
ENV GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPROXY=https://goproxy.io,direct

RUN mkdir -p /app
WORKDIR /app
COPY . .
RUN go mod download && go build -o main .
EXPOSE 8000
ENTRYPOINT ["./main"]