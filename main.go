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
	"strings"
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

	httpPort := flag.String("HTTP_PORT", "8000", "app http port")
	httpsPort := flag.String("HTTPS_PORT", "8080", "app https port")
	mainHost := flag.String("MAIN_HOST", "localhost:8080", "app main host")
	flag.Parse()

	certs, err := loadCerts()
	if err != nil {
		return err
	}

	httpServer := newRedirectServer(logger, *mainHost)
	httpServer.Addr = ":" + *httpPort
	httpServer.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             certs,
	}

	logger.Println("starting service...")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		logger.Println("started serving http")
		if err := httpServer.ListenAndServe(); err != nil {
			logger.Printf("stoped serving http : %v", err)
		}
		wg.Done()
	}()

	httpsServer := newMainServer(logger, *mainHost)
	httpsServer.Addr = ":" + *httpsPort
	httpsServer.TLSConfig = &tls.Config{
		MinVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		Certificates:             certs,
	}

	wg.Add(1)
	go func() {
		logger.Println("started serving https")
		if err := httpsServer.ListenAndServeTLS("", ""); err != nil {
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
		httpsServer.Shutdown(c)
	}()

	wg.Wait()
	return nil
}

func newRedirectServer(logger *log.Logger, mainHost string) http.Server {
	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Println("incoming request  : ", r.Method, r.URL.Path)

		now := time.Now()
		defer logger.Println("completed request : ", r.Method, r.URL.Path, time.Since(now))

		rw.Header().Add("Location", "https://"+mainHost+r.URL.RequestURI())
		rw.WriteHeader(http.StatusMovedPermanently)
	}))

	return http.Server{Handler: mux}
}

//go:embed public/*
var public embed.FS

func newMainServer(logger *log.Logger, mainHost string) http.Server {
	subPublic, err := fs.Sub(public, "public")
	if err != nil {
		panic(err)
	}
	publicHandler := http.FileServer(http.FS(subPublic))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logger.Println("incoming request  : ", r.Method, r.URL.Path)

		now := time.Now()
		defer logger.Println("completed request : ", r.Method, r.URL.Path, time.Since(now))

		if strings.Contains(r.Host, ".tech") {
			rw.Header().Add("Location", "https://"+mainHost+r.URL.RequestURI())
			rw.WriteHeader(http.StatusMovedPermanently)
			return
		}

		rw.Header().Add("Cache-Control", "max-age=3600")
		publicHandler.ServeHTTP(rw, r)
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

//go:embed certs/cesarfuhr.tech.crt
var techCrt []byte

//go:embed certs/cesarfuhr.tech.key
var techKey []byte

func loadCerts() ([]tls.Certificate, error) {
	wildcard, err := tls.X509KeyPair(wildCrt, wildKey)
	if err != nil {
		return nil, err
	}

	naked, err := tls.X509KeyPair(nakedCrt, nakedKey)
	if err != nil {
		return nil, err
	}

	tech, err := tls.X509KeyPair(techCrt, techKey)
	if err != nil {
		return nil, err
	}

	certs := append([]tls.Certificate{}, wildcard, naked, tech)

	return certs, nil
}
