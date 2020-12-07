// +build plan9

package model

import (
	sysPlan9 "golang.org/x/sys/plan9"
	"os"
)

func GetXid(info os.FileInfo) (string, string) {
	UID := "N/A"
	GID := "N/A"
	if stat, ok := info.Sys().(*sysPlan9.Dir); ok {
		UID = stat.Uid
		GID = stat.Gid
	}
	return UID, GID
}
