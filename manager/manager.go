package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/eladyarkoni/bpm/node"
	"github.com/hpcloud/tail"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	levelDBPath      = "/tmp/bulk-pm.db"
	projectPrefixKey = "project-"
	statePrefixKey   = "state-"
)

// LevelDB handler
var db *leveldb.DB

// Init initialize the manager resources
//
// Manager is using levelDB to store its data
func Init() {
	db, _ = leveldb.OpenFile(levelDBPath, nil)
}

// ClearDB Clears all database keys and values
//
// Mainly, this method will be used in the manager tests
func ClearDB() error {
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		db.Delete(iter.Key(), nil)
	}
	iter.Release()
	err := iter.Error()
	return err
}

// GetProject gets the project model
//
// Projects that added to the manager through the AddProject function, is saved inside
// the levelDB.
// This method gets the project object from the database and returns the Project model
func GetProject(packageName string) (*node.Project, error) {
	projectData, err := db.Get([]byte(projectPrefixKey+packageName), nil)
	if err != nil {
		return nil, err
	}
	var projectObject node.Project
	json.Unmarshal(projectData, &projectObject)
	return &projectObject, nil
}

// AddProject adds a node working directory as a project to the manager
//
// Node projects must have package.json file which contains the node package information
// The package name is used as the node project name
func AddProject(workingDir string) error {
	packageFilePath := fmt.Sprintf("%s/%s", workingDir, node.NodePackageFile)
	packageBytes, err := ioutil.ReadFile(packageFilePath)
	if err != nil || packageBytes == nil {
		return fmt.Errorf("%s is not found in %s", node.NodePackageFile, workingDir)
	}
	var projectObject node.Project
	projectObject.WorkingDir = workingDir
	json.Unmarshal(packageBytes, &projectObject.Package)
	if projectObject.Package.Name == "" {
		return fmt.Errorf("package name is empty")
	}
	if projectObject.Package.GetMainScript() == "" {
		return fmt.Errorf("main script definition is mandatory")
	}
	projectBytes, _ := json.Marshal(projectObject)
	db.Put([]byte(projectPrefixKey+projectObject.Package.Name), projectBytes, nil)
	return nil
}

// RemoveProject removes project from manager
func RemoveProject(packageName string) error {
	_, err := GetProject(packageName)
	if err != nil {
		return err
	}
	projectStatus, _ := GetProjectState(packageName)
	if projectStatus != nil && projectStatus.IsRunning() {
		return fmt.Errorf("project is running")
	}
	deleteError := db.Delete([]byte(projectPrefixKey+packageName), nil)
	if deleteError != nil {
		return deleteError
	}
	db.Delete([]byte(statePrefixKey+packageName), nil)
	return nil
}

// GetProjectState gets the project state
func GetProjectState(packageName string) (*ProjectState, error) {
	projectStateData, err := db.Get([]byte(statePrefixKey+packageName), nil)
	if err != nil {
		return nil, err
	}
	var projectState ProjectState
	json.Unmarshal(projectStateData, &projectState)
	// Check again with os to verify that the process is running
	if projectState.IsRunning() {
		if !IsProcessRunning(projectState.PID) {
			projectState.PID = 0
			SaveProjectState(packageName, &projectState)
		}
	}
	return &projectState, nil
}

// SaveProjectState saves the project state in the db
func SaveProjectState(packageName string, projectState *ProjectState) error {
	projectStateBytes, err := json.Marshal(projectState)
	if err != nil {
		return err
	}
	saveErr := db.Put([]byte(statePrefixKey+packageName), projectStateBytes, nil)
	if saveErr != nil {
		return saveErr
	}
	return nil
}

// StartProject starts the project processes
//
// This function is using go routine to start the project process and wait for it to finish
// If the process is stopped, the function checks if process is stopped by kill signal or not.
// If the process is not stopped from kill signal, it means that the process is crashed and should be restarted.
//
// If project should run in a cluster mode (clusterProcesses != 0) the method generates the cluster node
// script and use it as the project main script.
func StartProject(packageName string, clusterProcesses int, procStateChannel chan *ProjectState) error {
	projectData, projectDataErr := GetProject(packageName)
	if projectDataErr != nil {
		return fmt.Errorf("project is not found")
	}
	projectState, _ := GetProjectState(packageName)
	if projectState != nil && projectState.IsRunning() {
		return fmt.Errorf("project is already running")
	}
	mainScript := projectData.Package.GetMainScript()
	if mainScript == "" {
		return fmt.Errorf("project has no main script")
	}
	if clusterProcesses > 0 {
		scriptErr := CreateClusterModeScript(projectData, clusterProcesses)
		if scriptErr != nil {
			return fmt.Errorf("can't create the cluster mode script")
		}
		mainScript = ClusterModeScript
	}

	// Start scripts and monitor
	go func() {
		command := exec.Command("node", mainScript)
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Dir = projectData.WorkingDir
		logPath := fmt.Sprintf("/tmp/%s.log", packageName)
		logFile, _ := os.Create(logPath)
		command.Stdout = logFile
		command.Stderr = logFile
		runError := command.Start()
		if runError != nil {
			return
		}
		runningProjectState := &ProjectState{
			PID:       command.Process.Pid,
			LogPath:   logPath,
			StartTime: time.Now(),
		}
		SaveProjectState(packageName, runningProjectState)
		if procStateChannel != nil {
			procStateChannel <- runningProjectState
		}
		// Wait for the process to finish
		procError := command.Wait()
		// Process is finished, lets check the cause of this
		runningProjectState.EndTime = time.Now()
		runningProjectState.PID = 0
		SaveProjectState(packageName, runningProjectState)
		if procStateChannel != nil {
			procStateChannel <- runningProjectState
		}
		if procError != nil {
			errorCode := command.ProcessState.Sys().(syscall.WaitStatus)
			if errorCode != 9 && errorCode != 0 {
				// Process is terminated not by kill command, lets restart it
				log.Printf("Process %d of package %s is crashed, autorestart is activated\n", command.Process.Pid, packageName)
				autoRestartErr := StartProject(packageName, clusterProcesses, procStateChannel)
				if autoRestartErr != nil {
					log.Printf("package %s is failed to auto restart itself\n", packageName)
				}
			}
		}
	}()
	return nil
}

// StopProject stops the project processes
func StopProject(packageName string) error {
	projectState, _ := GetProjectState(packageName)
	if projectState == nil || !projectState.IsRunning() {
		return fmt.Errorf("project is not running")
	}
	// Negative PID value is used to stop the process group (process and its childs)
	syscall.Kill(-projectState.PID, syscall.SIGKILL)
	return nil
}

// GetStatus gets all project status as a dictionary of package names and project state
func GetStatus() map[string]ProjectState {
	stateMap := make(map[string]ProjectState)
	projectIter := db.NewIterator(util.BytesPrefix([]byte(projectPrefixKey)), nil)
	for projectIter.Next() {
		var projectData node.Project
		json.Unmarshal(projectIter.Value(), &projectData)
		projectState, _ := GetProjectState(projectData.Package.Name)
		if projectState == nil {
			projectState = &ProjectState{
				PID: 0,
			}
		}
		stateMap[projectData.Package.Name] = *projectState
	}
	projectIter.Release()
	return stateMap
}

// GetProjectLogContent gets last lines of the project log
func GetProjectLogContent(packageName string, numOfLines int) ([]string, error) {
	projState, projStateErr := GetProjectState(packageName)
	if projStateErr != nil {
		return nil, fmt.Errorf("Can't get project status")
	}
	if projState.LogPath == "" {
		return nil, fmt.Errorf("Logfile is not created")
	}
	t, err := tail.TailFile(projState.LogPath, tail.Config{Follow: false, MaxLineSize: numOfLines})
	if err != nil {
		return nil, err
	}
	logLines := make([]string, 0)
	for line := range t.Lines {
		logLines = append(logLines, line.Text)
	}
	return logLines, nil
}
