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
	blobs map[string]string
	mu    sync.RWMutex
}

func (c *Blobabase) Set(key, blob string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.blobs[key] = blob
	return nil
}

func (c *Blobabase) Get(key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	blob, ok := c.blobs[key]
	if !ok {
		return "", ErrorNoSuchKey
	}
	return blob, nil
}

type server struct {
	store *Blobabase
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /set", s.handleSet)
	mux.HandleFunc("GET /get", s.handleGet)
	return logging(mux)
}

func (s *server) handleSet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	key := q.Get("key")
	value := q.Get("value")
	if key == "" || value == "" {
		http.Error(w, "key and value are required", http.StatusBadRequest)
		return
	}
	if err := s.store.Set(key, value); err != nil {
		slog.Error("cannot write key to store", "key", key, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key too short", http.StatusBadRequest)
		return
	}
	blob, err := s.store.Get(key)
	if errors.Is(err, ErrorNoSuchKey) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("cannot read key from store", "key", key, "err", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := w.Write([]byte(blob)); err != nil {
		slog.Error("write response", "err", err, "key", key)
	}
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slog.Info("blob in", "method", r.Method, "path", r.URL.Path, "from", r.RemoteAddr)
		defer func() {
			slog.Info("blob out", "method", r.Method, "path", r.URL.Path, "took", time.Since(start).String())
		}()
		next.ServeHTTP(w, r)
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

	store := &Blobabase{blobs: make(map[string]string)}
	srv := &server{store: store}

	s := http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      srv.routes(),
	}

	slog.Info("listening", "addr", s.Addr)
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}
