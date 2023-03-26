FROM alpine:3.17 as build

WORKDIR /build

RUN apk add --no-cache clang musl-dev binutils gcc
COPY main.c ./
RUN clang main.c

FROM alpine:3.17

COPY --from=build /build/a.out /app

CMD ["/app", "/tmp/data", "30"]
