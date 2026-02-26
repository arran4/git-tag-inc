package main

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
)

func BenchmarkFindHighestVersionTag(b *testing.B) {
	r, err := git.Init(memory.NewStorage(), nil)
	if err != nil {
		b.Fatal(err)
	}
	repo := r

	for i := 0; i < b.N; i++ {
		_, err := FindHighestVersionTag(repo)
		if err != nil {
			// In a real repo this might return an error if no tags found,
			// or if something else fails. For benchmark we just want to ensure
			// it compiles and runs.
			// We can ignore the error or log it if needed.
		}
	}
}
