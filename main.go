// Before your interview, write a program that runs a server that is accessible on `http://localhost:4000/`. When your server receives a request on `http://localhost:4000/set?somekey=somevalue` it should store the passed key and value in memory. When it receives a request on `http://localhost:4000/get?key=somekey` it should return the value stored at `somekey`.

// During your interview, you'll pair on improving your server. For example, you might decide to save the data to a file; you could start with simply appending each write to the file, and work on making it more efficient if you have time.

package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type Blob struct {
	bytes []byte
}

type Blobabase struct {
	Blobs map[string][]byte
	mu    sync.Mutex
}

func (c *Blobabase) Set(name string, blob []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Blobs[name] = blob
	return nil
}

func (c *Blobabase) Get(name string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	blob, ok := c.Blobs[name]
	// TODO: remember how to figure out if the key exists or not
	return blob, nil
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
	mux.HandleFunc("PUT /set/{key}?{value}", func(w http.ResponseWriter, r *http.Request) {
		count := c.Set(r.PathValue("key"), []byte(r.PathValue("value")))
		if _, err := fmt.Fprintf(w, "The count is %d", count); err != nil {
			log.Printf("write failed: %v", err)
		}
	})
	mux.HandleFunc("PUT /reset", func(w http.ResponseWriter, r *http.Request) {
		count := c.Reset()
		if _, err := fmt.Fprintf(w, "The count is %d", count); err != nil {
			log.Printf("write failed: %v", err)
		}
	})
	mux.HandleFunc("GET /get?{key}", func(w http.ResponseWriter, r *http.Request) {
		value, err := c.Get(r.PathValue("key"))
		if err != nil || value == nil {
			log.Printf("Couldn't retrieve key %s: %w", r.PathValue("key"), err)
			// return not found http error
		}
		fmt.Fprint(w, value)
	})
	return mux
}

func main() {
	var blobabase Blobabase

	mux := newMux(&blobabase)

	s := http.Server{
		Addr:         ":4000",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	}
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
}
