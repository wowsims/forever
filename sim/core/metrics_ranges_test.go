package core

import (
	"math"
	"testing"

	"github.com/wowsims/forever/sim/core/proto"
)

func rangeOf(values ...float64) *valueRange {
	r := &valueRange{}
	for _, v := range values {
		r.add(v)
	}
	return r
}

func TestValueRangeAddTracksMinMaxAndAggregate(t *testing.T) {
	r := rangeOf(4, 2, 9, 5)
	if r.n != 4 || r.mean != 5 || math.Abs(r.m2-(1+9+16+0)) > 1e-12 {
		t.Fatalf("aggregate wrong: n=%d mean=%v m2=%v", r.n, r.mean, r.m2)
	}
	if r.min != 2 || r.max != 9 {
		t.Fatalf("min/max wrong: %v/%v", r.min, r.max)
	}
}

func TestValueRangeFirstValueSetsMinAndMax(t *testing.T) {
	r := rangeOf(-3)
	if r.min != -3 || r.max != -3 {
		t.Fatalf("min/max after one negative value: %v/%v", r.min, r.max)
	}
}

func TestValueRangeMerge(t *testing.T) {
	a := rangeOf(3, 7)
	b := rangeOf(1, 5)

	merged := &valueRange{}
	merged.merge(a)
	if *merged != *a {
		t.Fatalf("empty receiver should copy: %+v vs %+v", *merged, *a)
	}
	merged.merge(&valueRange{})
	if *merged != *a {
		t.Fatalf("empty argument should be a no-op: %+v", *merged)
	}
	merged.merge(b)
	if merged.n != 4 || merged.mean != 4 || merged.min != 1 || merged.max != 7 {
		t.Fatalf("merge wrong: %+v", *merged)
	}
}

func TestValueRangeMergedHalvesEqualOneRange(t *testing.T) {
	var values []float64
	for i := range 1000 {
		values = append(values, 1000+math.Sin(float64(i))*3+float64(i%7))
	}
	whole := rangeOf(values...)
	merged := rangeOf(values[:377]...)
	merged.merge(rangeOf(values[377:]...))

	w, m := whole.toProto(), merged.toProto()
	if math.Abs(w.Avg-m.Avg) > 1e-12*math.Abs(w.Avg) || math.Abs(w.Stdev-m.Stdev) > 1e-12*w.Stdev {
		t.Fatalf("merged %+v, whole %+v", m, w)
	}
	if m.AggregatorData.N != 1000 || m.Min != w.Min || m.Max != w.Max {
		t.Fatalf("merged %+v, whole %+v", m, w)
	}
}

func TestValueRangeToProto(t *testing.T) {
	m := rangeOf(2, 4, 4, 4, 5, 5, 7, 9).toProto()
	if m.Avg != 5 || math.Abs(m.Stdev-2) > 1e-9 || m.Min != 2 || m.Max != 9 {
		t.Fatalf("unexpected summary: %+v", m)
	}
	if m.AggregatorData.N != 8 {
		t.Fatalf("unexpected aggregator data: %+v", m.AggregatorData)
	}
}

func TestValueRangeConstantValuesHaveExactlyZeroStdev(t *testing.T) {
	const v = 1234.5678
	added := rangeOf(v, v, v, v, v, v, v)
	if s := added.toProto().Stdev; s != 0 {
		t.Fatalf("added stdev %v", s)
	}
	merged := rangeOf(v, v, v)
	merged.merge(rangeOf(v, v, v, v))
	if s := merged.toProto().Stdev; s != 0 {
		t.Fatalf("merged stdev %v", s)
	}
	if s := combineHitRanges(rangeOf(v, v).toProto(), rangeOf(v, v, v).toProto()).Stdev; s != 0 {
		t.Fatalf("combined stdev %v", s)
	}
}

func TestCombineHitRangesWeightsByCount(t *testing.T) {
	base := rangeOf(10, 20).toProto()
	add := rangeOf(1, 2, 3, 40).toProto()
	got := combineHitRanges(base, add)

	want := rangeOf(10, 20, 1, 2, 3, 40).toProto()
	if got.AggregatorData.N != 6 || math.Abs(got.Avg-want.Avg) > 1e-9 || math.Abs(got.Stdev-want.Stdev) > 1e-9 {
		t.Fatalf("combined %+v, want %+v", got, want)
	}
	if got.Min != 1 || got.Max != 40 {
		t.Fatalf("min/max %v/%v", got.Min, got.Max)
	}
	if base.AggregatorData.N != 2 {
		t.Fatalf("base must stay untouched")
	}
}

func TestCombineHitRangesSkipsEmptySides(t *testing.T) {
	r := rangeOf(5, 9).toProto()
	for name, got := range map[string]*proto.DistributionMetrics{
		"nil base":    combineHitRanges(nil, r),
		"nil add":     combineHitRanges(r, nil),
		"empty base":  combineHitRanges(&proto.DistributionMetrics{}, r),
		"empty add":   combineHitRanges(r, &proto.DistributionMetrics{AggregatorData: &proto.AggregatorData{}}),
		"zero-N base": combineHitRanges(&proto.DistributionMetrics{Min: 0, Max: 0, AggregatorData: &proto.AggregatorData{}}, r),
	} {
		if got.Min != 5 || got.Max != 9 || got.Avg != 7 || got.AggregatorData.N != 2 {
			t.Fatalf("%s: %+v", name, got)
		}
	}
}

func TestCombineTargetHitRangesMergesIntoNilMapsAndCustomKeys(t *testing.T) {
	base := &proto.TargetedActionMetrics{}
	first := &proto.TargetedActionMetrics{
		Ranges: map[int32]*proto.DistributionMetrics{
			int32(proto.HitKind_HitKindHit):  rangeOf(100, 200).toProto(),
			int32(proto.HitKind_HitKindCrit): rangeOf(300).toProto(),
		},
		CustomMetrics: []*proto.CustomActionMetric{{Key: "a", Range: rangeOf(1).toProto()}},
	}
	second := &proto.TargetedActionMetrics{
		Ranges: map[int32]*proto.DistributionMetrics{
			int32(proto.HitKind_HitKindHit): rangeOf(400).toProto(),
		},
		CustomMetrics: []*proto.CustomActionMetric{
			{Key: "a", Range: rangeOf(3).toProto()},
			{Key: "b", Range: rangeOf(8).toProto()},
		},
	}
	combineTargetHitRanges(base, first)
	combineTargetHitRanges(base, second)

	hit := base.Ranges[int32(proto.HitKind_HitKindHit)]
	if hit.AggregatorData.N != 3 || hit.Min != 100 || hit.Max != 400 || math.Abs(hit.Avg-700.0/3) > 1e-9 {
		t.Fatalf("hit range %+v", hit)
	}
	crit := base.Ranges[int32(proto.HitKind_HitKindCrit)]
	if crit.AggregatorData.N != 1 || crit.Min != 300 || crit.Max != 300 {
		t.Fatalf("crit range kept only by the first side %+v", crit)
	}
	if len(base.CustomMetrics) != 2 || base.CustomMetrics[0].Key != "a" || base.CustomMetrics[1].Key != "b" {
		t.Fatalf("custom metrics %+v", base.CustomMetrics)
	}
	if a := base.CustomMetrics[0].Range; a.AggregatorData.N != 2 || a.Avg != 2 || a.Min != 1 || a.Max != 3 {
		t.Fatalf("custom range a %+v", a)
	}
}

func TestActionRangesFromDotTicksMatchCounters(t *testing.T) {
	sim := SetupFakeSim()
	fa := sim.Raid.Parties[0].Players[0].(*FakeAgent)
	fa.Dot.Apply(sim)

	const ticks = 200
	for range ticks {
		fa.Dot.TickOnce(sim)
	}
	target := sim.Encounter.ActiveTargetUnits[0]
	fa.Spell.RecordCustomMetric(target, "bonus", 7)
	fa.Spell.RecordCustomMetric(target, "bonus", 9)

	spellMetrics := fa.Spell.SpellMetrics[target.UnitIndex]
	fa.Spell.doneIteration()
	fa.Spell.doneIteration()

	am := fa.Metrics.actions[fa.Spell.ActionID]
	if len(am.rangeSources) != 1 {
		t.Fatalf("one spell split should register one range source, got %d", len(am.rangeSources))
	}
	tam := am.ToProto(fa.Spell.ActionID).Targets[target.UnitIndex]

	tick := tam.Ranges[int32(proto.HitKind_HitKindTick)]
	critTick := tam.Ranges[int32(proto.HitKind_HitKindCritTick)]
	tickN, critTickN := tick.GetAggregatorData().GetN(), critTick.GetAggregatorData().GetN()
	if critTickN != spellMetrics.CritTicks {
		t.Fatalf("crit tick N %d, counter %d", critTickN, spellMetrics.CritTicks)
	}
	if tickN+critTickN != ticks {
		t.Fatalf("every tick landed, got %d + %d of %d", tickN, critTickN, ticks)
	}

	var total float64
	for kind, r := range tam.Ranges {
		if r.Min > r.Avg || r.Avg > r.Max {
			t.Fatalf("kind %d: min %v avg %v max %v out of order", kind, r.Min, r.Avg, r.Max)
		}
		total += r.Avg * float64(r.AggregatorData.N)
	}
	if !WithinToleranceFloat64(spellMetrics.TotalDamage, total, 0.001) {
		t.Fatalf("ranges sum to %v, spell total %v", total, spellMetrics.TotalDamage)
	}

	if len(tam.CustomMetrics) != 1 || tam.CustomMetrics[0].Key != "bonus" || tam.CustomMetrics[0].Range.Avg != 8 {
		t.Fatalf("custom metrics %+v", tam.CustomMetrics)
	}
}
