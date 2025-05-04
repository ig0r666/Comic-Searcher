package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustMakeLogger(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{"DEBUG test", "DEBUG", false},
		{"INFO test", "INFO", false},
		{"ERROR test", "ERROR", false},
		{"wrong test", "wrong", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				assert.Panics(t, func() {
					mustMakeLogger(tt.logLevel)
				})
			} else {
				logger := mustMakeLogger(tt.logLevel)
				assert.NotNil(t, logger)
			}
		})
	}
}
