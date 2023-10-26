package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	log.Print("main: start")

	counter := atomic.Int64{}
	srv := http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := counter.Add(1)
			log.Printf("handler: start [%3d]", v)
			time.Sleep(time.Second + time.Millisecond*time.Duration(rand.Int63n(5000)))
			fmt.Fprintf(w, "hello %d\n", v)
			log.Printf("handler: exit [%3d]", v)
		}),
	}

	stop := make(chan os.Signal, 1)
	kill := make(chan os.Signal, 1)
	closed := make(chan struct{}, 1)

	signal.Notify(stop, os.Interrupt)
	signal.Notify(kill, os.Kill)

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Print("goroutine: start")
		defer log.Print("goroutine: exit")

		for {
			select {
			case <-stop:
				log.Print("goroutine: interrupt signal: server shutdown: start")

				if err := srv.Shutdown(context.TODO()); err != nil {
					log.Print("goroutine: interrupt signal: server shutdown: ", err)
				}

				log.Print("goroutine: interrupt signal: server shutdown: exit")
				// do nothing
			case <-kill:
				log.Print("goroutine: kill signal")
				// do noting?
			case <-closed:
				log.Print("goroutine: server closed")
				return
			}
		}

	}()

	log.Print("server: listen start")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Print("server error: ", err)
	}
	log.Print("server: listen exit")

	closed <- struct{}{}

	wg.Wait()

	log.Print("main: exit")
}
