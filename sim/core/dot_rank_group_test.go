package core_test

import (
	"testing"
	"time"

	"github.com/wowsims/forever/sim/core"
	"github.com/wowsims/forever/sim/core/proto"
	"github.com/wowsims/forever/sim/core/simsignals"
	"github.com/wowsims/forever/sim/core/spelldata"
)

const (
	shadowWordPain1, shadowWordPain2 int32 = 589, 594
	holyFire1, holyFire2             int32 = 14914, 15262
	renew1, renew2                   int32 = 139, 6074
	mindFlay1, mindFlay2             int32 = 15407, 17311
	explosiveTrap1, explosiveTrap2   int32 = 13812, 14314
)

// What each test priest registers in Initialize, set by the test before it builds the sim.
var rankSetup func(c *core.Character)

type rankCaster struct {
	core.Character
}

func init() {
	core.RegisterAgentFactory(
		proto.Player_HealerPriest{},
		proto.Spec_SpecHealerPriest,
		func(character *core.Character, _ *proto.Player, _ *proto.Raid) core.Agent {
			return &rankCaster{Character: *character}
		},
		func(player *proto.Player, spec any) {
			player.Spec = spec.(*proto.Player_HealerPriest)
		},
	)
}

func (c *rankCaster) GetCharacter() *core.Character       { return &c.Character }
func (c *rankCaster) ApplyTalents()                       {}
func (c *rankCaster) Reset(_ *core.Simulation)            {}
func (c *rankCaster) OnGCDReady(_ *core.Simulation)       {}
func (c *rankCaster) OnEncounterStart(_ *core.Simulation) {}

func (c *rankCaster) Initialize() {
	c.EnableManaBar()
	rankSetup(&c.Character)
}

// One priest per name, each registering what setup gives it, against one level 60 target.
func newRankSim(t *testing.T, setup func(c *core.Character), names ...string) (*core.Simulation, *core.Unit, []*core.Character) {
	t.Helper()
	rankSetup = setup
	t.Cleanup(func() { rankSetup = nil })

	players := make([]*proto.Player, len(names))
	for i, name := range names {
		players[i] = &proto.Player{Name: name, Class: proto.Class_ClassPriest, Race: proto.Race_RaceHuman,
			Buffs: &proto.IndividualBuffs{}, Spec: &proto.Player_HealerPriest{}, Equipment: &proto.EquipmentSpec{}}
	}
	sim := core.NewSim(&proto.RaidSimRequest{
		SimOptions: &proto.SimOptions{RandomSeed: 1},
		Raid:       &proto.Raid{Parties: []*proto.Party{{Players: players, Buffs: &proto.PartyBuffs{}}}},
		Encounter: &proto.Encounter{
			Targets:  []*proto.Target{{Name: "target", Level: 60, MobType: proto.MobType_MobTypeDemon}},
			Duration: 180,
		},
	}, simsignals.CreateSignals())
	sim.Reset()

	casters := make([]*core.Character, len(names))
	for i, player := range sim.Raid.Parties[0].Players {
		casters[i] = player.GetCharacter()
	}
	return sim, sim.Encounter.ActiveTargetUnits[0], casters
}

// One priest registering two ranks through register, against one level 60 target.
func newRankPair(t *testing.T, register func(c *core.Character, id int32) *core.Spell, id1, id2 int32) (
	sim *core.Simulation, target *core.Unit, priest *core.Character, rank1, rank2 *core.Spell) {
	t.Helper()
	sim, target, casters := newRankSim(t, func(c *core.Character) {
		rank1 = register(c, id1)
		rank2 = register(c, id2)
	}, "Priest")
	return sim, target, casters[0], rank1, rank2
}

func stepTo(t *testing.T, sim *core.Simulation, until time.Duration) {
	t.Helper()
	sim.AddPendingAction(core.NewDelayedAction(core.DelayedActionOptions{DoAt: until, OnAction: func(*core.Simulation) {}}))
	for steps := 0; sim.CurrentTime < until; steps++ {
		if steps > 10000 {
			t.Fatalf("the sim did not reach %v", until)
		}
		sim.Step()
	}
}

// A rank whose cast puts the row's periodic on the target, with the row's direct damage first where it
// states one.
func registerTargetDot(c *core.Character, id int32) *core.Spell {
	row := spelldata.MustFind(id)
	config := spelldata.SpellConfig(&c.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
	config.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
	direct := row.DamageEffect()
	config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		if direct != spelldata.NilEffect {
			spell.CalcAndDealDamage(sim, target, direct.Roll(sim), spell.OutcomeAlwaysHit)
		}
		spell.Dot(target).Apply(sim)
	}
	return c.RegisterSpell(config)
}

func TestDotRankHigherReplacesLower(t *testing.T) {
	sim, target, _, rank1, rank2 := newRankPair(t, registerTargetDot, shadowWordPain1, shadowWordPain2)

	if !rank1.Cast(sim, target) {
		t.Fatalf("Shadow Word: Pain Rank 1 could not be cast on a target without it")
	}
	stepTo(t, sim, 2*time.Second)
	if !rank2.Cast(sim, target) {
		t.Fatalf("Shadow Word: Pain Rank 2 could not be cast over Rank 1")
	}
	if rank1.Dot(target).IsActive() || !rank2.Dot(target).IsActive() {
		t.Errorf("after Rank 2 over Rank 1: Rank 1 up %v, Rank 2 up %v; want only Rank 2",
			rank1.Dot(target).IsActive(), rank2.Dot(target).IsActive())
	}
}

// Shadow Word: Pain's only effect is the periodic, so its lower rank is refused before it spends
// anything wherever the dot lives; a rank registered with Proc() has no cast requirement and casts,
// with its periodic skipped. The higher rank runs on to its own expiry either way.
func TestDotRankLowerPureRankUnderHigher(t *testing.T) {
	for _, tc := range []struct {
		name     string
		register func(c *core.Character, id int32) *core.Spell
		refused  bool
	}{
		{"own dot", registerTargetDot, true},
		{"RelatedDotSpell", func(c *core.Character, id int32) *core.Spell {
			row := spelldata.MustFind(id)
			dotConfig := spelldata.SpellConfig(&c.Unit, row, spelldata.Proc(), spelldata.Tag(1))
			dotConfig.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
			config := spelldata.SpellConfig(&c.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
			config.RelatedDotSpell = c.RegisterSpell(dotConfig)
			config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Dot(target).Apply(sim)
			}
			return c.RegisterSpell(config)
		}, true},
		{"Proc", func(c *core.Character, id int32) *core.Spell {
			row := spelldata.MustFind(id)
			config := spelldata.SpellConfig(&c.Unit, row, spelldata.Proc())
			config.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect())
			config.ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
				spell.Dot(target).Apply(sim)
			}
			return c.RegisterSpell(config)
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sim, target, priest, rank1, rank2 := newRankPair(t, tc.register, shadowWordPain1, shadowWordPain2)

			rank2.Cast(sim, target)
			expiry := rank2.Dot(target).ExpiresAt()
			stepTo(t, sim, 2*time.Second)

			mana := priest.CurrentMana()
			if canCast, cast := rank1.CanCast(sim, target), rank1.Cast(sim, target); canCast == tc.refused || cast == tc.refused {
				t.Errorf("Rank 1 under Rank 2: CanCast %v, Cast %v; want both %v", canCast, cast, !tc.refused)
			}
			if priest.CurrentMana() != mana || !priest.GCD.IsReady(sim) {
				t.Errorf("Rank 1 spent %v mana and left the global cooldown ready %v; want none spent and ready",
					mana-priest.CurrentMana(), priest.GCD.IsReady(sim))
			}
			if casts := rank1.SpellMetrics[target.UnitIndex].Casts; tc.refused && casts != 0 {
				t.Errorf("the refused Rank 1 counted %d casts, want 0", casts)
			}
			if rank1.Dot(target).IsActive() || rank2.Dot(target).ExpiresAt() != expiry {
				t.Errorf("Rank 1 up %v, Rank 2 expires at %v; want Rank 1 down and Rank 2 at %v",
					rank1.Dot(target).IsActive(), rank2.Dot(target).ExpiresAt(), expiry)
			}
		})
	}
}

// Holy Fire also deals direct damage, so the lower rank casts and lands it while the higher rank's
// periodic stays.
func TestDotRankLowerMixedRankLandsItsDirectPart(t *testing.T) {
	sim, target, _, rank1, rank2 := newRankPair(t, registerTargetDot, holyFire1, holyFire2)

	rank2.Cast(sim, target)
	stepTo(t, sim, rank2.DefaultCast.CastTime+time.Millisecond)
	expiry := rank2.Dot(target).ExpiresAt()

	if !rank1.CanCast(sim, target) || !rank1.Cast(sim, target) {
		t.Fatalf("Holy Fire Rank 1 was refused while Rank 2 is up")
	}
	stepTo(t, sim, sim.CurrentTime+rank1.DefaultCast.CastTime+time.Millisecond)

	if rank1.SpellMetrics[target.UnitIndex].TotalDamage == 0 {
		t.Errorf("Rank 1 dealt no direct damage")
	}
	if rank1.Dot(target).IsActive() || rank2.Dot(target).ExpiresAt() != expiry {
		t.Errorf("Rank 1's periodic up %v, Rank 2's expires at %v; want Rank 1's skipped and Rank 2's at %v",
			rank1.Dot(target).IsActive(), rank2.Dot(target).ExpiresAt(), expiry)
	}
}

// Both ranks keep the default label, the ranked name, so the lower rank's aura is its own and a cast
// of it over itself refreshes.
func TestDotRankSameRankRefreshes(t *testing.T) {
	sim, target, _, rank1, _ := newRankPair(t, registerTargetDot, shadowWordPain1, shadowWordPain2)

	rank1.Cast(sim, target)
	firstExpiry := rank1.Dot(target).ExpiresAt()
	stepTo(t, sim, 6*time.Second)
	if !rank1.Cast(sim, target) {
		t.Fatalf("Rank 1 was refused over itself")
	}
	if got, want := rank1.Dot(target).ExpiresAt(), firstExpiry+6*time.Second; got != want {
		t.Errorf("Rank 1 cast at 0 s expiring at %v, recast at 6 s, expires at %v; want %v", firstExpiry, got, want)
	}
}

// A config built by hand leaves the spell's Rank at 0: the group still orders the ranks by the rows.
func TestDotRankHandBuiltSpellsCompareByRow(t *testing.T) {
	sim, target, _, rank1, rank2 := newRankPair(t, func(c *core.Character, id int32) *core.Spell {
		row := spelldata.MustFind(id)
		return c.RegisterSpell(core.SpellConfig{
			ActionID:         core.ActionID{SpellID: id},
			SpellSchool:      core.SpellSchoolShadow,
			ProcMask:         core.ProcMaskSpellDamage,
			DamageMultiplier: 1,
			ThreatMultiplier: 1,
			Dot:              spelldata.DotConfig(row, row.PeriodicDamageEffect()),
		})
	}, shadowWordPain1, shadowWordPain2)

	rank2.Dot(target).Apply(sim)
	rank1.Dot(target).Apply(sim)
	if rank1.Dot(target).IsActive() || !rank2.Dot(target).IsActive() {
		t.Errorf("Rank 1 applied over Rank 2: Rank 1 up %v, Rank 2 up %v; want only Rank 2",
			rank1.Dot(target).IsActive(), rank2.Dot(target).IsActive())
	}
}

func TestDotRankTwoCastersStayApart(t *testing.T) {
	ranks := map[*core.Character][2]*core.Spell{}
	sim, target, casters := newRankSim(t, func(c *core.Character) {
		ranks[c] = [2]*core.Spell{registerTargetDot(c, shadowWordPain1), registerTargetDot(c, shadowWordPain2)}
	}, "Priest A", "Priest B")
	a, b := ranks[casters[0]], ranks[casters[1]]

	a[1].Cast(sim, target)
	if !b[0].CanCast(sim, target) || !b[0].Cast(sim, target) {
		t.Fatalf("Priest B's Rank 1 was refused under Priest A's Rank 2")
	}
	if !a[1].Dot(target).IsActive() || !b[0].Dot(target).IsActive() {
		t.Errorf("Priest A's Rank 2 up %v, Priest B's Rank 1 up %v; want both",
			a[1].Dot(target).IsActive(), b[0].Dot(target).IsActive())
	}
}

// A channel keeps no group: a new channel already ends the one running.
func TestDotRankChannelsKeepNoGroup(t *testing.T) {
	sim, target, _, rank1, rank2 := newRankPair(t, registerTargetDot, mindFlay1, mindFlay2)

	rank2.Dot(target).Apply(sim)
	rank1.Dot(target).Apply(sim)
	if !rank1.Dot(target).IsActive() {
		t.Errorf("Mind Flay Rank 1 did not take hold with Rank 2 up")
	}
}

func TestDotRankSelfHotKeepsOneRank(t *testing.T) {
	sim, _, caster, rank1, rank2 := newRankPair(t, func(c *core.Character, id int32) *core.Spell {
		row := spelldata.MustFind(id)
		config := spelldata.SpellConfig(&c.Unit, row, spelldata.Magic(core.ProcMaskSpellHealing))
		config.Hot = spelldata.DotConfig(row, row.PeriodicEffect())
		config.Hot.SelfOnly = true
		config.Hot.OnTick = func(*core.Simulation, *core.Unit, *core.Dot) {}
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.SelfHot().Apply(sim)
		}
		return c.RegisterSpell(config)
	}, renew1, renew2)
	priest := &caster.Unit

	rank1.Cast(sim, priest)
	stepTo(t, sim, 2*time.Second)
	rank2.Cast(sim, priest)
	if rank1.SelfHot().IsActive() || !rank2.SelfHot().IsActive() {
		t.Fatalf("after Renew Rank 2 over Rank 1: Rank 1 up %v, Rank 2 up %v; want only Rank 2",
			rank1.SelfHot().IsActive(), rank2.SelfHot().IsActive())
	}
	stepTo(t, sim, 4*time.Second)
	if rank1.CanCast(sim, priest) {
		t.Errorf("CanCast answered true for Renew Rank 1 while Rank 2 is up on the caster")
	}
}

// One ground effect of an ability per caster: a cast of any rank ends the one down.
func TestDotRankGroundEffectReplacedByAnyRank(t *testing.T) {
	sim, target, _, rank1, rank2 := newRankPair(t, func(c *core.Character, id int32) *core.Spell {
		row := spelldata.MustFind(id)
		config := spelldata.SpellConfig(&c.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
		config.Dot = spelldata.DotConfig(row, &row.Effects[1]) // The persistent area aura after the direct damage.
		config.Dot.IsAOE = true
		config.ApplyEffects = func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			spell.AOEDot().Apply(sim)
		}
		return c.RegisterSpell(config)
	}, explosiveTrap1, explosiveTrap2)

	rank2.Cast(sim, target)
	stepTo(t, sim, 2*time.Second)
	if !rank1.CanCast(sim, target) || !rank1.Cast(sim, target) {
		t.Fatalf("Explosive Trap Effect Rank 1 was refused while Rank 2's ground effect is down")
	}
	if !rank1.AOEDot().IsActive() || rank2.AOEDot().IsActive() {
		t.Errorf("after Rank 1 over Rank 2: Rank 1 up %v, Rank 2 up %v; want only Rank 1",
			rank1.AOEDot().IsActive(), rank2.AOEDot().IsActive())
	}
}

func TestDotRanksSharingAnAuraPanic(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("two ranks registered with one aura label did not panic")
		}
	}()
	newRankSim(t, func(c *core.Character) {
		for _, id := range []int32{shadowWordPain1, shadowWordPain2} {
			row := spelldata.MustFind(id)
			config := spelldata.SpellConfig(&c.Unit, row, spelldata.Magic(core.ProcMaskSpellDamage))
			config.Dot = spelldata.DotConfig(row, row.PeriodicDamageEffect(), spelldata.Label(row.Name))
			c.RegisterSpell(config)
		}
	}, "Priest")
}
