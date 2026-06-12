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

func (b *Blobabase) Set(key, blob string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blobs[key] = blob
	return nil
}

func (b *Blobabase) Get(key string) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	blob, ok := b.blobs[key]
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
	mux.HandleFunc("/set", s.handleSet)
	mux.HandleFunc("GET /get", s.handleGet)
	return logging(mux)
}

// handleSet stores each query pair as a key/value: /set?somekey=somevalue.
func (s *server) handleSet(w http.ResponseWriter, r *http.Request) {
	pairs := r.URL.Query()
	if len(pairs) == 0 {
		http.Error(w, "a key=value query parameter is required", http.StatusBadRequest)
		return
	}
	for key, values := range pairs {
		if key == "" {
			http.Error(w, "key must not be empty", http.StatusBadRequest)
			return
		}
		// for a repeated key, last value wins
		if err := s.store.Set(key, values[len(values)-1]); err != nil {
			slog.Error("cannot write key to store", "key", key, "err", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	// blind write-- so 204 "success, no body"
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
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
