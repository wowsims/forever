import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import { Faction, Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { ActionId } from '@sim/proto/action_id';
import { Party } from '@sim/raid/party';
import { Raid } from '@sim/raid/raid';

import * as InputHelpers from './input_helpers';

export type IconInputConfig<ModObject, T> = InputHelpers.TypedIconPickerConfig<ModObject, T> | InputHelpers.TypedIconEnumPickerConfig<ModObject, T>;

interface BooleanInputConfig<T> {
	actionId: ActionId;
	fieldName: keyof T;
	value?: number;
	label?: string;
	// Faction comes off the player's race, so only the factories whose mod object is the player can
	// honour it; the party- and raid-scoped ones omit it rather than accept it and drop it.
	faction?: Faction;
	showWhen?: (player: Player<any>) => boolean;
	enableWhen?: (player: Player<any>) => boolean;
}

export const makeBooleanRaidBuffInput = <SpecType extends Spec>(
	config: BooleanInputConfig<RaidBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> => {
	return InputHelpers.makeBooleanIconInput<any, RaidBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getRaid()!.getBuffs(),
			setValue: (player: Player<SpecType>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(newVal),
			storeField: ['raid:buffs', 'race'],
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
};
export const makeBooleanIndividualBuffInput = <SpecType extends Spec>(
	config: BooleanInputConfig<IndividualBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> => {
	return InputHelpers.makeBooleanIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => (!config.faction || config.faction == player.getFaction()) && (!config.showWhen || config.showWhen(player)),
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(newVal),
			storeField: ['buffs', 'race'],
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
};

export const makeBooleanPartyBuffInput = <SpecType extends Spec>(
	config: Omit<BooleanInputConfig<PartyBuffs>, 'showWhen' | 'faction' | 'enableWhen'> & {
		showWhen?: (party: Party) => boolean;
		enableWhen?: (party: Party) => boolean;
	},
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> => {
	return InputHelpers.makeBooleanIconInput<any, PartyBuffs, Party>(
		{
			getModObject: (player: Player<SpecType>) => player.getParty()!,
			showWhen: config.showWhen,
			enableWhen: config.enableWhen,
			getValue: (party: Party) => party.getBuffs(),
			setValue: (party: Party, newVal: PartyBuffs) => party.setBuffs(newVal),
			storeField: ['raid:partyBuffs'],
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
};

export const makeBooleanDebuffInput = <SpecType extends Spec>(
	config: BooleanInputConfig<Debuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, boolean> => {
	return InputHelpers.makeBooleanIconInput<any, Debuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			getValue: (player: Player<SpecType>) => player.getRaid()!.getDebuffs(),
			setValue: (player: Player<SpecType>, newVal: Debuffs) => player.getRaid()!.setDebuffs(newVal),
			storeField: 'raid:debuffs',
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.fieldName,
		config.value,
		config.label,
	);
};

interface TristateInputConfig<T, ModObject = Player<any>> {
	actionId: ActionId;
	impId: ActionId;
	fieldName: keyof T;
	// See BooleanInputConfig: unhonourable off the player, so the party and raid variants omit it.
	faction?: Faction;
	label?: string;
	enableWhen?: (obj: ModObject) => boolean;
}

export const makeTristateRaidBuffInput = <SpecType extends Spec>(
	config: TristateInputConfig<RaidBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeTristateIconInput<any, RaidBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getRaid()!.getBuffs(),
			setValue: (player: Player<SpecType>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(newVal),
			storeField: ['raid:buffs', 'race'],
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
};

export const makeTristateIndividualBuffInput = <SpecType extends Spec>(
	config: TristateInputConfig<IndividualBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeTristateIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(newVal),
			storeField: ['buffs', 'race'],
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
};

export const makeTristatePartyBuffInput = <SpecType extends Spec>(
	config: Omit<TristateInputConfig<PartyBuffs, Party>, 'faction'>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeTristateIconInput<any, PartyBuffs, Party>(
		{
			getModObject: (player: Player<SpecType>) => player.getParty()!,
			getValue: (party: Party) => party.getBuffs(),
			setValue: (party: Party, newVal: PartyBuffs) => party.setBuffs(newVal),
			storeField: 'raid:partyBuffs',
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
};

export const makeTristateDebuffInput = <SpecType extends Spec>(
	config: Omit<TristateInputConfig<Debuffs, Raid>, 'faction'>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeTristateIconInput<any, Debuffs, Raid>(
		{
			getModObject: (player: Player<SpecType>) => player.getRaid()!,
			getValue: (raid: Raid) => raid.getDebuffs(),
			setValue: (raid: Raid, newVal: Debuffs) => raid.setDebuffs(newVal),
			storeField: 'raid:debuffs',
			enableWhen: config.enableWhen,
		},
		config.actionId,
		config.impId,
		config.fieldName,
		config.label,
	);
};

interface MultiStateInputConfig<T> {
	actionId: ActionId;
	label?: string;
	numStates: number;
	fieldName: keyof T;
	multiplier?: number;
	// See BooleanInputConfig: unhonourable off the player, so the party variant omits it.
	faction?: Faction;
}

export const makeMultistateRaidBuffInput = <SpecType extends Spec>(
	config: MultiStateInputConfig<RaidBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeMultistateIconInput<any, RaidBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getRaid()!.getBuffs(),
			setValue: (player: Player<SpecType>, newVal: RaidBuffs) => player.getRaid()!.setBuffs(newVal),
			storeField: ['raid:buffs', 'race'],
		},
		config.actionId,
		config.numStates,
		config.fieldName,
		config.multiplier,
		config.label,
	);
};
export const makeMultistateIndividualBuffInput = <SpecType extends Spec>(
	config: MultiStateInputConfig<IndividualBuffs>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeMultistateIconInput<any, IndividualBuffs, Player<SpecType>>(
		{
			getModObject: (player: Player<SpecType>) => player,
			showWhen: (player: Player<SpecType>) => !config.faction || config.faction == player.getFaction(),
			getValue: (player: Player<SpecType>) => player.getBuffs(),
			setValue: (player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(newVal),
			storeField: ['buffs', 'race'],
		},
		config.actionId,
		config.numStates,
		config.fieldName,
		config.multiplier,
		config.label,
	);
};

export const makeMultistatePartyBuffInput = <SpecType extends Spec>(
	config: Omit<MultiStateInputConfig<PartyBuffs>, 'faction'>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => {
	return InputHelpers.makeMultistateIconInput<any, PartyBuffs, Party>(
		{
			getModObject: (player: Player<SpecType>) => player.getParty()!,
			getValue: (party: Party) => party.getBuffs(),
			setValue: (party: Party, newVal: PartyBuffs) => party.setBuffs(newVal),
			storeField: 'raid:partyBuffs',
		},
		config.actionId,
		config.numStates,
		config.fieldName,
		config.multiplier,
		config.label,
	);
};

interface RankedInputRank {
	actionId: ActionId;
	label: string;
}

interface RankedInputConfig<T> {
	ranks: readonly [RankedInputRank, RankedInputRank];
	fieldName: keyof T;
	label: string;
}

// A field whose number picks the buff's rank: 0 is off, 1 the first rank, which is the icon, and 2 the second, which is
// its improved badge.
const rankedInput = <SpecType extends Spec, Message>(
	make: (config: TristateInputConfig<Message>) => InputHelpers.TypedIconPickerConfig<Player<SpecType>, number>,
	{ ranks, fieldName, label }: RankedInputConfig<Message>,
): InputHelpers.TypedIconPickerConfig<Player<SpecType>, number> => ({
	...make({ actionId: ranks[0].actionId, impId: ranks[1].actionId, fieldName, label }),
	labelTooltip: i18n.t('settings_tab.raid_buffs.rank_select_tooltip', { ranks: ranks.map(rank => rank.label).join(', ') }),
});

export const makeRankedRaidBuffInput = <SpecType extends Spec>(config: RankedInputConfig<RaidBuffs>) =>
	rankedInput<SpecType, RaidBuffs>(makeTristateRaidBuffInput, config);

export const makeRankedIndividualBuffInput = <SpecType extends Spec>(config: RankedInputConfig<IndividualBuffs>) =>
	rankedInput<SpecType, IndividualBuffs>(makeTristateIndividualBuffInput, config);

// interface EnumInputConfig<ModObject, Message, T> {
// 	fieldName: keyof Message
// 	values: Array<IconEnumValueConfig<ModObject, T>>
// 	direction?: IconEnumPickerDirection
// 	numColumns?: number
// 	faction?: Faction
// }

// export function makeEnumIndividualBuffInput<SpecType extends Spec>(config: EnumInputConfig<Player<SpecType>, IndividualBuffs, number>): InputHelpers.TypedIconEnumPickerConfig<Player<SpecType>, number> {
// 	return InputHelpers.makeEnumIconInput<any, IndividualBuffs, Player<SpecType>, number>({
// 		getModObject: (player: Player<SpecType>) => player,
// 		showWhen: (player: Player<SpecType>) =>
// 			(!config.faction || config.faction == player.getFaction()),
// 		getValue: (player: Player<SpecType>) => player.getBuffs(),
// 		setValue: (player: Player<SpecType>, newVal: IndividualBuffs) => player.setBuffs(newVal),
// 	}, config.fieldName, config.values, config.numColumns, config.direction || IconEnumPickerDirection.Vertical)
// };
