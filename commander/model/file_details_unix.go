// +build linux

package model

import (
	//sysUnix "golang.org/x/sys/unix"
	"os"
	"strconv"
	"syscall"
)

func GetXid(info os.FileInfo) (string, string) {
	UID := "N/A"
	GID := "N/A"
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = strconv.Itoa(int(stat.Uid))
		GID = strconv.Itoa(int(stat.Gid))
	}
	return UID, GID
}
