package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func BenchmarkGetHash(b *testing.B) {
	// Setup repo with many tags
	dir, err := os.MkdirTemp("", "bench")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	r, err := git.PlainInit(dir, false)
	if err != nil {
		b.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		b.Fatal(err)
	}

	filename := "hello.txt"
	_ = os.WriteFile(dir+"/"+filename, []byte("hello"), 0644)
	_, _ = w.Add(filename)
	commit, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		b.Fatal(err)
	}

	// Create 100 tags
	for i := 0; i < 100; i++ {
		tagName := fmt.Sprintf("v0.0.%d", i)
		if i%2 == 0 {
			// Lightweight
			_, err = r.CreateTag(tagName, commit, nil)
		} else {
			// Annotated
			_, err = r.CreateTag(tagName, commit, &git.CreateTagOptions{
				Message: "Annotated tag",
				Tagger: &object.Signature{
					Name:  "Test",
					Email: "test@example.com",
					When:  time.Now(),
				},
			})
		}
		if err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We are benchmarking the sequence: Find tag, then GetHash
		// This simulates the logic in main.go
		highest := FindHighestVersionTag(r)
		_, _ = GetHash(r, highest)
	}
}

func BenchmarkGetHashOnly(b *testing.B) {
	// Setup repo with many tags
	dir, err := os.MkdirTemp("", "bench_only")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(dir)

	r, err := git.PlainInit(dir, false)
	if err != nil {
		b.Fatal(err)
	}

	w, err := r.Worktree()
	if err != nil {
		b.Fatal(err)
	}

	filename := "hello.txt"
	_ = os.WriteFile(dir+"/"+filename, []byte("hello"), 0644)
	_, _ = w.Add(filename)
	commit, err := w.Commit("Initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test",
			Email: "test@example.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		b.Fatal(err)
	}

	// Create 100 tags
	for i := 0; i < 100; i++ {
		tagName := fmt.Sprintf("v0.0.%d", i)
		if i%2 == 0 {
			// Lightweight
			_, err = r.CreateTag(tagName, commit, nil)
		} else {
			// Annotated
			_, err = r.CreateTag(tagName, commit, &git.CreateTagOptions{
				Message: "Annotated tag",
				Tagger: &object.Signature{
					Name:  "Test",
					Email: "test@example.com",
					When:  time.Now(),
				},
			})
		}
		if err != nil {
			b.Fatal(err)
		}
	}

    highest := FindHighestVersionTag(r)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GetHash(r, highest)
	}
}
