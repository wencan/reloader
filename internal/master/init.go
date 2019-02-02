/*
 * 初始化工作
 * 创建公共通讯管道
 *
 * wencan
 * 2019-01-25
 */

package master

import (
	"log"
	"os"
)

const (
	// _workerTag 环境变量key，标示进程是否为worker
	_workerTag = "_RELOADER_WORKER"

	// _addrTag 环境变量worker侦听地址key
	_addrTag = "_RELOADER_LISTEN_ADDR"
)

var (
	_wPipe, _rPipe *os.File
)

func init() {
	if os.Getenv(_workerTag) != "" {
		// isWorker
		return
	}

	// 创建公共通讯管道
	// 当父进程因任何原因退出，管道关闭——造成子进程读管道失败，跟随退出
	var err error
	_rPipe, _wPipe, err = os.Pipe()
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}

func rPipe() *os.File {
	return _rPipe
}
