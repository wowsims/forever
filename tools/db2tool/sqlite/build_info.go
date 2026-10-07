package sqlite

import (
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// Records the game build the tables were extracted from, and the highest hotfix push applied to them
// (0 for none), as the one row of BuildInfo. The spell store generator stamps each spell data group
// it takes from the database with both.
func WriteBuildInfo(conn *sqlite.Conn, version, product string, hotfixPush int32) (err error) {
	endFn, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return err
	}
	defer endFn(&err)

	if err := sqlitex.ExecuteTransient(conn, `CREATE TABLE BuildInfo (Version TEXT NOT NULL, Product TEXT NOT NULL, HotfixPush INTEGER NOT NULL)`, nil); err != nil {
		return err
	}
	return sqlitex.Execute(conn, `INSERT INTO BuildInfo (Version, Product, HotfixPush) VALUES (?, ?, ?)`,
		&sqlitex.ExecOptions{Args: []any{version, product, hotfixPush}})
}
