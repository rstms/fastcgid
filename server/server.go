package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

const Version = "1.0.2"

const SHUTDOWN_TIMEOUT = 5 * time.Second

type Server struct {
	verbose    bool
	shutdown   chan struct{}
	listener   net.Listener
	ScriptRoot string
}

func NewServer() (*Server, error) {

	log.Printf("fastcgid v%s uid=%d gid=%d started as PID %d", Version, os.Getuid(), os.Getgid(), os.Getpid())
	ViperSetDefault("root", "/var/www")
	h := Server{
		verbose:    ViperGetBool("verbose"),
		shutdown:   make(chan struct{}, 1),
		ScriptRoot: ViperGetString("root"),
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

var URL_PATTERN = regexp.MustCompile(`.*cgi-bin/([^?]*)[?](.*)$`)
var HEADER_PATTERN = regexp.MustCompile(`^([^:]*):\s*(.*)\s*$`)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	if s.verbose {
		log.Printf("request: %s\n", r.URL)
	}

	fields := URL_PATTERN.FindStringSubmatch(r.URL.String())
	var script string
	var query string
	if len(fields) > 2 {
		script = fields[1]
		query = fields[2]
	}

	if s.verbose {
		log.Printf("script: %s\n", script)
		log.Printf("query: %s\n", query)
	}

	env := fcgi.ProcessEnv(r)
	env["REQUEST_URI"] = r.URL.String()
	env["SCRIPT_NAME"] = script
	env["QUERY_STRING"] = query
	addr, port, _ := strings.Cut(r.RemoteAddr, ":")
	env["REMOTE_ADDR"] = addr
	env["REMOTE_PORT"] = port
	env["REQUEST_METHOD"] = r.Method

	scriptPath := filepath.Join(s.ScriptRoot, env["SCRIPT_FILENAME"])

	if s.verbose {
		log.Printf("scriptPath: %s\n", scriptPath)
	}

	cmd := exec.Command(scriptPath)
	cmd.Env = os.Environ()
	for key, value := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	if s.verbose {
		for key, value := range env {
			log.Printf("env: %s=%s\n", key, value)
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/html")

	var obuf bytes.Buffer
	var ebuf bytes.Buffer

	cmd.Stdin = r.Body
	cmd.Stdout = &obuf
	cmd.Stderr = &ebuf
	err := cmd.Run()
	if err != nil {
		s.fail(w, http.StatusInternalServerError)
		return
	}

	if ebuf.Len() > 0 {
		log.Printf("script %s stderr: %s\n", scriptPath, ebuf.String())
	}
	scanner := bufio.NewScanner(&obuf)
	var headerComplete bool
	for scanner.Scan() {
		line := scanner.Text()
		if s.verbose {
			log.Printf("stdout: %s\n", line)
		}
		switch {
		case headerComplete:
			if !s.writeString(w, line+"\n") {
				return
			}
		case line == "":
			headerComplete = true
		default:
			fields := HEADER_PATTERN.FindStringSubmatch(line)
			if len(fields) == 3 {
				w.Header().Set(fields[1], fields[2])
			} else {
				log.Printf("failed parsing output header: %s\n", line)
			}
		}
	}
	err = scanner.Err()
	if err != nil {
		s.fail(w, http.StatusInternalServerError)
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
