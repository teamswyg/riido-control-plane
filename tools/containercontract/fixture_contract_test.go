package main

func fixtureBuildContract() buildContract {
	return buildContract{
		BuildArg:   buildArgContract{Name: "GO_IMAGE", Default: "golang:1.26"},
		StageName:  "build",
		Workdir:    "/src",
		CGOEnabled: "0",
		GoBuild: goBuildContract{
			Package:  "./cmd/riido_ai_server",
			Output:   "/out/riido_ai_server",
			Trimpath: true,
			LDFlags:  []string{"-s", "-w"},
		},
	}
}

func fixtureFinalContract(finalUser string) finalContract {
	return finalContract{
		BaseImage:  "scratch",
		CopyFrom:   "build",
		CopySource: "/out/riido_ai_server",
		Binary:     "/riido_ai_server",
		RequiredCopies: []requiredCopyContract{
			{
				From:        "build",
				Source:      "/etc/ssl/certs/ca-certificates.crt",
				Destination: "/etc/ssl/certs/ca-certificates.crt",
			},
		},
		ExposedPorts: []int{8080},
		Env: map[string]string{
			"RIIDO_AI_SERVER_ADDR": ":8080",
		},
		User:       finalUser,
		Entrypoint: []string{"/riido_ai_server"},
	}
}
