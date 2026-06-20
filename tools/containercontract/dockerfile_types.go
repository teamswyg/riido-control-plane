package main

type dockerfile struct {
	Args   map[string]string
	Stages []stage
}

type stage struct {
	Base       string
	Alias      string
	Workdir    string
	Env        map[string]string
	Runs       []string
	Copies     []copyInstruction
	Exposes    []int
	User       string
	Entrypoint []string
}

type copyInstruction struct {
	From string
	Src  string
	Dst  string
}

type taskDefinitionIR struct {
	RuntimePlatform struct {
		OperatingSystemFamily string `json:"operatingSystemFamily"`
	} `json:"runtime_platform"`
	Container struct {
		PortMappings []struct {
			ContainerPort int `json:"containerPort"`
		} `json:"portMappings"`
		Environment []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"environment"`
	} `json:"container"`
}
