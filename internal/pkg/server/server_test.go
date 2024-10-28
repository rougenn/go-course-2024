package server

import (
	"encoding/json"
	"hw1/internal/pkg/storage"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSeverHealth(t *testing.T) {
	store, err := storage.NewStorage(time.Minute*20, time.Minute*60, "my-storage.json")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	s := New("localhost:8090", store)
	s.newAPI().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGet(t *testing.T) {
	store, err := storage.NewStorage(time.Minute*20, time.Minute*60, "my-storage.json")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/scalar/get/asdf", nil)

	s := New("localhost:8090", store)
	s.newAPI().ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

type Req struct {
	Value string `json:"value"`
}

func TestSetGet(t *testing.T) {
	store, err := storage.NewStorage(time.Minute*20, time.Minute*60, "my-storage.json")
	if err != nil {
		log.Fatalf("Failed to create storage: %v", err)
	}

	w := httptest.NewRecorder()

	v := Req{
		Value: "2345",
	}
	jsonreq, err := json.Marshal(v)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, "/scalar/set/asdf", strings.NewReader(string(jsonreq)))
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	s := New("localhost:8090", store)
	s.newAPI().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	//  get req
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/scalar/get/asdf", nil)

	s.newAPI().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result Req
	err = json.NewDecoder(w.Body).Decode(&result)
	if err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}
	assert.Equal(t, v.Value, result.Value)
}
