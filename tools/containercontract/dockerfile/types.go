package dockerfile

type File struct {
	Args   map[string]string
	Stages []Stage
}

type Stage struct {
	Base       string
	Alias      string
	Workdir    string
	Env        map[string]string
	Runs       []string
	Copies     []CopyInstruction
	Exposes    []int
	User       string
	Entrypoint []string
}

type CopyInstruction struct {
	From string
	Src  string
	Dst  string
}

func (d File) StageByAlias(alias string) *Stage {
	for i := range d.Stages {
		if d.Stages[i].Alias == alias {
			return &d.Stages[i]
		}
	}
	return nil
}

func (d File) FinalStage() *Stage {
	if len(d.Stages) == 0 {
		return nil
	}
	return &d.Stages[len(d.Stages)-1]
}

func IntSetEqual(a, b []int) bool {
	a = SortedInts(a)
	b = SortedInts(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func SortedInts(in []int) []int {
	out := append([]int(nil), in...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j] < out[i] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}
