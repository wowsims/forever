package core

import (
	"cmp"
	"hash/fnv"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/stats"
	googleproto "google.golang.org/protobuf/proto"
)

func DurationFromSeconds(numSeconds float64) time.Duration {
	return time.Duration(float64(time.Second) * numSeconds)
}

// A float is truncated to whole milliseconds, which is how the game files' millisecond columns read.
func DurationFromMillis[T int32 | float64](ms T) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

func StringFromStatTypes(statTypes []stats.Stat) string {
	statNames := MapSlice(statTypes, func(statType stats.Stat) string {
		return statType.StatName()
	})

	return strings.Join(statNames, ", ")
}

func StringFromActionIDs(actionIDs []ActionID) string {
	names := MapSlice(actionIDs, func(actionID ActionID) string {
		return actionID.String()
	})

	return strings.Join(names, ", ")
}
func (unit *Unit) ExecuteResourceGain(sim *Simulation, resource proto.ResourceType, amount float64, metrics *ResourceMetrics) {
	switch {
	case resource == proto.ResourceType_ResourceTypeMana && amount > 0:
		unit.AddMana(sim, amount, metrics)
	case resource == proto.ResourceType_ResourceTypeMana && amount < 0:
		unit.SpendMana(sim, -amount, metrics)
	case resource == proto.ResourceType_ResourceTypeHealth && amount > 0:
		unit.GainHealth(sim, amount, metrics)
	case resource == proto.ResourceType_ResourceTypeHealth && amount < 0:
		unit.RemoveHealth(sim, -amount)
	case resource == proto.ResourceType_ResourceTypeRage && amount < 0:
		unit.SpendRage(sim, -amount/10, metrics)
	case resource == proto.ResourceType_ResourceTypeRage && amount > 0:
		unit.AddRage(sim, amount/10, metrics)
	case resource == proto.ResourceType_ResourceTypeEnergy && amount < 0:
		unit.SpendEnergy(sim, -amount, metrics)
	case resource == proto.ResourceType_ResourceTypeEnergy && amount > 0:
		unit.AddEnergy(sim, amount, metrics)
	default:
		panic("Unsupported Resource Type in ExecuteResourceGain")
	}
}

func GetTristateValueInt32(effect proto.TristateEffect, regularValue int32, impValue int32) int32 {
	if effect == proto.TristateEffect_TristateEffectRegular {
		return regularValue
	} else if effect == proto.TristateEffect_TristateEffectImproved {
		return impValue
	} else {
		return 0
	}
}

func MakeTristateValue(hasRegular bool, hasImproved bool) proto.TristateEffect {
	if !hasRegular {
		return proto.TristateEffect_TristateEffectMissing
	} else if !hasImproved {
		return proto.TristateEffect_TristateEffectRegular
	} else {
		return proto.TristateEffect_TristateEffectImproved
	}
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

func Ternary[T any](condition bool, val1 T, val2 T) T {
	if condition {
		return val1
	} else {
		return val2
	}
}

func TernaryInt(condition bool, val1 int, val2 int) int {
	if condition {
		return val1
	} else {
		return val2
	}
}

func TernaryInt32(condition bool, val1 int32, val2 int32) int32 {
	if condition {
		return val1
	} else {
		return val2
	}
}

func TernaryInt64(condition bool, val1 int64, val2 int64) int64 {
	if condition {
		return val1
	} else {
		return val2
	}
}

func TernaryFloat64(condition bool, val1 float64, val2 float64) float64 {
	if condition {
		return val1
	} else {
		return val2
	}
}

func TernaryDuration(condition bool, val1 time.Duration, val2 time.Duration) time.Duration {
	if condition {
		return val1
	} else {
		return val2
	}
}

func UnitLevelFloat64(levelDiff int32, plus0Val float64, plus1Val float64, plus2Val float64, plus3Val float64) float64 {
	if levelDiff <= 0 {
		return plus0Val
	} else if levelDiff == 1 {
		return plus1Val
	} else if levelDiff == 2 {
		return plus2Val
	} else {
		return plus3Val
	}
}

func WithinToleranceFloat64(expectedValue float64, actualValue float64, tolerance float64) bool {
	return actualValue >= (expectedValue-tolerance) && actualValue <= (expectedValue+tolerance)
}

// Returns a new slice by applying f to each element in src.
func MapSlice[I any, O any](src []I, f func(I) O) []O {
	dst := make([]O, len(src))
	for i, e := range src {
		dst[i] = f(e)
	}
	return dst
}

// Returns a new map by applying f to each key/value pair in src.
func MapMap[KI comparable, VI any, KO comparable, VO any](src map[KI]VI, f func(KI, VI) (KO, VO)) map[KO]VO {
	dst := make(map[KO]VO, len(src))
	for ki, vi := range src {
		ko, vo := f(ki, vi)
		dst[ko] = vo
	}
	return dst
}

// Returns a new slice containing only the elements for which f returns true.
func FilterSlice[T any](src []T, f func(T) bool) []T {
	dst := make([]T, 0, len(src))
	for _, e := range src {
		if f(e) {
			dst = append(dst, e)
		}
	}
	return dst
}

// Iterate over a slice and calls f
func Each[T any](src []T, f func(T)) {
	for _, e := range src {
		f(e)
	}
}

// Returns a new map containing only the key/value pairs for which f returns true.
func FilterMap[K comparable, V any](src map[K]V, f func(K, V) bool) map[K]V {
	dst := make(map[K]V, len(src))
	for k, v := range src {
		if f(k, v) {
			dst[k] = v
		}
	}
	return dst
}

// Flattens a 2D slice into a 1D slice.
func Flatten[T any](src [][]T) []T {
	var n int
	for _, sublist := range src {
		n += len(sublist)
	}
	dst := make([]T, 0, n)
	for _, sublist := range src {
		dst = append(dst, sublist...)
	}
	return dst
}

// Returns true if two slices share any common elements, and false otherwise.
func CheckSliceOverlap[T comparable](s1 []T, s2 []T) bool {
	for _, v1 := range s1 {
		for _, v2 := range s2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

// Allows Go maps to be used like the "Set" type commonly found in other languages
func SetToSortedSlice[K cmp.Ordered](src map[K]bool) []K {
	dst := make([]K, 0, len(src))
	for k, exists := range src {
		if exists {
			dst = append(dst, k)
		}
	}
	slices.Sort(dst)
	return dst
}

func Clamp(val float64, min float64, max float64) float64 {
	return math.Max(min, math.Min(val, max))
}

func (char *Character) CalcAndRollDamageRange(sim *Simulation, min float64, max float64) float64 {
	return sim.Roll(min, max)
}

type aggregator struct {
	n     int
	sum   float64
	sumSq float64
}

func (x *aggregator) add(v float64) {
	x.n++
	x.sum += v
	x.sumSq += v * v
}

func (x *aggregator) scale(f float64) {
	x.sum *= f
	x.sumSq *= f * f
}

func (x *aggregator) merge(y *aggregator) *aggregator {
	return &aggregator{n: x.n + y.n, sum: x.sum + y.sum, sumSq: x.sumSq + y.sumSq}
}

func (x *aggregator) meanAndStdDev() (float64, float64) {
	mean := x.sum / float64(x.n)
	stdDev := math.Sqrt(x.sumSq/float64(x.n) - mean*mean)
	return mean, stdDev
}

// The zero value is empty. Mean and m2 (the sum of squared deviations) are kept Welford-style, so the
// stdev stays stable for large counts of near-equal values and is exactly 0 for constant ones.
type valueRange struct {
	n        int
	mean, m2 float64
	min, max float64
}

func (r *valueRange) add(v float64) {
	if r.n == 0 {
		r.min, r.max = v, v
	} else {
		r.min = math.Min(r.min, v)
		r.max = math.Max(r.max, v)
	}
	r.n++
	d := v - r.mean
	r.mean += d / float64(r.n)
	r.m2 += d * (v - r.mean)
}

func (r *valueRange) merge(o *valueRange) {
	if o.n == 0 {
		return
	}
	if r.n == 0 {
		*r = *o
		return
	}
	n := float64(r.n + o.n)
	delta := o.mean - r.mean
	r.m2 += o.m2 + delta*delta*float64(r.n)*float64(o.n)/n
	r.mean += delta * float64(o.n) / n
	r.n += o.n
	r.min = math.Min(r.min, o.min)
	r.max = math.Max(r.max, o.max)
}

func (r *valueRange) toProto() *proto.DistributionMetrics {
	stdev := 0.0
	if r.n > 0 {
		stdev = math.Sqrt(math.Max(0, r.m2/float64(r.n)))
	}
	return &proto.DistributionMetrics{
		Avg:            r.mean,
		Stdev:          stdev,
		Min:            r.min,
		Max:            r.max,
		AggregatorData: &proto.AggregatorData{N: int32(r.n)},
	}
}

func GetCurrentProtoVersion() int32 {
	versionMessage := &proto.ProtoVersion{}
	options := versionMessage.ProtoReflect().Descriptor().Options()
	optionValue := googleproto.GetExtension(options, proto.E_CurrentVersionNumber)
	return optionValue.(int32)
}

// A percentage as a multiplier: 20 is 1.2 and -6 is 0.94.
func MultiplierFromPercent(percent float64) float64 {
	return 1 + percent/100
}

// A percentage as a multiplier whose negative divides by the factor its positive multiplies by: 20 is
// 1.2 and -20 is 1/1.2.
func SymmetricMultiplierFromPercent(percent float64) float64 {
	if percent < 0 {
		return 1 / MultiplierFromPercent(-percent)
	}
	return MultiplierFromPercent(percent)
}
