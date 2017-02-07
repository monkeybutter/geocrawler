package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"syscall"
)

type POSIXDescriptor struct {
	GID   uint32 `json:"gid"`
	Group string `json:"group"`
	UID   uint32 `json:"uid"`
	User  string `json:"user"`
	Size  int64  `json:"size"`
	Mode  string `json:"mode"`
	Type  string `json:"type"`
	INode uint64 `json:"inode"`
	MTime int64  `json:"mtime"`
	ATime int64  `json:"atime"`
	CTime int64  `json:"ctime"`
}

type PosixInfo struct {
	In    chan string
	Out   chan string
	Error chan error
}

func NewPosixInfo(errChan chan error) *PosixInfo {
	return &PosixInfo{
		In:    make(chan string, 100),
		Out:   make(chan string, 100),
		Error: errChan,
	}
}

func (pi *PosixInfo) Run() {
	defer close(pi.Out)

	for path := range pi.In {
		finfo, err := os.Stat(path)
		if err != nil {
			pi.Error <- err
			continue
		}
		var rec syscall.Stat_t
		err = syscall.Lstat(path, &rec)
		if err != nil {
			pi.Error <- err
			continue
		}
		gid, err := user.LookupGroupId(fmt.Sprintf("%d", rec.Gid))
		if err != nil {
			pi.Error <- err
			continue
		}
		uid, err := user.LookupId(fmt.Sprintf("%d", rec.Uid))
		if err != nil {
			pi.Error <- err
			continue
		}

		descr := POSIXDescriptor{Group: gid.Name, User: uid.Username, INode: rec.Ino, MTime: rec.Mtim.Sec, CTime: rec.Ctim.Sec, ATime: rec.Atim.Sec, Size: finfo.Size(), Mode: finfo.Mode().String(), Type: "file", UID: rec.Uid, GID: rec.Gid}

		out, _ := json.Marshal(&descr)
		fmt.Printf("%s\tposix\t%s\n", path, string(out))

		pi.Out <- path
	}
}
