package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"encoding/json"

	"strconv"

	"github.com/eladyarkoni/bpm/manager"
	"github.com/gorilla/mux"
)

// Start Starts the server listener
func Start(port string) error {
	manager.Init()
	serverRouter := mux.NewRouter()
	serverRouter.HandleFunc("/status", GetServerStatus).Methods("GET")
	serverRouter.HandleFunc("/manager/status", GetManagerStatus).Methods("GET")
	serverRouter.HandleFunc("/manager/project", AddProject).Methods("POST")
	serverRouter.HandleFunc("/manager/project/{package}", GetProject).Methods("GET")
	serverRouter.HandleFunc("/manager/project/{package}/log", GetProjectLog).Methods("GET")
	serverRouter.HandleFunc("/manager/project/{package}", RemoveProject).Methods("DELETE")
	serverRouter.HandleFunc("/manager/project/{package}/status", GetProjectStatus).Methods("GET")
	serverRouter.HandleFunc("/manager/project/{package}/start", StartProject).Methods("POST")
	serverRouter.HandleFunc("/manager/project/{package}/stop", StopProject).Methods("POST")
	http.Handle("/", serverRouter)
	srv := &http.Server{
		Handler:      serverRouter,
		Addr:         fmt.Sprintf("127.0.0.1:%s", port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Printf("Bulk Server is started on port %s", port)
	err := srv.ListenAndServe()
	return err
}

// GetServerStatus gets server status
func GetServerStatus(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	SendSuccess(res, "Server is available", nil)
}

// GetManagerStatus gets the manager status
func GetManagerStatus(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	statusMap := manager.GetStatus()
	statusMapData, _ := json.Marshal(statusMap)
	SendSuccess(res, "Status is available", statusMapData)
}

// AddProject adds a new project to manager
func AddProject(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	var requestBody struct {
		WorkingDir string `json:"working_dir"`
	}
	ReadBodyJSON(req, &requestBody)
	if requestBody.WorkingDir == "" {
		SendError(res, "working_dir is empty")
		return
	}
	if err := manager.AddProject(requestBody.WorkingDir); err != nil {
		SendError(res, fmt.Sprintf("%s", err))
		return
	}
	SendSuccess(res, "Project is added successfully", nil)
}

// GetProject gets the project data
func GetProject(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	params := mux.Vars(req)
	packageName := params["package"]
	projectModel, err := manager.GetProject(packageName)
	if err != nil {
		SendError(res, fmt.Sprintf("%s", err))
		return
	}
	projectModelData, _ := json.Marshal(projectModel)
	SendSuccess(res, "Project data is available", projectModelData)
}

// GetProjectLog gets the project log lines
// Query params: lines = <last lines limit>
func GetProjectLog(res http.ResponseWriter, req *http.Request) {
	linesLimit := 10
	defer req.Body.Close()
	params := mux.Vars(req)
	packageName := params["package"]
	queryParams := req.URL.Query()
	if queryParams.Get("lines") != "" {
		queryLinesLimit, err := strconv.Atoi(queryParams.Get("lines"))
		if err == nil {
			linesLimit = queryLinesLimit
		}
	}
	logLines, logErr := manager.GetProjectLogContent(packageName, linesLimit)
	if logErr != nil {
		SendError(res, fmt.Sprintf("%s", logErr))
		return
	}
	logLinesJSON, _ := json.Marshal(logLines)
	SendSuccess(res, "log is ready", logLinesJSON)
}

// RemoveProject removes the project
func RemoveProject(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	params := mux.Vars(req)
	packageName := params["package"]
	err := manager.RemoveProject(packageName)
	if err != nil {
		SendError(res, fmt.Sprintf("%s", err))
		return
	}
	SendSuccess(res, "Project is removed successfully", nil)
}

// GetProjectStatus gets the project status
func GetProjectStatus(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	params := mux.Vars(req)
	packageName := params["package"]
	projectStateModel, err := manager.GetProjectState(packageName)
	if err != nil {
		SendError(res, fmt.Sprintf("%s", err))
		return
	}
	projectStateModelData, _ := json.Marshal(projectStateModel)
	SendSuccess(res, "Project status is available", projectStateModelData)
}

// StartProject starts the project processes
func StartProject(res http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	clusterProcesses := 0
	params := mux.Vars(req)
	packageName := params["package"]
	queryParams := req.URL.Query()
	if queryParams.Get("clusterProcesses") != "" {
		clusterProcesses, _ = strconv.Atoi(queryParams.Get("clusterProcesses"))
	}
	startProjectErr := manager.StartProject(packageName, clusterProcesses, nil)
	if startProjectErr != nil {
		log.Printf("StartProject %s error: %s\n", packageName, startProjectErr)
		SendError(res, fmt.Sprintf("%v", startProjectErr))
		return
	}
	SendSuccess(res, "Project is started successfully", nil)
}

// StopProject stops the project processes
func StopProject(res http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	packageName := params["package"]
	stopProjectErr := manager.StopProject(packageName)
	if stopProjectErr != nil {
		SendError(res, fmt.Sprintf("%s", stopProjectErr))
		return
	}
	SendSuccess(res, "Project is stopped successfully", nil)
}
