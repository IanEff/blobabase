// Blobabase is a dead-simple, thread-safe blobabase.

package main

import (
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
	mu    sync.Mutex
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
	if err := delete(c.Blobs, key); err != nil {
		log.Printf("coudn't delete key %s: %v", key, err)
		return err
	}
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

func newMux(c *Blobabase) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("PUT /set?{key}={value}", func(w http.ResponseWriter, r *http.Request) {
		if err := c.Set(r.PathValue("key"), []byte(r.PathValue("value"))); err != nil {
			log.Printf("write failed: %v", err)
			http.Error(w, err.Error(), http.StatusFailedDependency)
		}
		return
	})
	mux.HandleFunc("PUT /reset", func(w http.ResponseWriter, r *http.Request) {
		count := c.Reset()
		if _, err := fmt.Fprintf(w, "The blobabase has been reset."); err != nil {
			log.Printf("write failed: %v", err)
			http.Error(w, err.Error(), http.StatusFailedDependency)
		}
		return
	})
	mux.HandleFunc("GET /get?{key}", func(w http.ResponseWriter, r *http.Request) {
		value, err := c.Get(r.PathValue("key"))
		if err != nil || value == nil {
			log.Printf("Couldn't retrieve key %s: %v", r.PathValue("key"), err)
			if errors.Is(err, ErrorNoSuchKey) {
				http.Error(w, err.Error(), http.StatusNotFound)
			}
		}
		if err := fmt.Fprint(w, value); err != nil {
			log.Printf("Couldn't write to line: %v", err)
			http.Error(w, err.Error(), http.StatusFailedDependency)
		}
	})
	return mux
}

func main() {
	var blobabase Blobabase

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

	if port <= 1000 {
		fmt.Println("Not so fast, bucko.")
	}



	mux := newMux(&blobabase)

	s := http.Server{
		Addr:         fmt.Sprintf(":%d", port)
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
