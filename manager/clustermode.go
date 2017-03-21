package manager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eladyarkoni/bpm/node"
)

// ClusterModeScript the cluster mode script name
const ClusterModeScript = "cluster_mode_script.js"

// getClusterModeScript get cluster mode node script
//
// Get node script that runs node sub processes in cluster mode
// scriptPath: the package script path
// processCount: how many subprocesses will be run under the cluster master process
func getClusterModeScript(scriptPath string, processCount int) []byte {
	const clusterModeScriptTemplate = `
var cluster = require('cluster');
if (cluster.isMaster) {
	for (var i = 0; i < %d; i++) {
		cluster.fork();
	}
} else {
	// Require main script path
	require('%s');
}
`
	return []byte(fmt.Sprintf(clusterModeScriptTemplate, processCount, scriptPath))
}

// CreateClusterModeScript creates the cluster mode script inside the project working dir
//
// Cluster mode script can fork node script to multiple processes that run as a child
// processes inside the node cluster
//
// This function created the script inside the project working directory and requires the
// original script of the project inside.
func CreateClusterModeScript(project *node.Project, processCount int) error {
	clusterModeScriptName := filepath.Join(project.WorkingDir, ClusterModeScript)
	scriptFile, fileErr := os.Create(clusterModeScriptName)
	if fileErr != nil {
		return fileErr
	}
	defer scriptFile.Close()
	_, writeErr := scriptFile.Write(getClusterModeScript(fmt.Sprintf("./%s", project.Package.GetMainScript()), processCount))
	if writeErr != nil {
		return writeErr
	}
	return nil
}
