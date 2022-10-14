FROM golang:1.19-alpine as build-stage
WORKDIR /build
COPY ./ ./
RUN CGO_ENABLED=0 go build -o GoforPomodoroCheck cmd/GoforPomodoroCheck/main.go
RUN CGO_ENABLED=0 go build -o GoforPomodoroBot cmd/GoforPomodoroBot/main.go

FROM ubuntu:latest
WORKDIR /app
COPY --from=build-stage /build/GoforPomodoroCheck ./
COPY --from=build-stage /build/GoforPomodoroBot ./
ENTRYPOINT [ "bash", "-c", "if ./GoforPomodoroCheck ; then ./GoforPomodoroBot ; fi" ]

