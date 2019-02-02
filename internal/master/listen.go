/*
 * master的主逻辑
 *
 * wencan
 * 2019-01-25
 */

package master

import (
	"fmt"
	"net"
	"os"
	"sync"
)

// Listen 侦听地址，并创建worker进程处理该侦听
func Listen(network, laddr string) (net.Listener, error) {
	// 侦听，不接受连接
	ln, err := newTCPListner(network, laddr)
	if err != nil {
		return nil, err
	}

	// 创建worker
	files := []*os.File{ln.ListenFile(), rPipe()} // 侦听文件，和进程通讯用的读管道
	apendedEnv := []string{
		fmt.Sprintf("%s=1", _workerTag),                   // worker进程标志
		fmt.Sprintf("%s=%s#%s", _addrTag, network, laddr), // worker侦听地址
	}
	wker, err := newWorker(files, apendedEnv, network, laddr)
	if err != nil {
		return nil, err
	}

	// 两次包装
	// tcpListener:Close需要Close(chan)，为其添加onceCloseListener
	// 第二次确保侦听器和worker一起被关闭
	return newWrapListener(&onceCloseListener{Listener: ln}, wker), nil
}

// wrapListener 将侦听器和worker打包，用于一并关闭
type wrapListener struct {
	net.Listener

	wker *worker
}

func newWrapListener(raw net.Listener, wker *worker) *wrapListener {
	ln := &wrapListener{raw, wker}
	return ln
}

// Close 关闭侦听器，并等待worker退出
func (ln *wrapListener) Close() error {
	err := ln.Listener.Close()
	if err != nil {
		return err
	}

	err = ln.wker.wait()
	if err != nil {
		return err
	}
	return nil
}

// onceCloseListener 用于确保每个侦听器只被关闭一次
type onceCloseListener struct {
	net.Listener
	closeOnce sync.Once
	closeErr  error
}

func (ln *onceCloseListener) Close() error {
	ln.closeOnce.Do(func() { ln.closeErr = ln.Listener.Close() })
	return ln.closeErr
}
