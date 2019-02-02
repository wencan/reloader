/*
 * reloader主逻辑
 *
 * wencan
 * 2019-01-13
 */

package reloader

import (
	"net"
	"os"

	"github.com/wencan/reloader/internal/master"
	"github.com/wencan/reloader/internal/worker"
)

const (
	// _workerTag 环境变量key，标示进程是否为worker
	_workerTag = "_RELOADER_WORKER"

	// _addrTag 环境变量worker侦听地址key
	_addrTag = "_RELOADER_LISTEN_ADDR"
)

var (
	_isWorker *bool
)

// Listener reloader的侦听器
// 支持自定义关闭处理。直接关闭net.Listener会引起上层逻辑出错
type Listener interface {
	net.Listener
	OnTerm(func())
}

// fakeListener 确保Listener实现Listener
type fakeListener struct {
	net.Listener
}

func (ln *fakeListener) OnTerm(f func()) {}

// IsWorker 当前进程是否为worker
func IsWorker() bool {
	// 确保任一init过程都能读到已解析的isWorker参数
	if _isWorker == nil {
		b := os.Getenv(_workerTag) != ""
		_isWorker = &b
	}

	return *_isWorker
}

// IsMaster 当前进程是否为master
func IsMaster() bool {
	return !IsWorker()
}

// Listen 侦听，返回侦听器对象。
// master进程侦听端口，并为每个地址创建一个worker进程。
// worker进程仅处理自身负责的侦听，其它侦听使用idleListener做阻塞处理
func Listen(network, laddr string) (Listener, error) {
	switch network {
	case "tcp", "tcp4", "tcp6":
	default:
		panic(net.UnknownNetworkError(network))
	}

	if IsMaster() {
		ln, err := master.Listen(network, laddr)
		if err != nil {
			return nil, err
		}
		// master只返回一个阻塞的侦听器
		return &fakeListener{Listener: ln}, nil
	}

	// _isWorker
	return worker.Listen(network, laddr)
}
