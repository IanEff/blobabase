package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newStore() *Blobabase {
	return &Blobabase{Blobs: make(map[string]string)}
}

func TestSetThenGetReturnsTheStoredBlob(t *testing.T) {
	// place
	b := newStore()

	// act
	if err := b.Set("hello", "world"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := b.Get("hello")

	// assert
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "world" {
		t.Fatalf("got %q, want %q", got, "world")
	}
}

func TestGetMissingKeyReturnsNoSuchKey(t *testing.T) {
	// place
	b := newStore()

	// act
	_, err := b.Get("nada")

	// assert
	if !errors.Is(err, ErrorNoSuchKey) {
		t.Fatalf("got %v, want %v", err, ErrorNoSuchKey)
	}
}

func TestHandleGetReturnsBlobForKnownKey(t *testing.T) {
	// place
	s := &server{store: newStore()}
	s.store.Set("hello", "world")
	req := httptest.NewRequest("GET", "/get?key=hello", nil)
	w := httptest.NewRecorder()

	// act
	s.handleGet(w, req)

	// assert
	if w.Code != http.StatusOK {
		t.Fatalf("code: got %d, want %d", w.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(w.Body.String()); got != "world" {
		t.Fatalf("body: got %q, want %q", got, "world")
	}
}

func TestHandleGetReturns404ForUnknownKey(t *testing.T) {
	// place
	s := &server{store: newStore()}
	req := httptest.NewRequest("GET", "/get?key=nada", nil)
	w := httptest.NewRecorder()

	// act
	s.handleGet(w, req)

	// assert
	if w.Code != http.StatusNotFound {
		t.Fatalf("code: got %d, want %d", w.Code, http.StatusNotFound)
	}
}
