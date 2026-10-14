package duplicates

type Policy struct {
	Enabled         bool     `json:"enabled"`
	GroupBy         string   `json:"group_by"`
	PackagePrefixes []string `json:"package_prefixes"`
	MaxGroups       int      `json:"max_groups"`
	Status          string   `json:"status"`
}

type Run struct {
	Enabled            bool     `json:"enabled"`
	Status             string   `json:"status"`
	GroupBy            string   `json:"group_by"`
	PackagePrefixes    []string `json:"package_prefixes"`
	GroupCount         int      `json:"group_count"`
	FileCount          int      `json:"file_count"`
	InternalGroupCount int      `json:"internal_group_count"`
	Groups             []Group  `json:"groups"`
}

type Group struct {
	ShapeHash string   `json:"shape_hash"`
	FileCount int      `json:"file_count"`
	Packages  []string `json:"packages"`
	Files     []string `json:"files"`
}

type Target struct {
	ID          string
	PackagePath string
	Files       []File
}

type File struct {
	Path      string
	ShapeHash string
}
