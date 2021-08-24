package main

import "testing"

func TestUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "run once"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Usage()
		})
	}
}
