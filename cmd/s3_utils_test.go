package cmd

import (
	"os"
	"path/filepath"
	"math"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sts"
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

func TestSafeUploadObjectKey(t *testing.T) {
	t.Run("normal path", func(t *testing.T) {
		got, err := safeUploadObjectKey("/tmp/reports/out.txt")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "out.txt" {
			t.Fatalf("expected out.txt, got %s", got)
		}
	})

	t.Run("invalid path", func(t *testing.T) {
		_, err := safeUploadObjectKey(string(filepath.Separator))
		if err == nil {
			t.Fatal("expected error for root path")
		}
	})
}

func TestWritePrivateFile(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "token.sh")

	err := writePrivateFile(target, []byte("secret-data"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("unable to read file: %v", err)
	}
	if string(content) != "secret-data" {
		t.Fatalf("unexpected file content: %s", string(content))
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("unable to stat file: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected file perm 0600, got %#o", info.Mode().Perm())
	}
}

func TestAWSCredentialFormatters_NilSafe(t *testing.T) {
	if got := awsCredentialsToUnixVariables(nil); got != "" {
		t.Fatalf("expected empty string for nil token, got %q", got)
	}

	tokenNoCreds := &sts.GetSessionTokenOutput{}
	if got := awsCredentialsToPowershellVariables(tokenNoCreds); got != "" {
		t.Fatalf("expected empty string for missing credentials, got %q", got)
	}
}
