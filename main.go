// Blobabase is a dead-simple, thread-safe blobabase.

package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var ErrorNoSuchKey = errors.New("no such key")

type Blobabase struct {
	Blobs map[string][]byte
	mu    sync.RWMutex
}

func (c *Blobabase) Set(key string, blob []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Blobs[key] = blob
	return nil
}

func (c *Blobabase) Get(key string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	blob, ok := c.Blobs[key]
	if !ok {
		return nil, ErrorNoSuchKey
	}
	return blob, nil
}

type server struct {
	store *Blobabase
}

func (s *server) handleSet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if len(q) == 0 {
		http.Error(w, "no key=value pairs", http.StatusBadRequest)
		return
	}
	for name, vals := range q {
		s.store.Set(name, []byte(vals[len(vals)-1]))
	}
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	blob, err := s.store.Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(blob)
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /set", s.handleSet)
	mux.HandleFunc("GET /get", s.handleGet)
	return logging(mux)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		slog.Info("request", "method", r.Method, "path", r.URL.Path)
	})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	showVersion := flag.Bool("version", false, "print version and exit")
	port := flag.Int("port", 4000, "set port on which to listen")

	flag.Parse()

	if *showVersion {
		fmt.Printf("blobabase %s (commit %s, build %s)\n", version, commit, date)
		return
	}

	if *port <= 1023 {
		slog.Error("cannot bind to privileged port", "port", *port)
		os.Exit(1)
	}

	store := &Blobabase{Blobs: make(map[string][]byte)}
	srv := &server{store: store}

	s := http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      srv.routes(),
	}

	slog.Info("listening", "addr", s.Addr)
	if err := s.ListenAndServe(); err != nil {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
