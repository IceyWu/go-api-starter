package repository

import (
	"database/sql"
	"errors"
	"time"
)

func noRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
func modelTime() time.Time  { return time.Now() }
