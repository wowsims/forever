package sqlite

import (
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

// WriteBuildInfo records the game build the tables were extracted from as the one row of BuildInfo.
// The spell store generator stamps each spell data group it takes from the database with this version.
func WriteBuildInfo(conn *sqlite.Conn, version, product string) (err error) {
	endFn, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return err
	}
	defer endFn(&err)

	if err := sqlitex.ExecuteTransient(conn, `CREATE TABLE BuildInfo (Version TEXT NOT NULL, Product TEXT NOT NULL)`, nil); err != nil {
		return err
	}
	return sqlitex.Execute(conn, `INSERT INTO BuildInfo (Version, Product) VALUES (?, ?)`,
		&sqlitex.ExecOptions{Args: []any{version, product}})
}
