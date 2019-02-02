package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/wencan/reloader"
)

func httpServe(wg *sync.WaitGroup, network, addr string) {
	defer wg.Done()

	ln, err := reloader.Listen(network, addr)
	if err != nil {
		log.Println(err)
		return
	}

	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)

		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			fmt.Fprintf(w, "pid#%d: Hello world, No: %d\n", os.Getpid(), i)
			if flusher != nil {
				flusher.Flush()
			}
		}
	})}
	// （optional）自定义收到终止事件时的处理
	ln.OnTerm(func() {
		err := srv.Close()
		if err != nil {
			log.Println(err)
		}
	})
	err = srv.Serve(ln)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if reloader.IsMaster() {
		log.Println("master process:", os.Getpid())
	} else {
		log.Println("worker process:", os.Getpid())
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go httpServe(&wg, "tcp", fmt.Sprintf("127.0.0.1:809%d", i))
	}

	if reloader.IsMaster() {
		for {
			time.Sleep(time.Second)
			syscall.Kill(os.Getpid(), syscall.SIGHUP)
		}
	} else {
		wg.Wait()
	}
}
