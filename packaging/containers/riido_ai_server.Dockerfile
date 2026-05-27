ARG GO_IMAGE=golang:1.26
FROM ${GO_IMAGE} AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal

ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/riido_ai_server ./cmd/riido_ai_server

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/riido_ai_server /riido_ai_server
EXPOSE 8080
ENV RIIDO_AI_SERVER_ADDR=:8080
USER 65532:65532
ENTRYPOINT ["/riido_ai_server"]
