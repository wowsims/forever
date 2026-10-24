//go:build !with_db

package siml

import "errors"

// The table rides on the with_db tag with the database it is generated from, so lib.wasm carries neither.
func defaultNames() (*Names, error) {
	return nil, errors.New("the name table sim/siml/names.json is built in only with the with_db build tag")
}
