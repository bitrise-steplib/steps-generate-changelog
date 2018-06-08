package git

import (
	"time"
)

// Commit ...
type Commit struct {
	Hash    string
	Message string
	Date    time.Time
	Author  string
	Tag     string
}
