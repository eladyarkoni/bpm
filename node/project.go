package node

// Project node project structure
type Project struct {
	WorkingDir string  `json:"working_dir"`
	Package    Package `json:"package"`
}
