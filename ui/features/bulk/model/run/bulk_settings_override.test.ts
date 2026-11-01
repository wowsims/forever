import { applySavedSettings } from '@features/settings/model/saved_settings';
import { BulkSettings, type BulkSimRequest, type Raid as RaidProto } from '@generated/proto/api';
import { Debuffs, IndividualBuffs, PartyBuffs, RaidBuffs } from '@generated/proto/buffs';
import {
	ConsumableType,
	ConsumesSpec,
	Encounter as EncounterProto,
	EquipmentSpec,
	HealingModel,
	ItemSlot,
	ItemSpec,
	ItemSwap,
	ItemType,
	Profession,
	Race,
	Stat,
	Target,
} from '@generated/proto/common';
import { Consumable } from '@generated/proto/db';
import { Mage } from '@generated/proto/mage';
import { SavedSettings, UIDatabase, UIEnchant, UIItem } from '@generated/proto/ui';
import { DpsWarrior, DpsWarrior_Options } from '@generated/proto/warrior';
import type { BulkSimOverrides } from '@sim/bulk/types';
import { DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID } from '@sim/proto/sim_tables_auto_gen';
import { Stats } from '@sim/proto/stats';
import type { Sim } from '@sim/sim';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { MACE, swapSet, SWORD, weaponsDatabase } from '@testing/sim/weapons';
import { makeEquipmentSpec } from '@testing/storybook/fixtures/protos';
import { beforeAll, describe, expect, it, vi } from 'vitest';

const sent = vi.hoisted(() => [] as BulkSimRequest[]);

vi.mock('@sim/workers/worker_pool', async () => {
	const idle = (await import('@testing/sim/worker_pool')).idleWorkerPool();
	return {
		...idle,
		WorkerPool: class extends idle.WorkerPool {
			bulkSimAsync = (request: BulkSimRequest) => {
				sent.push(request);
				return Promise.resolve({});
			};
		},
	};
});

const scalingOptions = { 0: { ilvl: 60 } };
const RING = UIItem.create({ id: 40, name: 'Band', type: ItemType.ItemTypeFinger, scalingOptions });
const RING_ENCHANT = UIEnchant.create({ effectId: 2930, itemId: 4, name: 'Enchant Ring - Striking', type: ItemType.ItemTypeFinger });
const SWAP_HELM = UIItem.create({ id: 50, name: 'Swap helm', type: ItemType.ItemTypeHead, scalingOptions });
const DARKSPEAR_ITEM = UIItem.create({ id: 12802, name: 'Darkspear item', type: ItemType.ItemTypeTrinket, scalingOptions });
const ORC_POTION = Consumable.create({ id: 60, name: 'Orc draught', type: ConsumableType.ConsumableTypePotion, raceAllowlist: [Race.RaceOrc] });
const ENGINEER_POTION = Consumable.create({
	id: 61,
	name: 'Engineer tonic',
	type: ConsumableType.ConsumableTypePotion,
	requiredProfession: Profession.Engineering,
});
const PLAIN_POTION = Consumable.create({ id: 62, name: 'Plain potion', type: ConsumableType.ConsumableTypePotion });
const FLASK = Consumable.create({ id: 63, name: 'Flask', type: ConsumableType.ConsumableTypeFlask });

const ENTRY = SavedSettings.create({
	raidBuffs: RaidBuffs.create({ arcaneBrilliance: true }),
	partyBuffs: PartyBuffs.create({ darkspear: true, bloodPact: true }),
	playerBuffs: IndividualBuffs.create({ greaterBlessingOfKings: true }),
	debuffs: Debuffs.create({ huntersMark: true }),
	consumables: ConsumesSpec.create({ flaskId: FLASK.id }),
	race: Race.RaceOrc,
	professions: [Profession.Enchanting, Profession.Engineering],
	enableItemSwap: true,
	itemSwap: ItemSwap.create({ items: [ItemSpec.create({ id: SWAP_HELM.id })], prepullBonusStats: new Stats().withStat(Stat.StatStrength, 10).toProto() }),
	reactionTimeMs: 250,
	channelClipDelayMs: 50,
	inFrontOfTarget: true,
	distanceFromTarget: 20,
	healingModel: HealingModel.create({ hps: 500 }),
	spec: { oneofKind: 'dpsWarrior', dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage: 40 } } }) },
});

beforeAll(() =>
	serveDatabase(
		UIDatabase.create({
			items: [RING, SWAP_HELM, DARKSPEAR_ITEM, ...weaponsDatabase().items],
			enchants: [RING_ENCHANT],
			consumables: [ORC_POTION, ENGINEER_POTION, PLAIN_POTION, FLASK],
		}),
	),
);

const setup = async () => {
	const built = await makeIndividualSim();
	built.player.setRace(Race.RaceHuman);
	return built;
};

const run = async (sim: Sim, overrides: BulkSimOverrides = {}) => {
	sent.length = 0;
	await sim.runBulkSim(() => {}, BulkSettings.create(), undefined, overrides);
	return sent[0].baseRequest!.raid!;
};

// Every field applySavedSettings writes on the raid, the party and the player; one it gains belongs here too.
const settingsFields = (raid: RaidProto) => {
	const player = raid.parties[0].players[0];
	return {
		raidBuffs: raid.buffs,
		debuffs: raid.debuffs,
		partyBuffs: raid.parties[0].buffs,
		race: player.race,
		profession1: player.profession1,
		profession2: player.profession2,
		prof1Rank: player.prof1Rank,
		prof2Rank: player.prof2Rank,
		consumables: player.consumables,
		buffs: player.buffs,
		enableItemSwap: player.enableItemSwap,
		itemSwap: player.itemSwap,
		reactionTimeMs: player.reactionTimeMs,
		channelClipDelayMs: player.channelClipDelayMs,
		inFrontOfTarget: player.inFrontOfTarget,
		distanceFromTarget: player.distanceFromTarget,
		healingModel: player.healingModel,
		spec: player.spec,
	};
};

const ringedGear = (sim: Sim) =>
	sim.db.lookupEquipmentSpec(EquipmentSpec.create({ items: [ItemSpec.create({ id: RING.id, enchant: RING_ENCHANT.effectId })] }));

describe('Sim.runBulkSim with a saved settings entry', () => {
	it('sends the raid, party and player the Settings tab would load from the entry', async () => {
		const { sim, host } = await setup();

		const sentFields = settingsFields(await run(sim, { settings: ENTRY }));
		applySavedSettings(host, ENTRY);

		expect(sentFields).toEqual(settingsFields(sim.makeRaidSimRequest().raid!));
		expect(sentFields.race).toBe(Race.RaceOrc);
		expect(sentFields.consumables!.potions).toEqual([ORC_POTION.id, ENGINEER_POTION.id, PLAIN_POTION.id]);
		expect(sentFields.itemSwap!.items[0].id).toBe(SWAP_HELM.id);
		expect(sentFields.spec).toEqual(ENTRY.spec);
	});

	it('sends the live spec options for an entry saved without them, or with another spec’s', async () => {
		const { sim, player } = await setup();
		player.setSpecOptions(DpsWarrior_Options.create({ classOptions: { startingRage: 20 } }));
		const live = { oneofKind: 'dpsWarrior', dpsWarrior: DpsWarrior.create({ options: { classOptions: { startingRage: 20 } } }) };
		const mageEntry = SavedSettings.create({ spec: { oneofKind: 'mage', mage: Mage.create({ options: {} }) } });

		expect((await run(sim, { settings: SavedSettings.create() })).parties[0].players[0].spec).toEqual(live);
		expect((await run(sim, { settings: mageEntry })).parties[0].players[0].spec).toEqual(live);
	});

	it('keeps the live rank of a profession the entry keeps and drops the one it changes', async () => {
		const { sim, player } = await setup();
		player.setProfessions([Profession.Enchanting, Profession.Tailoring]);
		player.setProfessionRanks(300, 150);

		const sentPlayer = (await run(sim, { settings: ENTRY })).parties[0].players[0];

		expect([sentPlayer.profession1, sentPlayer.prof1Rank]).toEqual([Profession.Enchanting, 300]);
		expect([sentPlayer.profession2, sentPlayer.prof2Rank]).toEqual([Profession.Engineering, 0]);
	});

	it('leaves the live raid, party and player as they are', async () => {
		const { sim, player } = await setup();
		const playerBefore = player.toProto();
		const raidBefore = sim.raid.toProto();

		await run(sim, { settings: ENTRY });

		expect(player.toProto()).toEqual(playerBefore);
		expect(sim.raid.toProto()).toEqual(raidBefore);
	});

	it('strips ring enchants by the entry’s professions, not the live ones', async () => {
		const sentRingEnchant = async (live: Profession[], entry: Profession[]) => {
			const { sim, player } = await setup();
			player.setProfessions(live);
			player.setGear(ringedGear(sim));
			const sentPlayer = (await run(sim, { settings: SavedSettings.create({ professions: entry }) })).parties[0].players[0];
			return sentPlayer.equipment!.items.find(item => item.id === RING.id)!.enchant;
		};

		expect(await sentRingEnchant([Profession.Enchanting], [])).toBeFalsy();
		expect(await sentRingEnchant([], [Profession.Enchanting])).toBe(RING_ENCHANT.effectId);
	});

	it('keeps the entry’s hand stones for the sim to match to each candidate’s weapons, and matches its swap stones to its swap weapons', async () => {
		const { sim, player } = await setup();
		player.setGear(sim.db.lookupEquipmentSpec(makeEquipmentSpec({ [ItemSlot.ItemSlotMainHand]: { id: SWORD } })));
		const dualWield = sim.db.lookupEquipmentSpec(
			makeEquipmentSpec({ [ItemSlot.ItemSlotMainHand]: { id: SWORD }, [ItemSlot.ItemSlotOffHand]: { id: MACE } }),
		);
		const entry = SavedSettings.create({
			itemSwap: swapSet({ mainHand: MACE }),
			consumables: ConsumesSpec.create({
				mhImbueId: DENSE_SHARPENING_STONE_ID,
				ohImbueId: DENSE_WEIGHTSTONE_ID,
				swapMhImbueId: DENSE_SHARPENING_STONE_ID,
			}),
		});

		const { mhImbueId, ohImbueId, swapMhImbueId } = (await run(sim, { settings: entry, baselineGear: dualWield })).parties[0].players[0].consumables!;

		expect([mhImbueId, ohImbueId, swapMhImbueId]).toEqual([DENSE_SHARPENING_STONE_ID, DENSE_WEIGHTSTONE_ID, DENSE_WEIGHTSTONE_ID]);
	});

	it('turns item swap on with the swap gear and ships it, the party-buff items and the consumables the entry adds', async () => {
		const { sim, player } = await setup();

		const sentPlayer = (await run(sim, { settings: ENTRY })).parties[0].players[0];
		const database = sentPlayer.database!;

		expect(player.itemSwapSettings.getEnableItemSwap()).toBe(false);
		expect(sentPlayer.enableItemSwap).toBe(true);
		expect(sentPlayer.itemSwap!.items[0].id).toBe(SWAP_HELM.id);
		expect(database.items.map(item => item.id)).toEqual(expect.arrayContaining([SWAP_HELM.id, DARKSPEAR_ITEM.id]));
		expect(database.consumables.map(consumable => consumable.id)).toEqual(expect.arrayContaining([FLASK.id]));
	});

	it('fills a healing spec’s zero healing model from the pair’s boss, as setHealingModel does', async () => {
		const { sim, player, host } = await setup();
		player.enableHealing();
		const boss = Target.create({ swingSpeed: 2, minBaseDamage: 4000 });
		const entry = SavedSettings.create({ healingModel: HealingModel.create() });

		const sentHealingModel = (await run(sim, { encounter: EncounterProto.create({ targets: [boss] }), settings: entry })).parties[0].players[0]
			.healingModel;
		sim.encounter.setTargets([boss]);
		applySavedSettings(host, entry);

		expect(sentHealingModel).toEqual(player.getHealingModel());
		expect(sentHealingModel!.hps).toBe(500);
	});

	it('sends the live settings when no entry is given', async () => {
		const { sim, player } = await setup();
		player.setProfessions([Profession.Tailoring]);
		player.setBuffs(IndividualBuffs.create({ greaterBlessingOfMight: true }));
		player.setConsumes(ConsumesSpec.create({ flaskId: FLASK.id }));

		const sentPlayer = (await run(sim)).parties[0].players[0];

		expect(sentPlayer.race).toBe(Race.RaceHuman);
		expect([sentPlayer.profession1, sentPlayer.profession2]).toEqual([Profession.Tailoring, Profession.ProfessionUnknown]);
		expect(sentPlayer.buffs).toEqual(IndividualBuffs.create({ greaterBlessingOfMight: true }));
		expect(sentPlayer.consumables).toEqual(ConsumesSpec.create({ flaskId: FLASK.id, potions: [PLAIN_POTION.id] }));
	});

	it('draws the same content seed with and without a settings override', async () => {
		const { sim } = await setup();

		await run(sim);
		const plain = sim.getLastUsedRngSeed();
		await run(sim, { settings: ENTRY });

		expect(sim.getLastUsedRngSeed()).toBe(plain);
	});
});
