FROM golang:1.19-alpine as build-stage
WORKDIR /build
COPY ./ ./
RUN --mount=type=cache,target=/go/pkg CGO_ENABLED=0 go build -o GoforPomodoroCheck cmd/GoforPomodoroCheck/main.go
RUN --mount=type=cache,target=/go/pkg CGO_ENABLED=0 go build -o GoforPomodoroBot cmd/GoforPomodoroBot/main.go

FROM ubuntu:latest
ARG INTERNAL_SERVER_PORT
WORKDIR /app
COPY --from=build-stage /build/GoforPomodoroCheck ./
COPY --from=build-stage /build/GoforPomodoroBot ./
RUN apt update && apt install -y ca-certificates
ENTRYPOINT [ "bash", "-c", "if ./GoforPomodoroCheck ; then ./GoforPomodoroBot ; fi" ]
EXPOSE $INTERNAL_SERVER_PORT
