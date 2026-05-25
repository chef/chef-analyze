package cmd

import (
	"math"
	"testing"
)

func TestSessionDurationSeconds_Valid(t *testing.T) {
	got, err := sessionDurationSeconds(60)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != 3600 {
		t.Fatalf("expected 3600, got %d", got)
	}
}

func TestSessionDurationSeconds_ZeroRejected(t *testing.T) {
	_, err := sessionDurationSeconds(0)
	if err == nil {
		t.Fatal("expected error for zero duration")
	}
}

func TestSessionDurationSeconds_NegativeRejected(t *testing.T) {
	_, err := sessionDurationSeconds(-1)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestSessionDurationSeconds_OverflowRejected(t *testing.T) {
	_, err := sessionDurationSeconds(math.MaxInt32/60 + 1)
	if err == nil {
		t.Fatal("expected overflow error")
	}
}

func TestSessionDurationSeconds_MaxBoundaryAccepted(t *testing.T) {
	minutes := int64(math.MaxInt32 / 60)
	got, err := sessionDurationSeconds(minutes)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got <= 0 {
		t.Fatalf("expected positive duration seconds, got %d", got)
	}
}
