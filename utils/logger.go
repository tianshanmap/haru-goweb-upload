package utils
import (
    "github.com/alecthomas/log4go"
)

var Log log4go.Logger

// SetLogger assigns the configured logger instance globally
func SetLogger(l log4go.Logger) {
	Log = l
}