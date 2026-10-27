import { UnitReference } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Player } from '@sim/player/player';
import { emptyUnitReference } from '@sim/proto/utils';

const INPUT_DELAY_FIELDS = ['reactionTime', 'channelClipDelay'] as const;
// The healing-model inputs below are enabled only while the player is one of the
// raid's tanks, so they have to re-evaluate on a tank-assignment change as well as
// on their own value.
const HEALING_MODEL_FIELDS = ['healingModel', 'raid:tanks'] as const;

export const InputDelay = {
	id: 'input-delay',
	type: 'number' as const,
	label: i18n.t('settings_tab.other.input_delay.label'),
	labelTooltip: i18n.t('settings_tab.other.input_delay.tooltip'),
	storeField: INPUT_DELAY_FIELDS,
	getValue: (player: Player<any>) => player.getReactionTime(),
	setValue: (player: Player<any>, newValue: number) => {
		player.setReactionTime(newValue);
	},
};

export const ChannelClipDelay = {
	id: 'channel-clip-delay',
	type: 'number' as const,
	label: i18n.t('settings_tab.other.channel_clip_delay.label'),
	labelTooltip: i18n.t('settings_tab.other.channel_clip_delay.tooltip'),
	storeField: INPUT_DELAY_FIELDS,
	getValue: (player: Player<any>) => player.getChannelClipDelay(),
	setValue: (player: Player<any>, newValue: number) => {
		player.setChannelClipDelay(newValue);
	},
};

export const InFrontOfTarget = {
	id: 'in-front-of-target',
	type: 'boolean' as const,
	label: i18n.t('settings_tab.other.in_front_of_target.label'),
	labelTooltip: i18n.t('settings_tab.other.in_front_of_target.tooltip'),
	storeField: 'inFrontOfTarget' as const,
	getValue: (player: Player<any>) => player.getInFrontOfTarget(),
	setValue: (player: Player<any>, newValue: boolean) => {
		player.setInFrontOfTarget(newValue);
	},
};

export const DistanceFromTarget = {
	id: 'distance-from-target',
	type: 'number' as const,
	float: true,
	maxDecimalDigits: 2,
	positive: true,
	label: i18n.t('settings_tab.other.distance_from_target.label'),
	labelTooltip: i18n.t('settings_tab.other.distance_from_target.tooltip'),
	storeField: 'distanceFromTarget' as const,
	getValue: (player: Player<any>) => player.getDistanceFromTarget(),
	setValue: (player: Player<any>, newValue: number) => {
		player.setDistanceFromTarget(newValue);
	},
};

export const TankAssignment = {
	id: 'tank-assignment',
	type: 'enum' as const,
	label: i18n.t('settings_tab.other.tank_assignment.label'),
	labelTooltip: i18n.t('settings_tab.other.tank_assignment.tooltip'),
	values: [
		{ name: i18n.t('common.none'), value: -1 },
		{ name: i18n.t('common.tanks.main_tank'), value: 0 },
		{ name: i18n.t('common.tanks.tank_2'), value: 1 },
		{ name: i18n.t('common.tanks.tank_3'), value: 2 },
		{ name: i18n.t('common.tanks.tank_4'), value: 3 },
	],
	storeField: ['raid:tanks', 'ui:showThreatMetrics'] as const,
	showWhen: (player: Player<any>) => player.sim.getShowThreatMetrics(),
	getValue: (player: Player<any>) => (player.getRaid()?.getTanks() || []).findIndex(tank => UnitReference.equals(tank, player.makeUnitReference())),
	setValue: (player: Player<any>, newValue: number) => {
		const newTanks = [];
		if (newValue != -1) {
			for (let i = 0; i < newValue; i++) {
				newTanks.push(emptyUnitReference());
			}
			newTanks.push(player.makeUnitReference());
		}
		player.getRaid()!.setTanks(newTanks);
	},
};

export const IncomingHps = {
	id: 'incoming-hps',
	type: 'number' as const,
	label: i18n.t('settings_tab.other.incoming_hps.label'),
	labelTooltip: i18n.t('settings_tab.other.incoming_hps.tooltip'),
	storeField: HEALING_MODEL_FIELDS,
	getValue: (player: Player<any>) => player.getHealingModel().hps,
	setValue: (player: Player<any>, newValue: number) => {
		const healingModel = player.getHealingModel();
		healingModel.hps = newValue;
		player.setHealingModel(healingModel);
	},
	enableWhen: (player: Player<any>) => (player.getRaid()?.getTanks() || []).find(tank => UnitReference.equals(tank, player.makeUnitReference())) != null,
};

export const HealingCadence = {
	id: 'healing-cadence',
	type: 'number' as const,
	float: true,
	label: i18n.t('settings_tab.other.healing_cadence.label'),
	labelTooltip: i18n.t('settings_tab.other.healing_cadence.tooltip'),
	storeField: HEALING_MODEL_FIELDS,
	getValue: (player: Player<any>) => player.getHealingModel().cadenceSeconds,
	setValue: (player: Player<any>, newValue: number) => {
		const healingModel = player.getHealingModel();
		healingModel.cadenceSeconds = newValue;
		player.setHealingModel(healingModel);
	},
	enableWhen: (player: Player<any>) => (player.getRaid()?.getTanks() || []).find(tank => UnitReference.equals(tank, player.makeUnitReference())) != null,
};

export const HealingCadenceVariation = {
	id: 'healing-cadence-variation',
	type: 'number' as const,
	float: true,
	label: i18n.t('settings_tab.other.healing_cadence_variation.label'),
	labelTooltip: i18n.t('settings_tab.other.healing_cadence_variation.tooltip'),
	storeField: HEALING_MODEL_FIELDS,
	getValue: (player: Player<any>) => player.getHealingModel().cadenceVariation,
	setValue: (player: Player<any>, newValue: number) => {
		const healingModel = player.getHealingModel();
		healingModel.cadenceVariation = newValue;
		player.setHealingModel(healingModel);
	},
	enableWhen: (player: Player<any>) => (player.getRaid()?.getTanks() || []).find(tank => UnitReference.equals(tank, player.makeUnitReference())) != null,
};

export const AbsorbFrac = {
	id: 'healing-model-absorb-frac',
	type: 'number' as const,
	float: true,
	label: i18n.t('settings_tab.other.absorb_frac.label'),
	labelTooltip: i18n.t('settings_tab.other.absorb_frac.tooltip'),
	storeField: 'healingModel' as const,
	getValue: (player: Player<any>) => player.getHealingModel().absorbFrac * 100,
	setValue: (player: Player<any>, newValue: number) => {
		const healingModel = player.getHealingModel();
		healingModel.absorbFrac = newValue / 100;
		player.setHealingModel(healingModel);
	},
};

export const BurstWindow = {
	id: 'burst-window',
	type: 'number' as const,
	float: false,
	label: i18n.t('settings_tab.other.burst_window.label'),
	labelTooltip: i18n.t('settings_tab.other.burst_window.tooltip'),
	storeField: HEALING_MODEL_FIELDS,
	getValue: (player: Player<any>) => player.getHealingModel().burstWindow,
	setValue: (player: Player<any>, newValue: number) => {
		const healingModel = player.getHealingModel();
		healingModel.burstWindow = newValue;
		player.setHealingModel(healingModel);
	},
	enableWhen: (player: Player<any>) => (player.getRaid()?.getTanks() || []).find(tank => UnitReference.equals(tank, player.makeUnitReference())) != null,
};

export const InspirationUptime = {
	id: 'inspiration-uptime',
	type: 'number' as const,
	float: true,
	label: i18n.t('settings_tab.other.inspiration_uptime.label'),
	labelTooltip: i18n.t('settings_tab.other.inspiration_uptime.tooltip'),
	storeField: HEALING_MODEL_FIELDS,
	getValue: (player: Player<any>) => player.getHealingModel().inspirationUptime * 100,
	setValue: (player: Player<any>, newValue: number) => {
		const healingModel = player.getHealingModel();
		healingModel.inspirationUptime = newValue / 100;
		player.setHealingModel(healingModel);
	},
	enableWhen: (player: Player<any>) => (player.getRaid()?.getTanks() || []).find(tank => UnitReference.equals(tank, player.makeUnitReference())) != null,
};

export const HpPercentForDefensives = {
	id: 'hp-percent-for-defensives',
	type: 'number' as const,
	float: true,
	label: i18n.t('settings_tab.other.hp_percent_for_defensives.label'),
	labelTooltip: i18n.t('settings_tab.other.hp_percent_for_defensives.tooltip'),
	storeField: 'rotation' as const,
	getValue: (player: Player<any>) => player.getSimpleCooldowns().hpPercentForDefensives * 100,
	setValue: (player: Player<any>, newValue: number) => {
		const cooldowns = player.getSimpleCooldowns();
		cooldowns.hpPercentForDefensives = newValue / 100;
		player.setSimpleCooldowns(cooldowns);
	},
};

// The Holy spell power of the paladin providing Retribution Aura as a party buff (sim/core's
// NewSpellDataRetributionAuraRank turns it into damage per hit). Only a tank takes the hits, and the number
// means nothing without the aura, so it shows under both conditions.
export const RetributionAuraSpellPower = {
	id: 'retributionAuraSpellPower',
	type: 'number' as const,
	label: i18n.t('settings_tab.other.retribution_aura_spell_power.label'),
	labelTooltip: i18n.t('settings_tab.other.retribution_aura_spell_power.tooltip'),
	storeField: 'raid:partyBuffs' as const,
	showWhen: (player: Player<any>) => player.getPlayerSpec().isTankSpec && !!player.getParty()?.getBuffs().retributionAura,
	getValue: (player: Player<any>) => player.getParty()!.getBuffs().retributionAuraSpellPower,
	setValue: (player: Player<any>, newValue: number) => {
		const party = player.getParty()!;
		const buffs = party.getBuffs();
		buffs.retributionAuraSpellPower = newValue;
		party.setBuffs(buffs);
	},
};

// The Nature spell power of the druid providing Thorns as an individual buff (sim/core's
// NewSpellDataThornsAuraWithSpellPower turns it into damage per hit). Only a tank takes the hits, and the
// number means nothing without Thorns, so it shows under both conditions.
export const ThornsSpellPower = {
	id: 'thornsSpellPower',
	type: 'number' as const,
	label: i18n.t('settings_tab.other.thorns_spell_power.label'),
	labelTooltip: i18n.t('settings_tab.other.thorns_spell_power.tooltip'),
	storeField: 'buffs' as const,
	showWhen: (player: Player<any>) => player.getPlayerSpec().isTankSpec && player.getBuffs().thorns,
	getValue: (player: Player<any>) => player.getBuffs().thornsSpellPower,
	setValue: (player: Player<any>, newValue: number) => {
		const buffs = player.getBuffs();
		buffs.thornsSpellPower = newValue;
		player.setBuffs(buffs);
	},
};
