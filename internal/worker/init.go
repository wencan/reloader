/*
 * worker初始化
 * 解析master传递的数据
 * 附加一个读取读管道，跟随master退出的逻辑
 *
 * wencan
 * 2019-01-25
 */

package worker

import (
	"log"
	"os"
	"strings"
)

const (
	// _workerTag 环境变量key，标示进程是否为worker
	_workerTag = "_RELOADER_WORKER"

	// _addrTag 环境变量worker侦听地址key
	_addrTag = "_RELOADER_LISTEN_ADDR"
)

var (
	_rPipe              *os.File
	_lnNetwork, _lnAddr string
	_lnFile             *os.File
)

func init() {
	if os.Getenv(_workerTag) == "" {
		// isMaster
		return
	}

	// 环境变量中的侦听地址信息
	// 一个worker进程只侦听（并接受）一个地址，其余地址只阻塞Accept
	addrs := os.Getenv(_addrTag)
	if addrs == "" {
		log.Println("not found addr variable")
		os.Exit(-1)
	}
	parts := strings.SplitN(addrs, "#", 2)
	if len(parts) != 2 {
		log.Println("addr variable invalid")
		os.Exit(-1)
	}
	_lnNetwork = parts[0]
	_lnAddr = parts[1]

	// 继承的侦听文件
	// 前三个为标准文件管道
	// 后一个为读管道文件
	_lnFile = os.NewFile(uintptr(3), "")
	if _lnFile == nil {
		log.Println("not found listen file")
		os.Exit(-1)
	}

	// 父进程创建的读管道
	_rPipe = os.NewFile(uintptr(4), "")
	if _rPipe == nil {
		log.Println("not found pipe file")
		os.Exit(-1)
	}

	go func() {
		// 当父进程退出，读管道返回EOF
		buff := make([]byte, 1)
		_rPipe.Read(buff)
		// 跟随退出
		// log.Println("parent exited")
		Terms()
		os.Exit(-1)
	}()
}
