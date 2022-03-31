#
# Build
#
FROM golang:1.17-alpine AS build-env
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app
COPY . /app

RUN go mod download \
    && CGO_ENABLED=0 GOOS=linux go build -o /demo-app

#
# Deploy
#
FROM gcr.io/distroless/static
COPY --from=build-env /demo-app /
EXPOSE 8000
ENTRYPOINT ["./demo-app"]