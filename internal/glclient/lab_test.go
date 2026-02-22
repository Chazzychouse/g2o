package glclient

import (
	"testing"
)

func TestNewGitlab(t *testing.T) {
	t.Run("empty token", func(t *testing.T) {
		_, err := NewGitlab("")
		assertError(t, err, ErrTokenRequired)
	})
}

func assertError(t *testing.T, err error, expected error) {
	if err != expected {
		t.Fatalf("expected error %v, got %v", expected, err)
	}
}
