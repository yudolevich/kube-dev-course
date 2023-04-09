FROM golang:1.20-alpine as build

ARG BUILD_VERSION
WORKDIR /build

COPY . ./
RUN GOOS=linux GOARCH=amd64 go build \
  -ldflags="-X main.Version=${BUILD_VERSION}" \
  -a -o app main.go

FROM alpine:3.17

COPY --from=build /build/app /app

CMD ["/app"]
