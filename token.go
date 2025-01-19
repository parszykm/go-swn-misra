package main

import "sync"

type Token struct {
	mu    sync.Mutex
	value int64
}

func (token *Token) GetValue() int64 {
	token.mu.Lock()
	defer token.mu.Unlock()
	return token.value
}

func (token *Token) SetValue(v int64) {
	token.mu.Lock()
	defer token.mu.Unlock()
	token.value = v
}
