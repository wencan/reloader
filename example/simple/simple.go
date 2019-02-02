package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/wencan/reloader"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	if reloader.IsMaster() {
		log.Println("master process:", os.Getpid())
	} else {
		log.Println("worker process:", os.Getpid())
	}

	ln, err := reloader.Listen("tcp", "127.0.0.1:8090")
	if err != nil {
		log.Println(err)
		return
	}

	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)

		for i := 0; i < 10; i++ {
			time.Sleep(time.Second)
			fmt.Fprintf(w, "pid#%d: Hello world\n", os.Getpid())
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
