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
	inputChannelSize   int
	openFileConnection *os.File
	endChannel         chan struct{}
	curMessage         string
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
	}
	// Todo: Open the  output file to write to
	f, err := os.OpenFile(goLogger.outputFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	goLogger.openFileConnection = f

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

func formatMessage(text string) string {
	now := time.Now().Format("2006-01-02 15:04:05")

	return fmt.Sprintf("%v - %v\n", now, text)
}

func QueueMessage(text string) error {
	goLogger.inputChannel <- text
	return nil
}

func (gLog *GoLogger) logMessage(logText string) error {

	if gLog.isPrinting {
		fmt.Print(logText)
	}

	if _, err := gLog.openFileConnection.WriteString(logText); err != nil {
		panic(err)
	}

	return nil
}

func EmptyMessageQueue() error {
	if len(goLogger.inputChannel) > 0 || len(goLogger.curMessage) > 0 {
		// fmt.Println("Length of inputChannel: ", len(goLogger.inputChannel))
		// fmt.Println("Length of curMessage: ", len(goLogger.curMessage))
		goLogger.endChannel <- struct{}{}
		goLogger.endChannel <- struct{}{}
	}
	return nil
}
