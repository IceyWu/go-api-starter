//go:build integration

package database

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestMySQLConnectionAndSchema(t *testing.T) {
	dsn := os.Getenv("MYSQL_TEST_DSN")
	if dsn == "" {
		t.Skip("MYSQL_TEST_DSN is not configured")
	}

	db, err := sqlx.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping())
	for _, table := range []string{"users", "files", "transcoding_tasks"} {
		var count int
		require.NoError(t, db.Get(&count, "SELECT COUNT(*) FROM "+table))
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS integration_probe (id BIGINT PRIMARY KEY AUTO_INCREMENT, value VARCHAR(64) NOT NULL)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO integration_probe (value) VALUES (?)", "ok")
	require.NoError(t, err)

	var value string
	require.NoError(t, db.Get(&value, "SELECT value FROM integration_probe ORDER BY id DESC LIMIT 1"))
	require.Equal(t, "ok", value)
	_, err = db.Exec("DROP TABLE integration_probe")
	require.NoError(t, err)
}
