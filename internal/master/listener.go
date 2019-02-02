/*
 * 侦听器
 * master的tcp侦听器：侦听端口，Accept仅阻塞
 *
 * wencan
 * 2019-01-13
 */

package master

import (
	"io"
	"net"
	"os"
	"sync"
)

// tcpListener master用的侦听器。监听，获取侦听文件，Accept仅阻塞
type tcpListener struct {
	network, laddr string

	raw net.Listener
	fd  *os.File

	closed    chan struct{}
	closeOnce sync.Once
	closeErr  error
}

func newTCPListner(network, laddr string) (*tcpListener, error) {
	ln := &tcpListener{
		network: network,
		laddr:   laddr,
		closed:  make(chan struct{}),
	}

	err := ln.listen()
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func (ln *tcpListener) ListenFile() *os.File {
	return ln.fd
}

func (ln *tcpListener) listen() error {
	var err error
	ln.raw, err = net.Listen(ln.network, ln.laddr)
	if err != nil {
		return err
	}

	ln.fd, err = ln.raw.(*net.TCPListener).File()
	if err != nil {
		return err
	}

	return nil
}

func (ln *tcpListener) Accept() (net.Conn, error) {
	<-ln.closed
	return nil, io.EOF
}

func (ln *tcpListener) Close() error {
	// 通知Accept返回
	// 由外层onceCloseListener确保不会重复close
	close(ln.closed)

	err := ln.raw.Close()
	if err != nil {
		return err
	}

	err = ln.fd.Close()
	if err != nil {
		return err
	}

	return nil
}

func (ln *tcpListener) Addr() (add net.Addr) {
	return ln.raw.Addr()
}
