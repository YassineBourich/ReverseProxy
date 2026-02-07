package logging

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"reverse_proxy/CustomErrors"
	"time"
)

type Logger interface {
	Init() error
	Log(method string, url_path string, remoteAddr string, host string, status_code int, duration time.Duration)
}

// no logging startegy
type NoLogger struct{}
func (l *NoLogger) Init() error {
	return nil
}
func (l *NoLogger) Log(method string, url_path string, remoteAddr string, host string, status_code int, duration time.Duration) {}

// logging to file and command line shell strategy
type FileLogger struct {}

func (fl *FileLogger) Init() error {
	// os.O_APPEND: add to end of file
    // os.O_CREATE: create if not exists
    // os.O_WRONLY: open for writing only
	file, err := os.OpenFile(filepath.Join("log", "reverse_proxy.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return customerrors.LoggerInitError
	}

	// creating multi output
	multi_output := io.MultiWriter(os.Stdout, file)
	// Set global log output to the multi output
	log.SetOutput(multi_output)
	return nil
}

func (l *FileLogger) Log(method string, url_path string, remoteAddr string, host string, status_code int, duration time.Duration) {
	log.Printf("%s %s -> RemoteAddr: %s | Backend: %s | Status: %d | Duration: %v", 
    method, url_path, remoteAddr, host, status_code, duration);
}