package model

import (
	"testing"
)

func checkError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf(message+": %+v", err)
	}
}
