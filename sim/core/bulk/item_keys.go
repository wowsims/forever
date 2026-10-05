package bulk

import (
	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
)

type bulkSimCandidateOption struct {
	spec    *proto.ItemSpec
	item    core.Item
	enchant int32 // The pooled enchant this option was expanded with; 0 keeps the carry-over.
	hand    weaponHand
}

// weaponHand limits an expanded dual-wield weapon to the hand whose pool supplied its enchant.
type weaponHand uint8

const (
	eitherHand weaponHand = iota
	mainHandOnly
	offHandOnly
)

// itemSpecCacheKey identifies the base item; copy counting runs on it, so enchant variants of one
// item count as one copy.
type itemSpecCacheKey struct {
	id           int32
	randomSuffix int32
}

type candidateOptionKey struct {
	item    itemSpecCacheKey
	enchant int32
}

func buildCandidateOptionKey(option bulkSimCandidateOption) candidateOptionKey {
	return candidateOptionKey{item: buildItemSpecKey(option.spec), enchant: option.enchant}
}

// itemSpecFingerprintKey is a zero-allocation alternative to the string fingerprint.
type itemSpecFingerprintKey struct {
	id           int32
	randomSuffix int32
	enchant      int32
}

func buildItemSpecFingerprintKey(item *proto.ItemSpec) itemSpecFingerprintKey {
	if item == nil {
		return itemSpecFingerprintKey{}
	}
	return itemSpecFingerprintKey{
		id:           item.GetId(),
		randomSuffix: item.GetRandomSuffix(),
		enchant:      item.GetEnchant(),
	}
}

func dedupeCandidateOptions(options []bulkSimCandidateOption) []bulkSimCandidateOption {
	if len(options) <= 1 {
		return options
	}

	seen := make(map[candidateOptionKey]struct{}, len(options))
	deduped := make([]bulkSimCandidateOption, 0, len(options))
	for _, option := range options {
		key := buildCandidateOptionKey(option)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		deduped = append(deduped, option)
	}

	return deduped
}

func optionsContainEquivalent(options []bulkSimCandidateOption, target bulkSimCandidateOption) bool {
	for _, option := range options {
		if candidateOptionsEqual(option, target) {
			return true
		}
	}
	return false
}

func candidateOptionsEqual(left bulkSimCandidateOption, right bulkSimCandidateOption) bool {
	return buildCandidateOptionKey(left) == buildCandidateOptionKey(right)
}

func candidateOptionsShareItem(left bulkSimCandidateOption, right bulkSimCandidateOption) bool {
	return buildItemSpecKey(left.spec) == buildItemSpecKey(right.spec)
}

func candidateOptionEqualsItem(option bulkSimCandidateOption, item core.Item) bool {
	return buildItemSpecKey(option.spec) == buildItemSpecKey(item.ToItemSpecProto())
}

func candidateOptionEqualsItemPtr(option *bulkSimCandidateOption, item *core.Item) bool {
	if option == nil || item == nil {
		return option == nil && item == nil
	}
	return candidateOptionEqualsItem(*option, *item)
}

func buildItemSpecKey(itemSpec *proto.ItemSpec) itemSpecCacheKey {
	if itemSpec == nil {
		return itemSpecCacheKey{}
	}
	return itemSpecCacheKey{
		id:           itemSpec.GetId(),
		randomSuffix: itemSpec.GetRandomSuffix(),
	}
}
