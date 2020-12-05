package commander

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
)

type EventChannel chan Event

type Event struct {
	Stdout      string
	Stderr      string
	Err         error
	ErrorOnExit bool
}

func (ec EventChannel) Message(msg string) {
	ec <- Event{
		Stdout: msg,
	}
}

func (ec EventChannel) ExitWithError(err error) {
	ec <- Event{
		Err:         err,
		ErrorOnExit: true,
	}
}

func (ec EventChannel) ExitWithErrorMessage(msg string, err error) {
	ec <- Event{
		Stderr:      msg,
		Err:         err,
		ErrorOnExit: true,
	}
}

func MainEventLoop(events EventChannel) (exitCode int){
	for event := range events {
		if event.Stdout != "" {
			fmt.Println(event.Stdout)
		}

		if event.Stderr != "" {
			_, err := fmt.Fprintln(os.Stderr, event.Stderr)
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if event.Err != nil {
			log.Error(event.Err)
			_, err := fmt.Fprintln(os.Stderr, event.Err.Error())
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if event.ErrorOnExit {
			exitCode = 1
		}
	}

	return exitCode
}