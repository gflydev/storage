package local

import (
	"io"
	"os"
	"testing"

	"github.com/gflydev/storage"
)

// newTestStorage returns a local storage rooted at a temporary directory so tests
// do not touch the working tree.
func newTestStorage(t *testing.T) *Storage {
	t.Helper()

	return &Storage{BaseDir: t.TempDir()}
}

// TestAutoRegister verifies that importing the package auto-registers the "local"
// backend (via init), so storage.Instance(local.Type) is non-nil without an
// explicit storage.Register call.
func TestAutoRegister(t *testing.T) {
	if got := storage.Instance(Type); got == nil {
		t.Fatalf("expected local storage to be auto-registered, got nil")
	}
}

func TestPutGet(t *testing.T) {
	s := newTestStorage(t)

	if ok := s.Put("hello.txt", "Hello world"); !ok {
		t.Fatalf("Put returned false")
	}

	data, err := s.Get("hello.txt")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if string(data) != "Hello world" {
		t.Fatalf("Get = %q, want %q", data, "Hello world")
	}
}

// TestGetMissingReturnsError ensures Get surfaces an error for a missing file
// rather than silently returning nil.
func TestGetMissingReturnsError(t *testing.T) {
	s := newTestStorage(t)

	if _, err := s.Get("does-not-exist.txt"); err == nil {
		t.Fatalf("expected error for missing file, got nil")
	}
}

// TestGetStream verifies GetStream streams file contents and no longer returns
// NotImplemented.
func TestGetStream(t *testing.T) {
	s := newTestStorage(t)

	if ok := s.PutData("stream.txt", []byte("stream me")); !ok {
		t.Fatalf("PutData returned false")
	}

	rc, err := s.GetStream("stream.txt")
	if err != nil {
		t.Fatalf("GetStream returned error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("reading stream: %v", err)
	}
	if string(data) != "stream me" {
		t.Fatalf("stream = %q, want %q", data, "stream me")
	}
}

func TestAppend(t *testing.T) {
	s := newTestStorage(t)

	if ok := s.Put("log.txt", "line1\n"); !ok {
		t.Fatalf("Put returned false")
	}
	if ok := s.Append("log.txt", "line2\n"); !ok {
		t.Fatalf("Append returned false")
	}

	data, err := s.Get("log.txt")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if string(data) != "line1\nline2\n" {
		t.Fatalf("Get = %q, want %q", data, "line1\nline2\n")
	}
}

func TestCopyMoveDelete(t *testing.T) {
	s := newTestStorage(t)

	if ok := s.Put("src.txt", "payload"); !ok {
		t.Fatalf("Put returned false")
	}
	if ok := s.Copy("src.txt", "copy.txt"); !ok {
		t.Fatalf("Copy returned false")
	}
	if !s.Exists("copy.txt") {
		t.Fatalf("copy.txt should exist after Copy")
	}
	if ok := s.Move("copy.txt", "moved.txt"); !ok {
		t.Fatalf("Move returned false")
	}
	if s.Exists("copy.txt") {
		t.Fatalf("copy.txt should not exist after Move")
	}
	if !s.Exists("moved.txt") {
		t.Fatalf("moved.txt should exist after Move")
	}
	if ok := s.Delete("moved.txt"); !ok {
		t.Fatalf("Delete returned false")
	}
	if s.Exists("moved.txt") {
		t.Fatalf("moved.txt should not exist after Delete")
	}
}

func TestSizeAndLastModified(t *testing.T) {
	s := newTestStorage(t)

	if ok := s.Put("size.txt", "12345"); !ok {
		t.Fatalf("Put returned false")
	}
	if got := s.Size("size.txt"); got != 5 {
		t.Fatalf("Size = %d, want 5", got)
	}
	if got := s.LastModified("size.txt"); got.IsZero() {
		t.Fatalf("LastModified should be non-zero for an existing file")
	}
	// Missing file returns zero time.
	if got := s.LastModified("missing.txt"); !got.IsZero() {
		t.Fatalf("LastModified for missing file = %v, want zero", got)
	}
}

func TestMakeDir(t *testing.T) {
	s := newTestStorage(t)

	if ok := s.MakeDir("foo/bar"); !ok {
		t.Fatalf("MakeDir returned false")
	}
	info, err := os.Stat(s.Path("foo/bar"))
	if err != nil {
		t.Fatalf("stat created dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected foo/bar to be a directory")
	}
}
