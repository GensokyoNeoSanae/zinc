package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Lock 是仓库级的独占锁，防止两个任务同时改索引
type Lock struct {
	path  string
	token string
}

// acquireLock 尝试获取锁，失败时等待 release，总等待超过 timeout 则报错。
// 不自动接管陈旧锁——大文件备份可能跑很久，自动接管会导致重复压缩。
func acquireLock(dir string, timeout time.Duration) (*Lock, error) {
	path := filepath.Join(dir, ".lock")
	deadline := time.Now().Add(timeout)
	announced := false
	for {
		f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err == nil {
			token := fmt.Sprintf("%d-%d", os.Getpid(), time.Now().UnixNano())
			fmt.Fprintf(f, "pid=%d\ntime=%s\ntoken=%s\n", os.Getpid(), time.Now().Format(time.RFC3339), token)
			f.Close()
			return &Lock{path: path, token: token}, nil
		}
		if !os.IsExist(err) {
			return nil, fmt.Errorf(M.CreateLockFailed, err)
		}
		if !announced {
			fmt.Printf(M.WaitingLock, timeout)
			announced = true
		}
		if time.Now().After(deadline) {
			return nil, fmt.Errorf(M.LockTimeout, timeout, path)
		}
		time.Sleep(5 * time.Second)
	}
}

// release 只在锁文件仍属于自己时才删除，避免误删别的进程重新创建的锁。
func (l *Lock) release() {
	if l == nil {
		return
	}
	b, err := os.ReadFile(l.path)
	if err != nil {
		return
	}
	if strings.Contains(string(b), "token="+l.token) {
		os.Remove(l.path)
	}
}
