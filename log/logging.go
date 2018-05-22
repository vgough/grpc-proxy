package log

import (
	"io/ioutil"
	"log"
	"os"
)

var (
	// Debug logging for grpc-proxy package.
	Debug = log.New(ioutil.Discard, "DEBUG ", log.LstdFlags)
	// Info logging for grpc-proxy package.
	Info = log.New(os.Stderr, "INFO ", log.LstdFlags)
	// Error logging for grpc-proxy package.
	Error = log.New(os.Stderr, "ERROR ", log.LstdFlags)
)
