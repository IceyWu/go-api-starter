package database

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
)

func TestValidateMySQLConfig(t *testing.T) {
	base := &Config{
		Driver:   "mysql",
		Host:     "localhost",
		Port:     3306,
		Username: "root",
		DBName:   "go_api_starter",
		Charset:  "utf8mb4",
	}
	if err := validateMySQLConfig(base); err != nil {
		t.Fatalf("expected valid MySQL configuration, got %v", err)
	}

	for _, name := range []string{"", "has-dash", "name;DROP DATABASE mysql", strings.Repeat("a", 65), "mysql"} {
		t.Run(name, func(t *testing.T) {
			cfg := *base
			cfg.DBName = name
			if err := validateMySQLConfig(&cfg); err == nil {
				t.Fatalf("expected database name %q to be rejected", name)
			}
		})
	}
}

func TestMySQLDSNIncludesBoundedTimeouts(t *testing.T) {
	cfg := &Config{Host: "localhost", Port: 3306, Username: "test", DBName: "test"}
	parsed, err := mysqldriver.ParseDSN(mysqlDSN(cfg, cfg.DBName))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Timeout != 10*time.Second || parsed.ReadTimeout != 30*time.Second || parsed.WriteTimeout != 30*time.Second {
		t.Fatal("missing default socket timeouts")
	}
	cfg.ConnectTimeout, cfg.ReadTimeout, cfg.WriteTimeout = time.Second, 2*time.Second, 3*time.Second
	parsed, err = mysqldriver.ParseDSN(mysqlDSN(cfg, cfg.DBName))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Timeout != time.Second || parsed.ReadTimeout != 2*time.Second || parsed.WriteTimeout != 3*time.Second {
		t.Fatal("custom socket timeouts were ignored")
	}
}

func TestMySQLInitializationTimesOutOnStalledHandshake(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	done := make(chan struct{})
	t.Cleanup(func() { close(done) })
	go func() {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = connection.Close() }()
		<-done
	}()
	host, portText, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	err = EnsureMySQLDatabase(&Config{Host: host, Port: port, Username: "test", DBName: "test", ConnectTimeout: 50 * time.Millisecond, ReadTimeout: 50 * time.Millisecond})
	if err == nil || time.Since(started) > 2*time.Second {
		t.Fatalf("stalled handshake was not bounded: %v", err)
	}
}

func TestInitRejectsInvalidMySQLDatabaseNameBeforeConnecting(t *testing.T) {
	_, err := Init(&Config{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		DBName:   "bad-name",
		Charset:  "utf8mb4",
	})
	if err == nil || !strings.Contains(err.Error(), "invalid MySQL database name") {
		t.Fatalf("expected database name validation error before connection, got %v", err)
	}
}
