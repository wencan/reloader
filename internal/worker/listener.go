/*
 * 侦听器
 *
 * 包含
 * 文件侦听器：继承文件，转换为侦听器，然后接受连接
 * idle侦听器：一个worker进程只处理自身负责的一个侦听，其它的使用idle侦听器阻塞Accept
 *
 * wencan
 * 2019-01-13
 */

package worker

import (
	"io"
	"net"
	"os"
	"sync"
)

// fileListener worker用的侦听器，使用master的侦听文件接受连接
type fileListener struct {
	network, laddr string

	raw net.Listener
	fd  *os.File

	wg sync.WaitGroup
}

func newFileListener(network, laddr string, fd *os.File) (*fileListener, error) {
	ln := &fileListener{
		network: network,
		laddr:   laddr,
		fd:      fd,
	}

	err := ln.listen()
	if err != nil {
		return nil, err
	}

	return ln, nil
}

func (ln *fileListener) listen() error {
	var err error
	ln.raw, err = net.FileListener(ln.fd)
	if err != nil {
		return err
	}
	return nil
}

func (ln *fileListener) Accept() (net.Conn, error) {
	conn, err := ln.raw.Accept()
	if err != nil {
		return nil, err
	}

	// 包装连接，引用计数器
	c := newConn(conn, &ln.wg)
	return c, nil
}

func (ln *fileListener) Close() error {
	// 先停止侦听
	// 即使发生错误。也必须等待所有连接关闭
	// 如果Close重入，就可能发生ln.raw.Close()错误
	err := ln.raw.Close()
	e := ln.fd.Close()
	if e != nil && err == nil {
		err = e
	}

	// 等待所有连接关闭
	ln.wg.Wait()

	return err
}

func (ln *fileListener) Addr() (add net.Addr) {
	return ln.raw.Addr()
}

type addr struct {
	network, addr string
}

func (a *addr) Network() string {
	return a.network
}

func (a *addr) String() string {
	return a.addr
}

// idleListener worker用的侦听器。Accept仅阻塞。
type idleListener struct {
	laddr *addr

	closed chan struct{}
}

func newIdleListener(network, laddr string, fd *os.File) (*idleListener, error) {
	ln := &idleListener{
		laddr:  &addr{network, laddr},
		closed: make(chan struct{}),
	}

	return ln, nil
}

func (ln *idleListener) Accept() (net.Conn, error) {
	<-ln.closed
	return nil, io.EOF
}

func (ln *idleListener) Close() error {
	// 通知Accept返回
	// 由外层onceCloseListener确保不会重复close
	close(ln.closed)

	return nil
}

func (ln *idleListener) Addr() (add net.Addr) {
	return ln.laddr
}
