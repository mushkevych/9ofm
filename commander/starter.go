package commander

import (
	"fmt"
	"github.com/jroimartin/gocui"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

var (
	once         sync.Once
	appSingleton *Application
)

// Run is the UI entrypoint.
func Run() error {
	var err error

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer gui.Close()

	once.Do(func() {
		var application *Application
		application, err = NewApplication(gui)
		if err != nil {
			return
		}

		// perform the first update and render now that all resources have been loaded
		err = application.UpdateAndRender()
		if err != nil {
			return
		}
	})

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Error("main loop error: ", err)
		return err
	}
	return nil
}

func start(events eventChannel) {
	var err error
	defer close(events)

	err = Run()
	if err != nil {
		events.exitWithError(err)
		return
	}
}

func Start() {
	var exitCode int
	var events = make(eventChannel)

	go start(events)

	for event := range events {
		if event.stdout != "" {
			fmt.Println(event.stdout)
		}

		if event.stderr != "" {
			_, err := fmt.Fprintln(os.Stderr, event.stderr)
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if event.err != nil {
			log.Error(event.err)
			_, err := fmt.Fprintln(os.Stderr, event.err.Error())
			if err != nil {
				fmt.Println("error: could not write to buffer:", err)
			}
		}

		if event.errorOnExit {
			exitCode = 1
		}
	}
	os.Exit(exitCode)
}
