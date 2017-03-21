package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// ResponseObject server success json response
type ResponseObject struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

// SendJSON send struct as a json response
func SendJSON(res http.ResponseWriter, code int, structClass interface{}) {
	res.Header().Set("Content-Type", "application/json")
	response, marshalErr := json.Marshal(structClass)
	if marshalErr != nil {
		fmt.Printf("Marshal err: %v\n", marshalErr)
	}
	res.WriteHeader(code)
	res.Write(response)
}

// SendText send text response
func SendText(res http.ResponseWriter, code int, bytes []byte) {
	res.Header().Set("Content-Type", "text/html")
	res.WriteHeader(code)
	res.Write(bytes)
}

// ReadBodyJSON read body as a json
func ReadBodyJSON(req *http.Request, structClass interface{}) {
	body, _ := ioutil.ReadAll(req.Body)
	json.Unmarshal(body, &structClass)
}

// SendSuccess send a success message
func SendSuccess(res http.ResponseWriter, message string, data []byte) {
	if data == nil {
		data = []byte("null")
	}
	SendJSON(res, 200, &ResponseObject{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError send an error message
func SendError(res http.ResponseWriter, message string) {
	SendJSON(res, 500, &ResponseObject{
		Success: false,
		Message: message,
		Data:    []byte("null"),
	})
}
