package duplicates

func Build(targets []Target, policy Policy) Run {
	run := Run{
		Enabled: policy.Enabled, Status: policy.Status,
		GroupBy: policy.GroupBy, PackagePrefixes: policy.PackagePrefixes,
	}
	if !policy.Enabled {
		run.Status = "disabled"
		return run
	}
	index := map[string]*Group{}
	for _, target := range targets {
		if !prefixMatch(target.PackagePath, policy.PackagePrefixes) {
			continue
		}
		for _, file := range target.Files {
			group := groupFor(index, file.ShapeHash)
			group.Packages = append(group.Packages, target.PackagePath)
			group.Files = append(group.Files, file.Path)
		}
	}
	groups := sortedGroups(index)
	run.GroupCount, run.FileCount = len(groups), fileCount(groups)
	run.InternalGroupCount = internalGroupCount(groups)
	run.Groups = limitGroups(groups, policy.MaxGroups)
	return run
}

func groupFor(index map[string]*Group, hash string) *Group {
	if index[hash] == nil {
		index[hash] = &Group{ShapeHash: hash}
	}
	return index[hash]
}
