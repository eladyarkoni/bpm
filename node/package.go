package node

// NodePackageFile the node package file name
const NodePackageFile = "package.json"

// Package npm package file structure
type Package struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Version      string            `json:"version"`
	Main         string            `json:"main"`
	Dependencies map[string]string `json:"dependencies"`
	Scripts      map[string]string `json:"scripts"`
}

// GetStartScript gets the package start script
func (pkg *Package) GetStartScript() string {
	if val, ok := pkg.Scripts["start"]; ok {
		return val
	}
	return ""
}

// GetMainScript gets the package main script
func (pkg *Package) GetMainScript() string {
	return pkg.Main
}
