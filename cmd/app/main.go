package main

import (
	"context"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

//go:generate go run ../gen/main.go

func main() {
	if err := run(); err != nil {
		log.Fatalf("app has run into an fatal error %v", err)
	}
}

func run() error {
	logger := log.New(os.Stdout, "APP : ", log.Lmicroseconds|log.Lmsgprefix)
	ctx := context.Background()

	httpPort := flag.String("HTTP_PORT", "8080", "app http port")
	flag.Parse()

	var wg sync.WaitGroup

	httpServer := newMainServer(logger)
	httpServer.Addr = ":" + *httpPort

	wg.Add(1)
	go func() {
		logger.Println("started serving http")
		if err := httpServer.ListenAndServe(); err != nil {
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

		httpServer.Shutdown(c)
	}()

	wg.Wait()
	return nil
}

//go:embed public/*
var public embed.FS

func newMainServer(logger *log.Logger) *http.Server {
	subPublic, err := fs.Sub(public, "public")
	if err != nil {
		panic(err)
	}
	publicHandler := http.FileServer(http.FS(subPublic))

	mux := http.NewServeMux()
	mux.Handle("/", loggerMiddleware(logger, themeMiddleware(publicHandler)))

	return &http.Server{Handler: mux}
}

func loggerMiddleware(logger *log.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Println("incoming request  : ", r.Method, r.URL.Path)

		now := time.Now()
		defer func() {
			logger.Println("completed request : ", r.Method, r.URL.Path, time.Since(now))
		}()

		rw.Header().Add("Cache-Control", "max-age=3600")
		h.ServeHTTP(rw, r)
	})
}

func themeMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		const themeSuffix = "theme.css"
		if !strings.HasSuffix(r.URL.Path, themeSuffix) {
			h.ServeHTTP(rw, r)
			return
		}

		pathPrefix := strings.TrimSuffix(r.URL.Path, themeSuffix)
		switch r.URL.Query().Get("theme") {
		case "light":
			r.URL.Path = pathPrefix + "light.css"
		default:
			r.URL.Path = pathPrefix + "dark.css"
		}

		h.ServeHTTP(rw, r)
	})
}
