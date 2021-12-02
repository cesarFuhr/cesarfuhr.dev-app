package main

import (
	"context"
	"crypto/tls"
	"embed"
	"flag"
	"io/fs"
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

	certs, err := loadCerts()
	if err != nil {
		return err
	}

	server := newHandler(logger)
	server.Addr = ":" + *port
	server.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             certs,
	}

	logger.Println("starting service...")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		logger.Println("started serving https")
		if err := server.ListenAndServeTLS("", ""); err != nil {
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

//go:embed public/*
var public embed.FS

func newHandler(logger *log.Logger) http.Server {
	subPublic, err := fs.Sub(public, "public")
	if err != nil {
		panic(err)
	}
	publicHandler := http.FileServer(http.FS(subPublic))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Println("incoming request  : ", r.Method, r.URL.Path)

		now := time.Now()

		rw.Header().Add("Cache-Control", "max-age=3600")
		publicHandler.ServeHTTP(rw, r)

		logger.Println("completed request : ", r.Method, r.URL.Path, time.Since(now))
	}))

	return http.Server{Handler: mux}
}

//go:embed certs/wildcesarfuhr.crt
var wildCrt []byte

//go:embed certs/wildcesarfuhr.key
var wildKey []byte

//go:embed certs/cesarfuhr.crt
var nakedCrt []byte

//go:embed certs/cesarfuhr.key
var nakedKey []byte

func loadCerts() ([]tls.Certificate, error) {
	wildcard, err := tls.X509KeyPair(wildCrt, wildKey)
	if err != nil {
		return nil, err
	}

	naked, err := tls.X509KeyPair(nakedCrt, nakedKey)
	if err != nil {
		return nil, err
	}

	certs := append([]tls.Certificate{}, wildcard, naked)

	return certs, nil
}
