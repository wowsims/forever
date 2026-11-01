// Version selection for .dbd definitions.
// Derived from DBDefsLib types (https://github.com/wowdev/WoWDBDefs).
// Copyright 2022 WoWDBDefs Contributors. BSD-3-Clause — see tools/db2tool/NOTICES.md.

package dbd

import "slices"

import "fmt"

// SelectVersion returns the LAST versionDefinition (in file order) whose
// Builds list contains an entry with trailing build number == buildNumber.
//
// Exact equality on the trailing build number only — buildRanges and
// layoutHashes are deliberately NOT consulted: WoWDBDefs lists the live
// builds explicitly for every configured table, and failing loud here beats
// silently decoding with a near-miss layout.
func SelectVersion(def DBDefinition, buildNumber uint32) (VersionDefinitions, error) {
	for _, v := range slices.Backward(def.VersionDefinitions) {
		for _, b := range v.Builds {
			if b.Build == buildNumber {
				return v, nil
			}
		}
	}
	return VersionDefinitions{}, fmt.Errorf(
		"build %d not found in the .dbd definition — WoWDBDefs may not contain this build yet; wait for upstream or refresh DBDCache", buildNumber)
}
