package shared

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/dbcenums"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/spelldata"
)

// Mind Quickening Gem 19339 casts 23723, Scrolls of Blinding Light 19343 casts 23733 and Kiss of the
// Spider 22954 casts 28866. The test's on-use runs the first two on a 5 min cooldown in category 1141
// for 20 s, and the third on a 2 min cooldown in category 1141 for 15 s.
const (
	mindQuickening int32 = 23723
	blindingLight  int32 = 23733
	kissOfSpider   int32 = 28866
)

type onUseTrinket struct {
	effect   *proto.ItemEffect
	register func(int32)
}

// A character of the given class and race wearing test trinkets carrying the on-use effects, each
// registered through its own constructor.
func newSpeedOnUseSim(t *testing.T, class proto.Class, race proto.Race, trinkets map[int32]onUseTrinket) (*core.Simulation, *testCaster) {
	t.Helper()
	items := make([]*proto.ItemSpec, proto.ItemSlot_ItemSlotTrinket2+1)
	for i := range items {
		items[i] = &proto.ItemSpec{}
	}

	slot := proto.ItemSlot_ItemSlotTrinket1
	for id, trinket := range trinkets {
		core.AddToDatabase(&proto.SimDatabase{Items: []*proto.SimItem{{Id: id, Name: "Test Trinket",
			Type: proto.ItemType_ItemTypeTrinket, ScalingOptions: map[int32]*proto.ScalingItemProperties{0: {}}}}})
		RegisterItemEffects(map[int32][]*proto.ItemEffect{id: {trinket.effect}})
		trinket.register(id)
		items[slot] = &proto.ItemSpec{Id: id}
		slot++
	}

	sim := newTestCasterSimAgainst(items, nil, 1, func(player *proto.Player) {
		player.Class = class
		player.Race = race
	})
	return sim, sim.Raid.Parties[0].Players[0].(*testCaster)
}

type speeds struct{ melee, ranged, cast float64 }

func speedsOf(caster *testCaster) speeds {
	return speeds{caster.PseudoStats.MeleeSpeedMultiplier, caster.PseudoStats.RangedSpeedMultiplier, caster.PseudoStats.CastSpeedMultiplier}
}

func (got speeds) differ(want speeds) bool {
	return math.Abs(got.melee-want.melee) > 1e-9 || math.Abs(got.ranged-want.ranged) > 1e-9 || math.Abs(got.cast-want.cast) > 1e-9
}

// Using the item multiplies the wearer's speeds by the ones its row states for the row's duration, on
// one aura, and every speed returns to what it was when the aura runs out.
func TestOnUseSpeedBuffMultipliesTheWearersSpeedsForItsDuration(t *testing.T) {
	haste := func(spellID int32, aura dbcenums.EffectAuraType) float64 {
		percent := spelldata.MustFind(spellID).FirstAura(aura).BasePoints
		needAmount(t, fmt.Sprintf("%d's %v", spellID, aura), percent)
		return 1 + percent/100
	}
	duration := func(spellID int32) time.Duration { return spelldata.MustFind(spellID).Duration() }

	for _, tc := range []struct {
		name     string
		itemID   int32
		class    proto.Class
		race     proto.Race
		effect   *proto.ItemEffect
		want     speeds
		duration time.Duration
	}{
		{"Mind Quickening Gem", 991140, proto.Class_ClassMage, proto.Race_RaceTroll,
			testOnUse(mindQuickening, 300000, 1141, 20000),
			speeds{1, 1, haste(mindQuickening, dbcenums.A_MOD_CASTING_SPEED_NOT_STACK)}, duration(mindQuickening)},
		{"Scrolls of Blinding Light", 991141, proto.Class_ClassPaladin, proto.Race_RaceHuman,
			testOnUse(blindingLight, 300000, 1141, 20000),
			speeds{haste(blindingLight, dbcenums.A_MOD_MELEE_HASTE_3), 1, haste(blindingLight, dbcenums.A_MOD_CASTING_SPEED_NOT_STACK)},
			duration(blindingLight)},
		{"Kiss of the Spider", 991142, proto.Class_ClassWarrior, proto.Race_RaceOrc,
			testOnUse(kissOfSpider, 120000, 1141, 15000),
			speeds{haste(kissOfSpider, dbcenums.A_MOD_MELEE_HASTE_3), haste(kissOfSpider, dbcenums.A_MOD_RANGED_HASTE), 1},
			duration(kissOfSpider)},
	} {
		sim, caster := newSpeedOnUseSim(t, tc.class, tc.race, map[int32]onUseTrinket{tc.itemID: {tc.effect, NewSpellDataSpeedOnUse}})
		spell := onUseSpell(t, caster, tc.itemID)
		if got := caster.GetInitialMajorCooldown(spell.ActionID); !got.Type.Matches(core.CooldownTypeDPS) {
			t.Errorf("%s is a major cooldown of type %v, want a DPS one", tc.name, got.Type)
		}

		before := speedsOf(caster)
		if !spell.Cast(sim, &caster.Unit) {
			t.Fatalf("%s could not be used", tc.name)
		}
		start := sim.CurrentTime
		want := speeds{before.melee * tc.want.melee, before.ranged * tc.want.ranged, before.cast * tc.want.cast}
		if got := speedsOf(caster); got.differ(want) {
			t.Errorf("%s: speeds after use %+v, want %+v", tc.name, got, want)
		}

		buff := spell.RelatedSelfBuff
		if buff == nil || !buff.IsActive() || buff.Duration != tc.duration {
			t.Fatalf("%s: buff %v, want one aura up for %v", tc.name, buff, tc.duration)
		}

		stepPast(t, sim, start+tc.duration-time.Millisecond)
		if got := speedsOf(caster); got.differ(want) {
			t.Errorf("%s: speeds just before the buff runs out %+v, want %+v", tc.name, got, want)
		}
		stepPast(t, sim, start+tc.duration+time.Millisecond)
		if got := speedsOf(caster); buff.IsActive() || got.differ(before) {
			t.Errorf("%s: buff up %v, speeds %+v after %v; want it gone and the speeds back to %+v",
				tc.name, buff.IsActive(), got, tc.duration, before)
		}
	}
}

// The test gem's own 5 min cooldown, and its category's 20 s that the test lasso, also in category
// 1141, waits out as well.
func TestOnUseSpeedBuffRunsOnTheItemsCooldownAndCategory(t *testing.T) {
	const gemID, lassoID int32 = 991143, 991144
	sim, caster := newSpeedOnUseSim(t, proto.Class_ClassMage, proto.Race_RaceTroll, map[int32]onUseTrinket{
		gemID:   {testOnUse(mindQuickening, 300000, lassoCategory, 20000), NewSpellDataSpeedOnUse},
		lassoID: {lassoOnUse(), NewSpellDataDamageOnUse},
	})
	gem, lasso := onUseSpell(t, caster, gemID), onUseSpell(t, caster, lassoID)

	gem.Cast(sim, &caster.Unit)
	start := sim.CurrentTime
	if offCooldown(sim, lasso) {
		t.Errorf("the lasso could be used while the gem's category %d cooldown runs", lassoCategory)
	}

	stepPast(t, sim, start+20*time.Second)
	if !offCooldown(sim, lasso) {
		t.Errorf("the lasso could not be used once the gem's 20 s category cooldown ran out")
	}
	if offCooldown(sim, gem) {
		t.Errorf("the gem could be used again after 20 s, inside its 5 min cooldown")
	}
	if got := gem.CD.ReadyAt(); got != start+5*time.Minute {
		t.Errorf("the gem is ready again at %v, want 5 min after its use at %v", got, start)
	}
}
