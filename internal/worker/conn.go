/*
 * 连接的包装。
 * worker侦听器返回的每个连接都带计数器。
 * 侦听器Close时，等待全部连接都Close了才能返回。
 *
 * wencan
 * 2019-01-13
 */

package worker

import (
	"net"
	"sync"
)

// refConn 连接。嵌入net.Conn，引用侦听器计数器
type refConn struct {
	net.Conn

	wg *sync.WaitGroup

	closeOnce sync.Once
	closeErr  error
}

// newConn 创建连接。将计数器加一
func newConn(conn net.Conn, wg *sync.WaitGroup) *refConn {
	wg.Add(1)
	return &refConn{
		Conn: conn,
		wg:   wg,
	}
}

// Close 关闭连接。将计数器减一
func (conn *refConn) Close() error {
	conn.closeOnce.Do(func() {
		conn.closeErr = conn.Conn.Close()
		conn.wg.Done()
	})
	return conn.closeErr
}
