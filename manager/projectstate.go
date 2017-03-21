package manager

import "time"

// ProjectState the project running state
type ProjectState struct {
	PID       int       `json:"pid"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	LogPath   string    `json:"log_path"`
}

// IsRunning returns true if PID is not zero
func (projectstate *ProjectState) IsRunning() bool {
	if projectstate.PID != 0 {
		return true
	}
	return false
}
