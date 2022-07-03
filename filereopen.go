package filereopen

import (
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"
)

type File struct {
	filename     string
	perm         os.FileMode
	currentFd    *os.File
	inode        uint64
	interval     time.Duration
	closed       bool
	errorHandler func(error)
}

func OpenFileForAppend(name string, perm os.FileMode) (*File, error) {
	file := File{}
	fd, err := os.OpenFile(name, os.O_WRONLY|os.O_APPEND|os.O_CREATE, perm)
	if err != nil {
		return nil, err
	}
	// get inode of file we actually opened
	fileinfo, _ := fd.Stat()
	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		return nil, fmt.Errorf("could not get inode info: Not a syscall.Stat_t")
	}
	file.inode = stat.Ino
	file.perm = perm
	file.currentFd = fd
	file.filename = fd.Name()
	file.interval = time.Second
	file.errorHandler = func(e error) {}
	if err != nil {
		return nil, err
	}
	go file.inodeWatcher()
	return &file, nil
}

type fileNotFoundErr struct{}

func (e fileNotFoundErr) Error() string {
	return ""
}

// SetInterval sets interval for polling
//
// The interval will be applied after one cycle.
//
// Will not allow shorter than 100ms for sanity reasons, you probably want a package
// that uses inotify if you want anything faster
//
func (f *File) SetInterval(t time.Duration) error {
	if t < (100 * time.Millisecond) {
		return fmt.Errorf("intervals below 100ms not allowed")
	}
	f.interval = t
	return nil
}
func (f *File) inodeWatcher() {
	for {
		time.Sleep(f.interval)
		if f.closed {
			f.currentFd.Sync()
			f.currentFd.Close()
			return
		}
		inode, err := f.getFilenameInode()
		if err == nil {
			if inode == f.inode {
				continue
			}
			err := f.Reopen()
			if err != nil {
				f.errorHandler(err)
			}
		} else {
			err := f.Reopen()
			if err != nil {
				f.errorHandler(err)
			}
		}
	}
}

func (f *File) getFdInode() (uint64, error) {
	fileinfo, err := f.currentFd.Stat()
	if err != nil {
		return uint64(time.Now().UnixMicro()), &fileNotFoundErr{}
	}
	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		return uint64(time.Now().UnixMicro()), fmt.Errorf("could not get inode info: Not a syscall.Stat_t")
	}
	return stat.Ino, nil
}

func (f *File) getFilenameInode() (uint64, error) {
	fileinfo, err := os.Stat(f.filename)
	if err != nil {
		return uint64(time.Now().UnixMicro()), &fileNotFoundErr{}
	}
	stat, ok := fileinfo.Sys().(*syscall.Stat_t)
	if !ok {
		return uint64(time.Now().UnixMicro()), fmt.Errorf("could not get inode info: Not a syscall.Stat_t")
	}
	return stat.Ino, nil
}

func (f *File) Reopen() error {
	oldFd := f.currentFd
	newFd, err := os.OpenFile(f.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, f.perm)
	if err != nil {

		return err
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&f.currentFd)), unsafe.Pointer(newFd))
	// synchronously sync
	oldFd.Sync()
	//but close later just in case.
	go func(fd *os.File) {
		time.Sleep(time.Second * 10)
		fd.Close()
	}(oldFd)
	inode, err := f.getFdInode()
	if err != nil {
		return fmt.Errorf("error getting inode of just re-opened file: %s", err)
	}
	f.inode = inode
	return nil
}

func (f *File) SetErrorFunction(errFunc func(e error)) {
	f.errorHandler = errFunc
}

func (f *File) Write(b []byte) (n int, err error) {
	return f.currentFd.Write(b)
}
func (f *File) WriteAt(b []byte, off int64) (n int, err error) {
	return f.currentFd.WriteAt(b, off)
}
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	return f.currentFd.Seek(offset, whence)
}
func (f *File) Sync() error {
	return f.currentFd.Sync()
}
func (f *File) Close() error {
	f.closed = true
	return f.currentFd.Close()
}
