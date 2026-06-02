// Blobabase is a dead-simple, thread-safe blobabase.

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
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

func (c *Blobabase) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return fmt.Sprintf("The count is %d", c.Blobs)
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
			log.Printf("Couldn't Set %s to %s: %v", vals, q, err)
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
	showHelp := flag.Bool("help", false, "print help and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	port := flag.Int("port", 4000, "set port on which to listen")

	flag.Parse()

	if *showHelp {
		fmt.Printf("Usage: blobabase [-port <PORT NUMBER>]")
	}

	if *showVersion {
		fmt.Printf("blobabase %s (commit %s, build %s)\n", version, commit, date)
		return
	}

	if *port <= 1023 {
		fmt.Println("Not so fast, bucko.")
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
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
