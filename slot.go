/*
 * 信号处理逻辑
 * 只处理HUP信号
 *
 * wencan
 * 2019-01-24
 */

package reloader

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/wencan/reloader/internal/master"
	"github.com/wencan/reloader/internal/worker"
)

func init() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP)

	if IsMaster() {
		go masterSignalServe(ch)
	} else {
		go workerSignalServe(ch)
	}
}

func masterSignalServe(ch <-chan os.Signal) {
	for sig := range ch {
		switch sig {
		case syscall.SIGHUP:
			master.Reloads()
		default:
		}
	}
}

func workerSignalServe(ch <-chan os.Signal) {
	for sig := range ch {
		switch sig {
		case syscall.SIGHUP:
			worker.Terms() //Terms之后，上层逻辑应该让worker进程正常结束
		default:
		}
	}
}
