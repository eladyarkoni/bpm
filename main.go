package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"

	"strconv"

	"encoding/json"

	"time"

	"github.com/eladyarkoni/bpm/manager"
	"github.com/eladyarkoni/bpm/node"
	"github.com/eladyarkoni/bpm/server"
)

const defaultServerPort = 9663

const usageString = `
----------------------------------------------
Bulk Process Manager
----------------------------------------------
Usage:
	bulk-pm [command] arg1,arg2,arg3...

Commands:
	server                                     starts the main process manager server
	add    <working_dir>                       Adds a new project to process manager
	status                                     Gets the status of all projects
	start  <project_name> [num_of_processes]   Starts project processes (if num_of_processes is defined or not 0, the project will run in cluster mode)
	stop   <project_name>                      Stops all project processes
	info   <project_name>                      Gets the information of the added project package name
	log    <project_name>                      Gets 50 last lines of the package log
`

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println(usageString)
		return
	}
	command := args[0]
	switch command {
	case "server":
		CommandServer(args)
	case "add":
		CommandAdd(args)
	case "status":
		CommandStatus(args)
	case "start":
		CommandStart(args)
	case "stop":
		CommandStop(args)
	case "info":
		CommandInfo(args)
	case "log":
		CommandLog(args, 50)
	default:
		color.Cyan(usageString)
	}
}

// printErrorAndExit prints error in red and exit program
func printErrorAndExit(format string, args ...interface{}) {
	color.Red(format, args...)
	os.Exit(1)
}

// printSuccess prints success in green to the console
func printSuccess(format string, args ...interface{}) {
	color.Green(format, args...)
}

// CommandServer starts the server process
func CommandServer(args []string) {
	err := server.Start(strconv.Itoa(defaultServerPort))
	if err != nil {
		printErrorAndExit("Can't start the server, error: %s\n", err)
	}
}

// CommandAdd adds a new project
func CommandAdd(args []string) {
	if len(args) < 2 {
		printErrorAndExit("project working dir is missing")
	}
	respObj, err := ServerRequest("POST", "manager/project", map[string]string{
		"working_dir": args[1],
	}, true)
	if err != nil {
		printErrorAndExit("Error: %s\n", err)
	}
	printSuccess(respObj.Message)
}

// CommandStatus gets the status of all added projects
func CommandStatus(args []string) {
	statusRes, err := ServerRequest("GET", "manager/status", nil, true)
	if err != nil {
		printErrorAndExit("Error: %s\n", err)
	}
	if !statusRes.Success {
		printErrorAndExit("Error: %s\n", statusRes.Message)
	}
	var projectStatus map[string]manager.ProjectState
	json.Unmarshal(statusRes.Data, &projectStatus)
	longestProjectNameLength := 0
	for projectName := range projectStatus {
		if len(projectName) > longestProjectNameLength {
			longestProjectNameLength = len(projectName)
		}
	}
	longestProjectNameLength += 5

	color.Cyan("%s\t%s\t%s\t%s\n",
		strToColumn("Project name", longestProjectNameLength),
		strToColumn("Process PID", 12),
		strToColumn("State", 12),
		strToColumn("Duration", 8),
	)
	for projectName, projectState := range projectStatus {
		runState := color.RedString(strToColumn("Not Running", 12))
		var durationMinutes int
		if projectState.IsRunning() {
			runState = color.GreenString(strToColumn("Running", 12))
			durationMinutes = int(time.Now().Sub(projectState.StartTime).Minutes())
		}
		fmt.Printf("%s\t%s\t%s\t%s\n",
			strToColumn(projectName, longestProjectNameLength),
			strToColumn(fmt.Sprintf("%d", projectState.PID), 12),
			runState,
			strToColumn(fmt.Sprintf("%dm", durationMinutes), 8),
		)
	}
}

// CommandInfo Gets the project information
func CommandInfo(args []string) {
	if len(args) < 2 {
		printErrorAndExit("project name is missing")
	}
	projectName := args[1]
	res, err := ServerRequest("GET", fmt.Sprintf("manager/project/%s", projectName), nil, true)
	if err != nil {
		printErrorAndExit("Error: %s\n", err)
	} else if !res.Success {
		printErrorAndExit("Server Error: %s\n", res.Message)
	}
	var projectModel node.Project
	json.Unmarshal(res.Data, &projectModel)

	fmt.Printf("Package Name:        %s\n", color.CyanString(projectModel.Package.Name))
	fmt.Printf("Package Description: %s\n", color.CyanString(projectModel.Package.Description))
	fmt.Printf("Package Version:     %s\n", color.CyanString(projectModel.Package.Version))
	fmt.Printf("Working Dir:         %s\n", color.CyanString(projectModel.WorkingDir))
}

// CommandLog Gets the project last X lines of the log file
func CommandLog(args []string, lastLinesLimit int) {
	if len(args) < 2 {
		printErrorAndExit("project name is missing")
	}
	projectName := args[1]
	res, err := ServerRequest("GET", fmt.Sprintf("manager/project/%s/log?lines=%d", projectName, lastLinesLimit), nil, true)
	if err != nil {
		printErrorAndExit("Error: %s\n", err)
	} else if !res.Success {
		printErrorAndExit("Server Error: %s\n", res.Message)
	}
	var logData []string
	json.Unmarshal(res.Data, &logData)
	for _, line := range logData {
		fmt.Println(line)
	}
}

// CommandStart Starts the project processes
func CommandStart(args []string) {
	if len(args) < 2 {
		printErrorAndExit("project name is missing")
	}
	clusterModeProcesses := 0
	var argErr error
	if len(args) == 3 {
		clusterModeProcesses, argErr = strconv.Atoi(args[2])
		if argErr != nil {
			clusterModeProcesses = 0
		}
	}
	projectName := args[1]
	res, err := ServerRequest("POST", fmt.Sprintf("manager/project/%s/start?clusterProcesses=%d", projectName, clusterModeProcesses), nil, true)
	if err != nil {
		printErrorAndExit("Error: %s\n", err)
	} else if !res.Success {
		printErrorAndExit("Error: %s\n", res.Message)
	}
	printSuccess("%s\n", res.Message)
}

// CommandStop stops the project processes
func CommandStop(args []string) {
	if len(args) < 2 {
		printErrorAndExit("project name is missing")
	}
	projectName := args[1]
	res, err := ServerRequest("POST", fmt.Sprintf("manager/project/%s/stop", projectName), nil, true)
	if err != nil {
		printErrorAndExit("Error: %s\n", err)
	} else if !res.Success {
		printErrorAndExit("Error: %s\n", res.Message)
	}
	printSuccess("%s\n", res.Message)
}

// ServerRequest sends request to the server and gets response.
//
// if server is not runnin, tries to restart the server
func ServerRequest(method string, uri string, body interface{}, restartIfFailed bool) (*server.ResponseObject, error) {
	bodyBytes, _ := json.Marshal(body)
	req, requestErr := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:%d/%s", defaultServerPort, uri), bytes.NewBuffer(bodyBytes))
	if requestErr != nil {
		fmt.Printf("Server request error: %s\n", requestErr)
		return nil, requestErr
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, clientErr := client.Do(req)
	if clientErr != nil {
		if !restartIfFailed {
			return nil, clientErr
		}
		startServerErr := StartServerProcess()
		if startServerErr != nil {
			return nil, startServerErr
		}
		return ServerRequest(method, uri, body, restartIfFailed)
	}
	defer resp.Body.Close()
	respBody, _ := ioutil.ReadAll(resp.Body)
	var serverResponse server.ResponseObject
	json.Unmarshal(respBody, &serverResponse)
	return &serverResponse, nil
}

// StartServerProcess starts the server in a different process
func StartServerProcess() error {
	color.Blue("Starting Bulk Daemon...\n")
	currentExecDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	ctx := context.Background()
	command := exec.CommandContext(ctx, os.Args[0], "server")
	command.Dir = currentExecDir
	runError := command.Start()
	if runError != nil {
		return runError
	}
	// Wait for server to be up
	for true {
		serverResp, _ := ServerRequest("GET", "status", nil, false)
		if serverResp != nil {
			color.Blue("Bulk Daemon is started with pid: %d, port: %d\n\n", command.Process.Pid, defaultServerPort)
			return nil
		}
		time.Sleep(time.Millisecond * 100)
	}
	color.Blue("Bulk Daemon is started with pid: %d, port: %d\n\n", command.Process.Pid, defaultServerPort)
	return nil
}

// strToColumn Gets the string in length size
func strToColumn(str string, size int) string {
	if len(str) > size {
		return str[:size] + "..."
	}
	newStr := str
	for i := len(str); i < size; i++ {
		newStr += " "
	}
	return newStr
}
