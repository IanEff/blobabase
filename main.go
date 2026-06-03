// Blobabase is a dead-simple, thread-safe blobabase.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	c.mu.Lock()
	defer c.mu.Unlock()
	blob, ok := c.Blobs[key]
	if !ok {
		return nil, ErrorNoSuchKey
	}
	return blob, nil
}

func (c *Blobabase) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Blobs, key)
	return nil
}

func (c *Blobabase) Reset() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Blobs = make(map[string][]byte)
	return nil
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
		if err := s.store.Set(name, []byte(vals[len(vals)-1])); err != nil {
			slog.Error("set failed", "key", name, "err", err)
		}
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
	return mux
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	showHelp := flag.Bool("help", false, "print help and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	port := flag.Int("port", 4000, "set port on which to listen")

	flag.Parse()

	if *showHelp {
		fmt.Printf("Usage: blobabase [-port <PORT NUMBER>]")
		return
	}

	if *showVersion {
		fmt.Printf("blobabase %s (commit %s, build %s)\n", version, commit, date)
		return
	}

	if *port <= 1023 {
		fmt.Println("Not so fast, bucko.")
		slog.Info("cannot bind to privileged port: %d")
		os.Exit(1)
	}

	store := &Blobabase{Blobs: make(map[string][]byte)}
	srv := &server{store: store}

	s := http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      srv.routes(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("listening", "addr", s.Addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "err", err)
			stop() // unblock main on failure
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received, draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
		os.Exit(1)
	}
	slog.Info("shutdown complete")
}
