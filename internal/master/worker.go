/*
 * 创建worker（进程），以及对worker的一些操作
 *
 * wencan
 * 2019-01-18
 */

package master

import (
	"log"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

var (
	_workers sync.Map
)

type worker struct {
	files      []*os.File
	apendedEnv []string

	cmd *exec.Cmd

	network, laddr string

	wg sync.WaitGroup
}

func registryWorker(wker *worker) {
	_workers.Store(wker, struct{}{})
}

func unregistryWoker(wker *worker) {
	_workers.Delete(wker)
}

// Reloads 全部worker重新加载，不等待旧worker进程退出
func Reloads() {
	_workers.Range(func(key, value interface{}) bool {
		wker := key.(*worker)
		err := wker.reload()
		if err != nil {
			log.Println(err)
		}
		return true
	})
}

// // Kills 通知全部worker退出，不等待
// func Kills() {
// 	_workers.Range(func(key, value interface{}) bool {
// 		wker := key.(*worker)
// 		wker.kill()
// 		return true
// 	})
// }

// // Waits 等待全部worker退出
// func Waits() {
// 	_workers.Range(func(key, value interface{}) bool {
// 		wker := key.(*worker)
// 		wker.wait()
// 		return true
// 	})
// }

// execWorker 创建worker进程
func execWorker(files []*os.File, apendedEnv []string, network, laddr string) (*exec.Cmd, error) {
	path := os.Args[0]

	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}

	env := append(os.Environ(), apendedEnv...)

	cmd := exec.Command(path, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	cmd.Env = env

	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

// newWorker 创建worker
func newWorker(files []*os.File, apendedEnv []string, network, laddr string) (*worker, error) {
	cmd, err := execWorker(files, apendedEnv, network, laddr)
	if err != nil {
		return nil, err
	}

	wker := &worker{
		files:      files,
		apendedEnv: apendedEnv,
		cmd:        cmd,
		network:    network,
		laddr:      laddr,
	}

	registryWorker(wker)
	return wker, nil
}

// Wait 等待worker全部相关进程退出
func (wker *worker) wait() error {
	// 等待全部等待进程退出的gorountine退出
	wker.wg.Wait()
	return nil
}

// Kill 通知worker进程退出，不等待
func (wker *worker) kill() error {
	err := syscall.Kill(wker.cmd.Process.Pid, syscall.SIGHUP)
	if err != nil {
		return err
	}

	// 等待进程退出——父进程的责任
	cmd := wker.cmd // 后面wker.cmd可能会被替换
	wker.wg.Add(1)
	go func() {
		defer wker.wg.Done()

		err := cmd.Wait()
		if err != nil {
			log.Println("worker process exited:", err)

			// 如果当前worker进程异常退出，master进程也异常退出，其它worker进程跟进
			// 重启工作交给监护程序
			if cmd.Process.Pid == wker.cmd.Process.Pid {
				exitCode := 1
				status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
				if ok {
					exitCode = status.ExitStatus()
				}
				os.Exit(exitCode)
			}
		}
	}()

	return nil
}

// Reload 通知worker退出，不等待，创建新的worker进程
func (wker *worker) reload() error {
	err := wker.kill()
	if err != nil {
		return err
	}

	// 创建新的worker
	cmd, err := execWorker(wker.files, wker.apendedEnv, wker.network, wker.laddr)
	if err != nil {
		return err
	}

	wker.cmd = cmd

	return nil
}
