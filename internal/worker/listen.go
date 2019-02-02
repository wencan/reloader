/*
 * worker主逻辑
 *
 * wencan
 * 2019-01-25
 */

package worker

import (
	"log"
	"net"
	"sync"
)

var (
	_listeners sync.Map
)

// Listen 侦听自己进程负责的地址，其它的地址仅阻塞Accept
func Listen(network, laddr string) (*Listener, error) {
	if _lnNetwork == network && _lnAddr == laddr {
		// 基于文件侦听，并接受连接
		// 每个worker进程只会有一个fileListener
		ln, err := newFileListener(network, laddr, _lnFile)
		if err != nil {
			return nil, err
		}

		// 如果fileListener:Close重入，会报use of closed network connection
		return newListener(&onceCloseListener{Listener: ln}), nil
	}

	// 不处理该侦听
	ln, err := newIdleListener(network, laddr, _lnFile)
	if err != nil {
		return nil, err
	}

	// idleListener:Close需要Close(chan)，为其添加onceCloseListener
	return newListener(&onceCloseListener{Listener: ln}), nil
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

// Listener worker的侦听器统一包装，支持上层逻辑定义TERM处理
type Listener struct {
	net.Listener

	termHandle func()
	termOnce   sync.Once
	termErr    error
}

func newListener(ln net.Listener) *Listener {
	l := &Listener{Listener: ln}
	_listeners.Store(l, struct{}{})
	return l
}

func (ln *Listener) Close() error {
	_listeners.Delete(ln)
	return ln.Listener.Close()
}

// OnTerm 自定义term处理逻辑
func (ln *Listener) OnTerm(f func()) {
	ln.termHandle = f
}

// term 终止侦听器
func (ln *Listener) term() error {
	ln.termOnce.Do(func() {
		// 优先调用上层定义的处理逻辑
		// 上层逻辑应该Close侦听器
		if ln.termHandle != nil {
			ln.termHandle()
		} else {
			// log.Println("TERM event handler is not defined", ln.Addr().String())
			// 如果上层未定义Term处理，直接关闭侦听器
			// 这易导致上层逻辑出错
			ln.termErr = ln.Close()
		}
	})

	return ln.termErr
}

// Terms 终止全部侦听器
func Terms() {
	_listeners.Range(func(key, value interface{}) bool {
		go func() {
			ln := key.(*Listener)
			err := ln.term()
			if err != nil {
				log.Println(err)
			}
		}()
		return true
	})
}
