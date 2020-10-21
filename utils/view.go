package utils

import (
	"github.com/jroimartin/gocui"
	log "github.com/sirupsen/logrus"
)

// IsNewView determines if a controller has already been created based on the set of errors given (a bit hokie)
func IsNewView(errs ...error) bool {
	for _, err := range errs {
		if err == nil {
			return false
		}
		if err != gocui.ErrUnknownView {
			log.Errorf("IsNewView() unexpected error: %+v", err)
			return true
		}
	}
	return true
}
