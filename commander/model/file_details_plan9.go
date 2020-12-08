// +build plan9

package model

import (
	//sysPlan9 "golang.org/x/sys/plan9"
	"syscall"
	"os"
)

func GetXid(info os.FileInfo) (string, string) {
	UID := "N/A"
	GID := "N/A"
	if stat, ok := info.Sys().(*syscall.Dir); ok {
		UID = stat.Uid
		GID = stat.Gid
	}
	return UID, GID
}
