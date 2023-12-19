##### December 20th, 2021

# How hard could it be to code a simple HTTPS server with Go?
#### The short answer: actually pretty easy...

In the last weeks I've been thinking about this little project that I'm working on, which is my own personal blog. I read about having a portfolio page and it seemed like a good idea, but, since I mainly work with backend, it could be quite challenging to find a way to showcase my work without exposing sensitive information of the projects I'm part of.

So, here I am, instead of going with an established media service, like Medium or LinkedIn, I decided to build my own website and, hopefully, turn it in a interesting resource to the devs out there. I know, I know this first paragraphs are not about HTTPS or TLS, or even Go, but I felt like, since this is my first post, I could talk a little about my motivations.

I really dig into the simplicity mindset behind all things Go, still, when I started my personal blog project, I thought: "this could be a nice opportunity to test my docker skills". And there I was, thinking about nginx and docker-compose...but...

I am a Go developer after all, so how hard could it be to build it myself?

Turns out its not hard at all...

## The long answer

Every time I stop to think about it, I am impressed by how much can be done with Go's standard library. It gives you many clean and extensible API's that can carry you a long way before you need to reach out for third party libraries and a HTTPS server is not a exception.

If you ever coded any simple HTTP server in Go you are only a few lines away from the HTTPS implementation. Since I'm not aiming to talk about how HTTPS guts work, I'll defer that to other articles, like [this](https://en.wikipedia.org/wiki/HTTPS) (from Wikipedia itself) and [this](https://eli.thegreenplace.net/2021/go-https-servers-with-tls/) (written by [Eli Bendersky](https://github.com/eliben)). You can check that out if want to know more about HTTPS, but, in Go, the simplest form is as simple as a handler function definition and a function call.


```go
//main.go
package main

import (
  "io"
  "log"
  "net/http"
)

const port = "8443"

func main() {
  http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, World!\n")
  })

  log.Printf("Starting server at port %s", port)
  // Here you pass as parameters the certificate and key to be used.
  // The last parameter is a http.Handler, if you pass nil the default
  // http handler will be used.
  err := http.ListenAndServeTLS(":"+port, "cert.pem", "key.pem", nil)
  log.Fatal(err)
}
```


You may be asking yourself if this is capable of serving production or if it fits only a simple test server. Well, that's the beauty of Go's standard library, the solutions are really robust. I've ran a little test using [Vegeta](https://github.com/tsenart/vegeta), my machine is not exactly a beefy one, but the results were no joke.

```
goos: linux
goarch: amd64
cpu: Intel(R) Core(TM) i5-8265U CPU @ 1.60GHz

Requests      [total, rate, throughput]         150000, 5000.03, 5000.01
Duration      [total, attack, wait]             30s, 30s, 121.729µs
Latencies     [min, mean, 50, 90, 95, 99, max]  67.644µs, 137.152µs, 108.307µs, 123.25µs, 129.877µs, 280.893µs, 37.418ms
Bytes In      [total, mean]                     2100000, 14.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:150000
Error Set:
```


## Multiple certificates

This is all great, but what if you wanted to serve two different domains with the same Go program? 

Not a problem, Go standard library supports multiple certificates. It is not only supported, but the standard library implementation will select the correct certificate for the requested URL. I used this feature to implement the "wildcard" and the "naked" domains in my server (that is serving you this blog post right now...).

```go
//main.go
package main

import (
  "crypto/tls"
  "fmt"
  "log"
  "net/http"
)

const port = "8443"

func main() {
  var mux http.ServeMux
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, world over HTTPS!\n")
  })

  var certs []tls.Certificate

  // The certicate and key files must contain PEM encoded data.
  cert, err := tls.LoadX509KeyPair("first.crt.pem", "first.key.pem")
  if err != nil {
    log.Fatal(err)
  }
  certs = append(certs, cert)

  cert, err = tls.LoadX509KeyPair("second.crt.pem", "second.key.pem")
  if err != nil {
    log.Fatal(err)
  }
  certs = append(certs, cert)

  tlsConfig := &tls.Config{
    // If you have any restrictions on TLS version you can use this
    // option.
    MinVersion:               tls.VersionTLS12,
    PreferServerCipherSuites: true,
    // Here we pass the certificates we loaded.
    Certificates: certs,
  }

  // Here we create the http server as usual, but set the TLSConfig
  // property to the one we created before.
  server := http.Server{
    Addr:      ":" + port,
    Handler:   &mux,
    TLSConfig: tlsConfig,
  }

  log.Printf("Starting server at port %s", port)
  // Since the certs are referenced inside the TLSConfig struct, we
  // can start the server passing empty strings as the arguments.
  err = server.ListenAndServeTLS("", "")
  log.Fatal(err)
}
```

## __Bonus__: All in one binary

If you miss the days where deploying an application was simply to copy some files to the server, this next section is for you.

Since [1.16](https://go.dev/doc/go1.16) Go has a feature that I'm in love with: Embed. Long story short, embed gives you the possibility of including static files on the final binary. This means that, instead of copying a bunch of files around, you could embed the files in the binary and have a single file to copy to your server in order to run your program (in our https server you should be careful with this file since it has the private key inside it).

Embed is such a simple and effective tool that I'm not only using for my certificates, but also for the actual public files of this website. It is a way of having a web server without it having access to the file system, which is a nice security feature. As I said before, really enjoy using it. Here is a simple code snippet of how to do it in our little server experiment.

```go
//main.go
package main

import (
  "crypto/tls"
  _ "embed"
  "fmt"
  "log"
  "net/http"
)

const port = "8443"

// Go embed uses directives to give you access to the functionality.
// After it you should write the path to the file you want to embed.
// This can also be done with directories, check it here:
// https://pkg.go.dev/embed#hdr-File_Systems

//go:embed first.crt.pem
var firstCrt []byte

//go:embed first.key.pem
var firstKey []byte

//go:embed second.crt.pem
var secondCrt []byte

//go:embed second.key.pem
var secondKey []byte

func main() {
  var mux http.ServeMux
  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, world over HTTPS!\n")
  })

  var certs []tls.Certificate

  // Changed from LoadX509KeyPair to X509KeyPair so we can pass the
  // bytes directly.
  cert, err := tls.X509KeyPair(firstCrt, firstKey)
  if err != nil {
    log.Fatal(err)
  }
  certs = append(certs, cert)

  cert, err = tls.X509KeyPair(secondCrt, secondKey)
  if err != nil {
    log.Fatal(err)
  }
  certs = append(certs, cert)

  // Everything else stays the same!
  .
  .
  .
```

Well, I feel like I have already gone little off topic here. So this must be a sign that I should end this post.

I know this is no novelty, but this blog is not about bleeding edge knowledge about Go, it is about this unique experience of being a developer and overcoming challenges every day. I really hope this reading has been of good use to you.
