package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("app has run into an fatal error %v", err)
	}
}

func run() error {
	logger := log.New(os.Stdout, "APP : ", log.Lmicroseconds|log.Lmsgprefix)
	ctx := context.Background()

	port := flag.String("PORT", "8080", "app https port")
	flag.Parse()

	server := newHandler(logger)
	server.Addr = ":" + *port
	server.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
	}

	logger.Println("starting service...")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		logger.Println("started serving https")
		if err := server.ListenAndServeTLS("full.crt", "priv.key"); err != nil {
			logger.Printf("stoped serving https : %v", err)
		}
		wg.Done()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s := <-sigs
		logger.Printf("shuting down : SIG %v", s)

		c, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()

		server.Shutdown(c)
	}()

	wg.Wait()
	return nil
}

func newHandler(logger *log.Logger) http.Server {
	publicHandler := http.FileServer(http.Dir("./public"))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Println("incoming request  : ", r.Method, r.URL.Path)

		now := time.Now()
		publicHandler.ServeHTTP(rw, r)

		logger.Println("completed request : ", r.Method, r.URL.Path, time.Since(now))
	}))

	return http.Server{Handler: mux}
}
