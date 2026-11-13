package greet_test

import (
	"net"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	greet "github.com/jotadrilo/ci/testdata/go"
)

func TestHello(t *testing.T) {
	if got, want := greet.Hello("CI"), "Hello, CI!"; got != want {
		t.Fatalf("Hello() = %q, want %q", got, want)
	}
}

// TestConcurrentHello gives -race something to look at.
func TestConcurrentHello(t *testing.T) {
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = greet.Hello("race")
		}()
	}
	wg.Wait()
}

// TestPostgresReachable checks the service container when the workflow
// exports DATABASE_URL, and is skipped otherwise.
func TestPostgresReachable(t *testing.T) {
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		t.Skip("DATABASE_URL not set")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	conn, err := net.DialTimeout("tcp", u.Host, 5*time.Second)
	if err != nil {
		t.Fatalf("dial %s: %v", u.Host, err)
	}
	_ = conn.Close()
}
