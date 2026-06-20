package main

import "strings"

func fixtureDockerfile(finalUser string) string {
	return strings.Join([]string{
		"ARG GO_IMAGE=golang:1.26",
		"FROM ${GO_IMAGE} AS build",
		"WORKDIR /src",
		"COPY go.mod go.sum ./",
		"RUN go mod download",
		"COPY cmd ./cmd",
		"COPY internal ./internal",
		"ENV CGO_ENABLED=0",
		`RUN go build -trimpath -ldflags="-s -w" -o /out/riido_ai_server ./cmd/riido_ai_server`,
		"FROM scratch",
		"COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt",
		"COPY --from=build /out/riido_ai_server /riido_ai_server",
		"EXPOSE 8080",
		"ENV RIIDO_AI_SERVER_ADDR=:8080",
		"USER " + finalUser,
		`ENTRYPOINT ["/riido_ai_server"]`,
		"",
	}, "\n")
}

func fixtureTaskIR() string {
	return `{
  "schema_version": "riido-aws-fargate-task-definition.v1",
  "family": "riido-ai-server",
  "runtime_platform": {
    "cpuArchitecture": "X86_64",
    "operatingSystemFamily": "LINUX"
  },
  "container": {
    "name": "riido_ai_server",
    "portMappings": [
      {
        "containerPort": 8080,
        "hostPort": 8080,
        "protocol": "tcp"
      }
    ],
    "environment": [
      {
        "name": "RIIDO_AI_SERVER_ADDR",
        "value": ":8080"
      }
    ]
  }
}
`
}
