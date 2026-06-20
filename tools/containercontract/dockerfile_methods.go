package main

func (d dockerfile) stageByAlias(alias string) *stage {
	for i := range d.Stages {
		if d.Stages[i].Alias == alias {
			return &d.Stages[i]
		}
	}
	return nil
}

func (d dockerfile) finalStage() *stage {
	if len(d.Stages) == 0 {
		return nil
	}
	return &d.Stages[len(d.Stages)-1]
}
