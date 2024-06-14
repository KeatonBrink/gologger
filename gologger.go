package gologger

import (
	"fmt"
	"os"
	"time"
)

type GoLogger struct {
	isPrinting         bool
	messageLog         []string
	messageLogLen      int
	outputFileName     string
	inputChannel       chan string
	openFileConnection *os.File
}

func InitGoLogger(isPrinting bool, outputFileName string) (GoLogger, error) {
	goLogger := GoLogger{
		isPrinting:         isPrinting,
		messageLog:         make([]string, 0),
		messageLogLen:      0,
		outputFileName:     outputFileName,
		inputChannel:       make(chan string),
		openFileConnection: nil,
	}
	// Todo: Open the  output file to write to
	f, err := os.OpenFile(goLogger.outputFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	goLogger.openFileConnection = f

	goLogger.LogMessage("Starting GoLogger")

	return goLogger, nil
}

func (gLog *GoLogger) LogMessage(text string) error {
	now := time.Now().Format("2006-01-02 15:04:05")

	logText := fmt.Sprintf("%v - %v\n", now, text)

	if gLog.isPrinting {
		fmt.Print(logText)
	}

	if _, err := gLog.openFileConnection.WriteString(logText); err != nil {
		return err
	}

	return nil
}
