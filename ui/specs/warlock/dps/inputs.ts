import * as InputHelpers from '@features/inputs/input_helpers';
import { Spec } from '@generated/proto/common';
import { WarlockOptions_Armor, WarlockOptions_CurseOptions, WarlockOptions_Summon as Summon } from '@generated/proto/warlock';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { WarlockSpecs } from '@sim/proto/spec_types';
import type { CustomSection } from '@sim/spec_config';

// Configuration for spec-specific UI elements on the settings tab.
// These don't need to be in a separate file but it keeps things cleaner.

type DemonField = 'summon' | 'sacrificedDemon';

// Settles the demon out and the demon sacrificed against each other and the talents, with `pick` (the
// value just chosen in one of the two pickers) winning over the other, and otherwise the rules the sim
// settles them by:
//   - Without Demonic Sacrifice, nothing is sacrificed.
//   - Without Demonic Pact, any sacrifice leaves no demon out.
//   - With it, resummoning the sacrificed demon cancels the effect, so the demon stays out.
export const settleDemons = <SpecType extends WarlockSpecs>(player: Player<SpecType>, pick?: { field: DemonField; value: Summon }) => {
	const options = player.getClassOptions();
	const talents = player.getTalents();
	if (pick) options[pick.field] = pick.value;

	if (!talents.demonicSacrifice) options.sacrificedDemon = Summon.NoSummon;

	const sacrificed = options.sacrificedDemon != Summon.NoSummon;
	if (sacrificed && options.summon != Summon.NoSummon && (options.summon == options.sacrificedDemon || !talents.demonicPact)) {
		const keepSummon = pick ? pick.field == 'summon' : talents.demonicPact;
		if (keepSummon) options.sacrificedDemon = Summon.NoSummon;
		else options.summon = Summon.NoSummon;
	}

	player.setClassOptions(options);
};

export const PetInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, Summon>({
		fieldName: 'summon',
		label: i18n.t('settings_tab.other.warlock_pet.label'),
		labelTooltip: i18n.t('settings_tab.other.warlock_pet.tooltip'),
		numColumns: 5,
		values: [
			{ value: Summon.NoSummon, tooltip: 'No Pet' },
			{ actionId: ActionId.fromSpellId(688), value: Summon.Imp },
			{ actionId: ActionId.fromSpellId(697), value: Summon.Voidwalker },
			{ actionId: ActionId.fromSpellId(712), value: Summon.Succubus },
			{ actionId: ActionId.fromSpellId(691), value: Summon.Felhunter },
		],
		storeField: 'player:*' as const,
		setValue: (player: Player<SpecType>, newValue: Summon) => settleDemons(player, { field: 'summon', value: newValue }),
	});

// Demon Armor on or off: the only armor the game files carry for Forever's warlock.
export const ArmorInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'armor',
		// The rank the sim casts (armors.go).
		id: ActionId.fromSpellId(11735),
		value: WarlockOptions_Armor.DemonArmor,
		label: i18n.t('settings_tab.other.warlock_armor.label'),
		labelTooltip: i18n.t('settings_tab.other.warlock_armor.tooltip'),
	});

// The Demonic Sacrifice effect up when the fight starts, by the demon sacrificed.
export const DemonicSacrificeInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsEnumIconInput<SpecType, Summon>({
		fieldName: 'sacrificedDemon',
		label: i18n.t('settings_tab.other.warlock_demonic_sacrifice.label'),
		labelTooltip: i18n.t('settings_tab.other.warlock_demonic_sacrifice.tooltip'),
		numColumns: 5,
		values: [
			{ value: Summon.NoSummon, tooltip: 'No Sacrifice' },
			// The demon's summon icon, with the tooltip of the effect sacrificing it gives.
			{ actionId: ActionId.fromSpellId(688), buffActionId: ActionId.fromSpellId(18789), value: Summon.Imp },
			{ actionId: ActionId.fromSpellId(697), buffActionId: ActionId.fromSpellId(18792), value: Summon.Voidwalker },
			{ actionId: ActionId.fromSpellId(712), buffActionId: ActionId.fromSpellId(18791), value: Summon.Succubus },
			{ actionId: ActionId.fromSpellId(691), buffActionId: ActionId.fromSpellId(18790), value: Summon.Felhunter },
		],
		// `talentsString` too, so untalenting Demonic Sacrifice hides the picker straight away.
		storeField: ['specOptions', 'talentsString'] as const,
		showWhen: (player: Player<SpecType>) => player.getTalents().demonicSacrifice,
		setValue: (player: Player<SpecType>, newValue: Summon) => settleDemons(player, { field: 'sacrificedDemon', value: newValue }),
	});

export const SoulShardsInput = <SpecType extends WarlockSpecs>() =>
	InputHelpers.makeClassOptionsNumberInput<SpecType>({
		fieldName: 'soulShards',
		label: i18n.t('settings_tab.other.warlock_soul_shards.label'),
		labelTooltip: i18n.t('settings_tab.other.warlock_soul_shards.tooltip'),
		positive: true,
	});

const makeCursePicker = <SpecType extends WarlockSpecs>(curse: WarlockOptions_CurseOptions, spellId: number) =>
	InputHelpers.makeClassOptionsBooleanIconInput<SpecType>({
		fieldName: 'curseOptions',
		id: ActionId.fromSpellId(spellId),

		getValue: (player: Player<SpecType>) => player.getClassOptions().curseOptions === curse,

		// Clicking the assigned curse again unassigns it.
		setValue: (player: Player<SpecType>, newValue: boolean) => {
			const newOptions = player.getClassOptions();
			newOptions.curseOptions = newValue ? curse : WarlockOptions_CurseOptions.NoCurse;

			player.setClassOptions(newOptions);
		},
	});

// The assigned-curse block is declared as data; `CustomSection` renders it from the settings tab.
export const CursesSection: CustomSection<Spec.SpecWarlock> = {
	id: 'assigned-curse-settings',
	title: i18n.t('settings_tab.other.warlock_assigned_curse.title'),
	description: i18n.t('settings_tab.other.warlock_assigned_curse.description'),
	// Agony and Doom are Banes now, cast directly rather than as the assigned curse. The ids are the
	// ranks the sim casts (curse_of_elements.go, curse_of_recklessness.go, curse_of_weakness.go).
	iconInputs: [
		makeCursePicker<Spec.SpecWarlock>(WarlockOptions_CurseOptions.Elements, 1311680),
		makeCursePicker<Spec.SpecWarlock>(WarlockOptions_CurseOptions.Recklessness, 11717),
		makeCursePicker<Spec.SpecWarlock>(WarlockOptions_CurseOptions.Weakness, 11708),
	],
};
