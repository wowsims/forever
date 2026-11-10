//go:build with_db

package siml

import _ "embed"

//go:embed names.json
var namesJSON []byte

var committedNames lazy[*Names]

func defaultNames() (*Names, error) {
	return committedNames.get(func() (*Names, error) { return ParseNames(namesJSON) })
}
