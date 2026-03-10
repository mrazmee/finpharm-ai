package repository

import "database/sql"

func isNoRows(err error) bool {
	return err == sql.ErrNoRows
}