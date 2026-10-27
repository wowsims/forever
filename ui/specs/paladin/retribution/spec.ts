import * as BuffDebuffInputs from '@features/settings/model/buffs_debuffs';
import * as OtherInputs from '@features/settings/model/other_inputs';
import { APLGroup, APLListItem, APLRotation, APLRotation_Type, APLValueVariable } from '@generated/proto/apl';
import { Cooldowns, ItemSlot, PseudoStat, Spec, Stat } from '@generated/proto/common';
import { PaladinSeal } from '@generated/proto/paladin';
import { PlayerClasses } from '@sim/player/classes';
import { Player } from '@sim/player/player';
import { WeaponSkillGroup } from '@sim/player/weapon_skill_groups';
import * as AplUtils from '@sim/proto/apl_utils';
import { SpecRotation } from '@sim/proto/spec_types';
import { Stats, UnitStat } from '@sim/proto/stats';
import { defineSpec } from '@sim/spec_config';

import * as Inputs from './inputs';
import * as Presets from './presets';

const aplFor = (player: Player<Spec.SpecRetributionPaladin>): APLRotation =>
	player.getTalents().holyShock ? Presets.HolyShockAplPreset.rotation.rotation! : Presets.DefaultAplPreset.rotation.rotation!;

// A group that casts the one spell.
const castGroup = (name: string, spellId: { spellId: number; rank?: number }) => APLGroup.fromJson({ name, actions: [{ action: { castSpell: { spellId } } }] });

export default defineSpec<Spec.SpecRetributionPaladin>({
	spec: Spec.SpecRetributionPaladin,

	className: 'retribution-paladin-sim-ui',
	cssScheme: PlayerClasses.getCssScheme(PlayerClasses.Paladin),
	// List any known bugs / issues here and they'll be shown on the site.
	knownIssues: [],
	consumableStats: [Stat.StatMana, Stat.StatHolyDamage],

	// All stats for which EP should be calculated.
	epStats: [
		Stat.StatStrength,
		Stat.StatAgility,
		Stat.StatIntellect,
		Stat.StatSpellDamage,
		Stat.StatSpellPiercing,
		Stat.StatAttackPower,
		Stat.StatArmorPenetration,
		Stat.StatHitRating,
		Stat.StatHasteRating,
		Stat.StatCritRating,
		Stat.StatExpertiseRating,
		Stat.StatMP5,
	],
	epPseudoStats: [PseudoStat.PseudoStatMainHandDps, PseudoStat.PseudoStatSchoolHitPercentHoly],
	// Reference stat against which to calculate EP. I think all classes use either spell power or attack power.
	epReferenceStat: Stat.StatAttackPower,
	// Which stats to display in the Character Stats section, at the bottom of the left-hand sidebar.
	displayStats: UnitStat.createDisplayStatArray(
		[
			Stat.StatStrength,
			Stat.StatAgility,
			Stat.StatIntellect,
			Stat.StatAttackPower,
			Stat.StatSpellDamage,
			Stat.StatMP5,
			Stat.StatMana,
			Stat.StatHealth,
			Stat.StatStamina,
			Stat.StatHolyDamage,
			Stat.StatSpellPiercing,
		],
		[
			PseudoStat.PseudoStatMeleeHitPercent,
			PseudoStat.PseudoStatMeleeCritPercent,
			PseudoStat.PseudoStatMeleeHastePercent,
			PseudoStat.PseudoStatMeleeSpeedMultiplier,
			PseudoStat.PseudoStatSpellHastePercent,
			PseudoStat.PseudoStatSpellCritPercent,
			PseudoStat.PseudoStatSpellHitPercent,
			PseudoStat.PseudoStatSchoolHitPercentHoly,
			PseudoStat.PseudoStatExpertisePercent,
		],
	),
	displayWeaponSkillGroups: [WeaponSkillGroup.Melee],

	defaults: {
		// Default equipped gear.
		gear: Presets.DefaultGearPreset.gear,
		// Default EP weights for sorting gear in the gear picker.
		epWeights: Presets.TwistOfLightEpWeights.epWeights,
		statCaps: (() => {
			const hitCap = new Stats().withPseudoStat(PseudoStat.PseudoStatMeleeHitPercent, 8);
			const expCap = new Stats().withPseudoStat(PseudoStat.PseudoStatExpertisePercent, 6.5);

			return hitCap.add(expCap);
		})(),
		// Default consumes settings.
		consumables: Presets.DefaultConsumables,
		// Default talents.
		talents: Presets.TwistOfLightTalents.data,
		// Default spec-specific settings.
		specOptions: Presets.DefaultOptions,
		other: Presets.OtherDefaults,
		// Default raid/party buffs settings.
		raidBuffs: Presets.DefaultRaidBuffs,
		partyBuffs: Presets.DefaultPartyBuffs,
		individualBuffs: Presets.DefaultIndividualBuffs,
		debuffs: Presets.DefaultDebuffs,

		rotationType: APLRotation_Type.TypeSimple,
		simpleRotation: Presets.DefaultSimpleRotation,
		aplRotation: Presets.DefaultAplPreset.rotation.rotation!,
	},

	// IconInputs to include in the 'Player' section on the settings tab.
	playerIconInputs: [],
	// Buff and Debuff inputs to include/exclude, overriding the EP-based defaults.
	includeBuffDebuffInputs: [Stat.StatDefenseRating, Stat.StatMP5, Stat.StatHolyDamage],
	excludeBuffDebuffInputs: [
		BuffDebuffInputs.AtieshDruid,
		BuffDebuffInputs.AtieshMage,
		BuffDebuffInputs.AtieshPriest,
		BuffDebuffInputs.AtieshWarlock,
		BuffDebuffInputs.ConcentrationAura,
		BuffDebuffInputs.CurseOfWeakness,
		BuffDebuffInputs.DemoralizingRoar,
		BuffDebuffInputs.DemoralizingScreech,
		BuffDebuffInputs.DemoralizingShout,
		BuffDebuffInputs.HexOfWeakness,
		BuffDebuffInputs.InsectSwarm,
		BuffDebuffInputs.ScorpidSting,
		BuffDebuffInputs.StoneskinTotem,
		BuffDebuffInputs.Thorns,
		BuffDebuffInputs.ThunderClap,
		BuffDebuffInputs.Thunderfury,
		BuffDebuffInputs.Vindication,
	],
	rotationInputs: Inputs.RotationInputs,
	rotationIconInputs: Inputs.RotationIconInputs,
	// Inputs to include in the 'Other' section on the settings tab.
	otherInputs: {
		inputs: [OtherInputs.InputDelay, OtherInputs.TankAssignment, OtherInputs.InFrontOfTarget],
	},
	itemSwapSlots: [ItemSlot.ItemSlotMainHand, ItemSlot.ItemSlotOffHand],
	encounterPicker: {
		// Whether to include 'Execute Duration (%)' in the 'Encounter' section of the settings tab.
		showExecuteProportion: false,
	},

	presets: {
		epWeights: [Presets.TwistOfLightEpWeights, Presets.HolyShockEpWeights],
		rotations: [Presets.SimpleRotation, Presets.DefaultAplPreset, Presets.HolyShockAplPreset],
		// Preset talents that the user can quickly select.
		talents: [Presets.TwistOfLightTalents, Presets.HolyShockTalents],
		// Preset gear configurations that the user can quickly select.
		gear: [Presets.DefaultGearPreset, Presets.HolyShockGearPreset],
		settings: [Presets.DefaultSettings, Presets.HolyShockSettings],
		builds: [Presets.DefaultBuild, Presets.HolyShockBuild],
	},

	autoRotation: aplFor,

	// The talents pick the APL, as for the Auto rotation; the options set its variables and the spells of
	// its aura and seal groups. With no assigned seal, Has assigned seal is false and the APL never
	// reaches Cast assigned seal.
	simpleRotation: (player: Player<Spec.SpecRetributionPaladin>, simple: SpecRotation<Spec.SpecRetributionPaladin>, cooldowns: Cooldowns): APLRotation => {
		const { holyShock, twistOfLight, sealOfCommand } = player.getTalents();
		const actions = AplUtils.simpleCooldownActions(cooldowns);
		const rotation = APLRotation.clone(aplFor(player));
		const aura = Inputs.AURA_SPELLS[simple.aura];
		const judgement = Inputs.JUDGE_SPELLS[simple.assignedJudgement];

		const values: Record<string, string> = {
			'Has assigned judgement': String(!!judgement),
			'Pull range': String(Math.max(Inputs.MIN_PULL_RANGE, simple.pullRange || 25)),
			'Delay Major CDs': `${simple.delayMajorCds}s`,
		};

		if (!holyShock) {
			values['Use Seal of Command'] = String(sealOfCommand && (twistOfLight || simple.nonTwistSeal === PaladinSeal.SealOfCommand));
			values['Use Seal of Righteousness'] = String(twistOfLight || simple.nonTwistSeal === PaladinSeal.SealOfRighteousness);
		}

		rotation.valueVariables = rotation.valueVariables.map(variable =>
			variable.name in values ? APLValueVariable.fromJson({ name: variable.name, value: { const: { val: values[variable.name] } } }) : variable,
		);

		const groups: Record<string, APLGroup> = {};
		if (aura) groups['Activate assigned aura'] = castGroup('Activate assigned aura', aura);
		if (judgement) groups['Cast assigned judgement'] = castGroup('Cast assigned judgement', judgement);
		rotation.groups = rotation.groups.map(group => groups[group.name] ?? group);

		return APLRotation.create({
			prepullActions: rotation.prepullActions,
			priorityList: [...actions.map(action => APLListItem.create({ action })), ...rotation.priorityList],
			groups: rotation.groups,
			valueVariables: rotation.valueVariables,
		});
	},

	reforge: {
		getEPDefaults: (player: Player<Spec.SpecRetributionPaladin>): Stats => {
			return player.getTalents().holyShock ? Presets.HolyShockEpWeights.epWeights : Presets.TwistOfLightEpWeights.epWeights;
		},
	},
});
