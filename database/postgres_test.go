package database

import (
	"fmt"
	"net"
	"strings"
	"testing"
)

func TestNewPostgresConnectionFailure(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	port := listener.Addr().(*net.TCPAddr).Port
	db, err := NewPostgres(PostgresConfig{
		DSN: fmt.Sprintf("host=127.0.0.1 port=%d user=test dbname=test sslmode=disable connect_timeout=1", port),
	})
	if db != nil || err == nil || !strings.HasPrefix(err.Error(), "connect postgres:") {
		t.Fatalf("NewPostgres() = %v, %v; want nil DB and connect postgres error", db, err)
	}
}
