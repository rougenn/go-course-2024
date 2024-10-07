package storage

import (
	"strconv"
)

type Val struct {
	kind string
	s    string
	n    int
}

type Storage struct {
	inner map[string]*Val
}

func NewStorage() *Storage {
	return &Storage{
		inner: make(map[string]*Val),
	}
}

func (r Storage) Set(key, input_val string) {
	int_val, err := strconv.Atoi(input_val)
	if err == nil {
		r.inner[key] = &Val{
			kind: "int",
			n:    int_val,
		}
		return
	}
	r.inner[key] = &Val{
		kind: "string",
		s:    input_val,
	}
}

func (r Storage) Get(key string) *string {
	val, ok := r.inner[key]
	if !ok {
		return nil
	}
	if val.kind == "string" {
		return &val.s
	}
	output := strconv.Itoa(val.n)
	return &output
}

func (r Storage) GetKind(key string) *string {
	val, ok := r.inner[key]
	if !ok {
		return nil
	}
	return &val.kind
}
