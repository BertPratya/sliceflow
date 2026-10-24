package slice

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BertPratya/sliceflow/internal/eventlog"
)

func lockPath(r Repo, name string) string {
	return filepath.Join(filepath.FromSlash(r.Root), ".sliceflow", "locks", name+".lock")
}

func writeLock(t *testing.T, r Repo, name, content string) string {
	t.Helper()
	p := lockPath(r, name)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLockFile(t *testing.T) {
	r := newRepo(t)
	free, err := takeLock(context.Background(), r, "login-google")
	if err != nil {
		t.Fatal(err)
	}
	p := lockPath(r, "login-google")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var got struct {
		Pid   int    `json:"pid"`
		Since string `json:"since"`
	}
	if b[len(b)-1] != '\n' {
		t.Errorf("lock not one line: %q", b)
	}
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.Pid != os.Getpid() {
		t.Errorf("pid = %d", got.Pid)
	}
	if got.Since != fixedNow.Format(eventlog.TimeLayout) {
		t.Errorf("since = %q", got.Since)
	}
	ig, err := os.ReadFile(filepath.Join(filepath.FromSlash(r.Root), ".sliceflow", ".gitignore"))
	if err != nil || string(ig) != "*\n" {
		t.Errorf("gitignore = %q, %v", ig, err)
	}
	free()
	if _, err := os.Stat(p); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("lock still there: %v", err)
	}
}

func TestLockBusy(t *testing.T) {
	r := newRepo(t)
	r.LockWait = 200 * time.Millisecond
	p := writeLock(t, r, "login-google", `{"pid":4312,"since":"2026-10-24T10:02:00Z"}`+"\n")
	_, err := takeLock(context.Background(), r, "login-google")
	want := "slice login-google is busy: another sliceflow command holds .sliceflow/locks/login-google.lock (pid 4312, since 10:02 UTC). If no sliceflow command is running, delete that file."
	var se *Error
	if !errors.As(err, &se) || se.Message != want {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Errorf("foreign lock was removed: %v", err)
	}
}

func TestLockUnreadable(t *testing.T) {
	r := newRepo(t)
	r.LockWait = 200 * time.Millisecond
	writeLock(t, r, "login-google", "x")
	_, err := takeLock(context.Background(), r, "login-google")
	want := "slice login-google is busy: another sliceflow command holds .sliceflow/locks/login-google.lock (pid unknown, since unknown UTC). If no sliceflow command is running, delete that file."
	var se *Error
	if !errors.As(err, &se) || se.Message != want {
		t.Fatalf("err = %v", err)
	}
}

func TestLockWaits(t *testing.T) {
	r := newRepo(t)
	r.LockWait = time.Second
	p := writeLock(t, r, "login-google", `{"pid":1,"since":"2026-10-24T10:02:00Z"}`+"\n")
	go func() {
		time.Sleep(100 * time.Millisecond)
		os.Remove(p)
	}()
	free, err := takeLock(context.Background(), r, "login-google")
	if err != nil {
		t.Fatal(err)
	}
	free()
}

func TestLockCancelled(t *testing.T) {
	r := newRepo(t)
	r.LockWait = 5 * time.Second
	writeLock(t, r, "login-google", "x")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := takeLock(ctx, r, "login-google"); err == nil {
		t.Fatal("want error")
	}
	if time.Since(start) > 2*time.Second {
		t.Error("did not stop on cancel")
	}
}
