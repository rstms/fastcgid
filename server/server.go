package server

import (
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const Version = "0.0.1"

const SHUTDOWN_TIMEOUT = 5 * time.Second

type Server struct {
	verbose  bool
	shutdown chan struct{}
	listener net.Listener
}

func NewServer() (*Server, error) {

	log.Printf("fastcgid v%s uid=%d gid=%d started as PID %d", Version, os.Getuid(), os.Getgid(), os.Getpid())

	h := Server{
		verbose:  ViperGetBool("verbose"),
		shutdown: make(chan struct{}, 1),
	}

	return &h, nil
}

func (s *Server) fail(w http.ResponseWriter, status int) {
	w.WriteHeader(status)
}

/*
func (s *Server) logResponse(status int, message string, result interface{}) {
	log.Printf("<--response [%d] %s\n", status, message)
	if s.verbose && result != nil {
		log.Printf("response body: %s\n", FormatJSON(result))
	}
}
*/

func (s *Server) writeString(w http.ResponseWriter, line string) bool {
	_, err := io.WriteString(w, line)
	if err != nil {
		s.fail(w, http.StatusInternalServerError)
		return false
	}
	return true
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	log.Printf("ServeHTTP: url=%s\n", r.URL)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")
	if !s.writeString(w, "<h1>FastCGI Response</h1>\n") {
		return
	}

	if !s.writeString(w, fmt.Sprintf("<h2>URL</h2><code>%s</code>\n", r.URL)) {
		return
	}

	if !s.writeString(w, fmt.Sprintf("<h2>Query</h2><code>%v</code>\n", r.URL.Query())) {
		return
	}

	if !s.writeString(w, "<h2>Headers:</h2>\n") {
		return
	}
	if !s.writeString(w, "<ul>\n") {
		return
	}
	for header, values := range r.Header {
		for _, value := range values {
			if !s.writeString(w, fmt.Sprintf("<li><code>%s=%s</code></li>\n", header, html.EscapeString(value))) {
				return
			}
		}
	}
	if !s.writeString(w, "</ul>\n") {
		return
	}

	env := fcgi.ProcessEnv(r)
	if !s.writeString(w, "<h2>Environment:</h2>\n") {
		return
	}
	if !s.writeString(w, "<ul>\n") {
		return
	}
	for key, value := range env {
		if !s.writeString(w, fmt.Sprintf("<li><code>%s=%s</code></li>\n", key, html.EscapeString(value))) {
			return
		}
	}
	if !s.writeString(w, "</ul>\n") {
		return
	}
}

func (s *Server) Run() error {
	err := s.Start()
	if err != nil {
		return err
	}
	err = s.Wait()
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Start() error {

	addr := ViperGetString("addr")
	port := ViperGetInt("port")
	listenAddress := fmt.Sprintf("%s:%d", addr, port)

	go func() {
		defer log.Println("Exiting FastCGI request handler")
		var err error
		s.listener, err = net.Listen("tcp", listenAddress)
		if err != nil {
			log.Printf("%s\n", Fatalf("Listen failed: %v", err))
		}
		log.Printf("Serving FastCGI on %s\n", listenAddress)
		err = fcgi.Serve(s.listener, s)
		if err != nil && err != net.ErrClosed {
			log.Printf("%s\n", Fatalf("FastCGI Serve failed: %v", err))
		}
	}()

	return nil
}

func (s *Server) Wait() error {

	if s.verbose {
		log.Println("Wait: waiting...")
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	if s.verbose {
		fmt.Println("CTRL-C to exit")
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-sigint:
				log.Println("received SIGINT")
				return
			case <-sigterm:
				log.Println("received SIGTERM")
				return
			case <-s.shutdown:
				log.Println("received shutdown request")
				return
			}
		}
	}()

	wg.Wait()

	if s.verbose {
		log.Println("Wait: initiating shutdown")
	}

	err := s.listener.Close()
	if err != nil {
		return Fatal(err)
	}

	if s.verbose {
		log.Println("Wait: shutdown complete")
	}
	return nil
}

func (s *Server) Stop() error {
	if s.verbose {
		log.Println("Stop: requesting shutdown")
	}
	s.shutdown <- struct{}{}
	err := s.Wait()
	if err != nil {
		return Fatal(err)
	}
	if s.verbose {
		log.Println("Stop: stopped")
	}
	return nil
}
