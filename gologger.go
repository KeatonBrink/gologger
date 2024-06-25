package gologger

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type GoLogger struct {
	isPrinting         bool
	messageLog         []string
	messageLogLen      int
	outputFileName     string
	inputChannel       chan string
	inputChannelSize   int
	openFileConnection *os.File
	endChannel         chan struct{}
	curMessage         string
	pauseChannel       chan struct{}
}

var goLogger GoLogger

func InitGoLogger(isPrinting bool, outputFileName string) error {
	inputChannelSize := 100
	goLogger = GoLogger{
		isPrinting:         isPrinting,
		messageLog:         make([]string, 0),
		messageLogLen:      0,
		outputFileName:     outputFileName,
		inputChannel:       make(chan string, inputChannelSize),
		inputChannelSize:   inputChannelSize,
		openFileConnection: nil,
		endChannel:         make(chan struct{}, 1),
		curMessage:         "",
		pauseChannel:       make(chan struct{}),
	}

	if len(goLogger.outputFileName) > 0 {
		f, err := os.OpenFile(goLogger.outputFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}

		goLogger.openFileConnection = f
	}

	if goLogger.isPrinting {
		go printHandler()
	}

	QueueMessage("Starting GoLogger")

	go messageHandler()

	return nil
}

func messageHandler() {
	var messageSize int
	for {
		goLogger.curMessage = <-goLogger.inputChannel
		goLogger.curMessage = formatMessage(goLogger.curMessage)
		// fmt.Println("Length of inputchannel: ", len(goLogger.inputChannel))
		messageSize = 1
		for ; len(goLogger.inputChannel) > 0 && messageSize < goLogger.inputChannelSize; messageSize++ {
			goLogger.curMessage += formatMessage(<-goLogger.inputChannel)
		}
		goLogger.logMessage(goLogger.curMessage)
		if len(goLogger.inputChannel) == 0 && len(goLogger.endChannel) > 0 {
			// fmt.Println("Ending GoLogger")
			<-goLogger.endChannel
			break
		}
		goLogger.curMessage = ""
	}
}

func printHandler() {
	// disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
	// do not display entered characters on the screen
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == 'p' {
			fmt.Println("Pausing")
			goLogger.isPrinting = false
			enableInput()
			goLogger.pauseChannel <- struct{}{}
			// Break would also work here, but this is more clear
			return
		}
	}
}

func enableInput() {
	err := exec.Command("stty", "-F", "/dev/tty", "sane").Run()
	if err != nil {
		log.Fatalf("Failed to reset terminal settings: %v", err)
	}
}

func formatMessage(text string) string {
	now := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%v - %v\n", now, text)
}

func (gLog *GoLogger) logMessage(logText string) error {

	if gLog.isPrinting {
		fmt.Print(logText)
	}

	if gLog.openFileConnection != nil {
		if _, err := gLog.openFileConnection.WriteString(logText); err != nil {
			panic(err)
		}
	}

	return nil
}

func EnablePrintHandler() {
	if !goLogger.isPrinting {
		goLogger.isPrinting = true
		go printHandler()
	}
}

func EmptyMessageQueue() error {
	if len(goLogger.inputChannel) > 0 || len(goLogger.curMessage) > 0 {
		// fmt.Println("Length of inputChannel: ", len(goLogger.inputChannel))
		// fmt.Println("Length of curMessage: ", len(goLogger.curMessage))
		goLogger.endChannel <- struct{}{}
		goLogger.endChannel <- struct{}{}
	}
	enableInput()
	return nil
}

func WaitForPause() {
	<-goLogger.pauseChannel
}

func QueueMessage(text string) error {
	goLogger.inputChannel <- text
	return nil
}
