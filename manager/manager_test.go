package manager

import (
	"os"
	"path/filepath"
	"testing"
)

var testProjectDirectory = filepath.Join(os.Getenv("GOPATH"), "/src/github.com/eladyarkoni/bpm/testdata/ExpressProject/")
var testProjectPackageName = "express-example-project"

func init() {
	Init()
}

func TestAddingAProject(t *testing.T) {
	// Initializing the manager
	ClearDB()
	addProjectErr := AddProject(testProjectDirectory)
	if addProjectErr != nil {
		t.Fatalf("adding a project to manager error: %s", addProjectErr)
		t.FailNow()
	}
}

func TestGetAddedProject(t *testing.T) {
	ClearDB()
	AddProject(testProjectDirectory)
	projectModel, getProjectErr := GetProject(testProjectPackageName)
	if getProjectErr != nil {
		t.Fatal(getProjectErr)
		t.FailNow()
	}
	t.Logf("Project data: %v", projectModel)
}

func TestGetProjectStatusNotRunning(t *testing.T) {
	ClearDB()
	addProjectErr := AddProject(testProjectDirectory)
	if addProjectErr != nil {
		t.Fatalf("adding a project to manager error: %s", addProjectErr)
		t.FailNow()
	}
	projectState, err := GetProjectState(testProjectPackageName)
	if err == nil || (projectState != nil && projectState.IsRunning()) {
		t.Fatal("Project that just created should not run")
		t.FailNow()
	}
}

func TestRunningACreatedProject(t *testing.T) {
	ClearDB()
	addProjectErr := AddProject(testProjectDirectory)
	if addProjectErr != nil {
		t.Fatalf("adding a project to manager error: %s", addProjectErr)
		t.FailNow()
	}
	receivingProjStateChan := make(chan *ProjectState)
	defer close(receivingProjStateChan)

	startProjectErr := StartProject(testProjectPackageName, 0, receivingProjStateChan)
	if startProjectErr != nil {
		t.Fatalf("project is failed to start: %s", startProjectErr)
		t.FailNow()
	}

	projectState := <-receivingProjStateChan

	if projectState == nil {
		t.Fatal("project state is not saved")
		t.FailNow()
	} else if !projectState.IsRunning() {
		t.Fatal("project is not running for at least 5 seconds")
		t.FailNow()
	}

	// Now, try to stop the project process
	stopProjectErr := StopProject(testProjectPackageName)
	if stopProjectErr != nil {
		t.Fatalf("could not stopping the project due to error: %s", stopProjectErr)
		t.FailNow()
	}

	projectStoppedState := <-receivingProjStateChan
	if projectStoppedState.IsRunning() {
		t.Fatal("Project is not stopped")
		t.FailNow()
	}
}
