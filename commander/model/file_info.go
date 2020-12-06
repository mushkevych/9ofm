// +build plan9 linux

package model

import (
	"fmt"
	"io"
	"os"
	"syscall"
	//sysPlan9 "golang.org/x/sys/plan9"
	//sysUnix "golang.org/x/sys/unix"

	"github.com/cespare/xxhash"
	log "github.com/sirupsen/logrus"
)

// FileInfo contains tar metadata for a specific FileNode
type FileInfo struct {
	// fully qualified file path: slash-delimited string from the root ('/') to the desired node (e.g. '/a/node/fqfp')
	Fqfp     string
	Linkname string
	hash     uint64
	Size     int64
	Mode     os.FileMode
	Uid      int   // User Id - owner of the file
	Gid      int   // Group Id - owner of the file
	Err      error // error discovered while retrieving metadata about this file, such as Insufficient Permission
}

//func GetXid(fqfp string, info os.FileInfo) (int, int) {
//	// FIXME: this section is incompatible with Plan9
//	UID := -1
//	GID := -1
//	if stat, ok := info.Sys().(*sysUnix.Stat_t); ok {
//		UID = int(stat.Uid)
//		GID = int(stat.Gid)
//	}
//	return UID, GID
//}

// NewFileInfo extracts the metadata from the info and file contents and generates a new FileInfo object.
func NewFileInfo(fqfp string, info os.FileInfo, err error) FileInfo {
	// FIXME: this section is incompatible with Plan9
	UID := -1
	GID := -1
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		UID = int(stat.Uid)
		GID = int(stat.Gid)
	}

	var hash uint64 = 0
	//if !info.IsDir() {
	//	hash = computeFileHash(fqfp)
	//}

	return FileInfo{
		Fqfp:     fqfp,
		Linkname: info.Name(),
		hash:     hash,
		Size:     info.Size(),
		Mode:     info.Mode(),
		Uid:      UID,
		Gid:      GID,
		Err:      err,
	}
}

func (info *FileInfo) String() string {
	return fmt.Sprintf("%v %v isDir=%v", info.Fqfp, info.Size, info.IsDir())
}

func (info *FileInfo) IsDir() bool {
	return info.Mode.IsDir()
}

// Clone clones given FileInfo
func (info *FileInfo) Clone() *FileInfo {
	if info == nil {
		return nil
	}
	return &FileInfo{
		Fqfp:     info.Fqfp,
		Linkname: info.Linkname,
		hash:     info.hash,
		Size:     info.Size,
		Mode:     info.Mode,
		Uid:      info.Uid,
		Gid:      info.Gid,
		Err:      info.Err,
	}
}

// Compare determines the DiffType between two FileInfos based on the type and contents of each given FileInfo
func (info *FileInfo) Compare(other FileInfo) DiffType {
	if info.Mode == other.Mode {
		if info.hash == other.hash &&
			info.Mode == other.Mode &&
			info.Uid == other.Uid &&
			info.Gid == other.Gid {
			return Unmodified
		}
	}
	return Modified
}

func computeFileHash(fqfp string) uint64 {
	//Open the passed argument and check for any error
	file, err := os.Open(fqfp)
	if err != nil {
		return 0
	}
	defer file.Close()

	h := xxhash.New()
	buf := make([]byte, 1024)

	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Panic(err)
		}
		if n == 0 {
			break
		}

		_, err = h.Write(buf[:n])
		if err != nil {
			log.Panic(err)
		}
	}

	return h.Sum64()
}
