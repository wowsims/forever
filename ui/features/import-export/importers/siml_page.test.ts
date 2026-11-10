// What a SIML import leaves on a real page: what the text writes, and the page's own value for everything it leaves
// out. The sim's reading is tested in Go (sim/siml/api_test.go); here it is the JSON `wowsimcli siml serve` gives.
import { APLRotation } from '@generated/proto/apl';
import { Cooldowns, Faction, Profession, Race, Spec } from '@generated/proto/common';
import { IndividualSimSettings, SimlParseResult } from '@generated/proto/ui';
import type { JsonObject } from '@protobuf-ts/runtime';
import { makeIndividualSim, serveDatabase } from '@testing/sim/individual_sim';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import { SIML_IMPORTER } from './siml';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const parseSiml = vi.hoisted(() => vi.fn());
vi.mock('../siml_client', async importOriginal => ({ ...(await importOriginal<typeof import('../siml_client')>()), parseSiml }));

// The sim reads `json` from the text, holding the sections of `categories`.
const reads = (json: JsonObject, categories: string[]) =>
	parseSiml.mockResolvedValue(SimlParseResult.create({ settings: IndividualSimSettings.fromJson(json), categories, spec: Spec.SpecHunter }));

const ROTATION = APLRotation.fromJson({ type: 'TypeAPL', priorityList: [{ action: { castSpell: { spellId: { spellId: 2643 } } } }] });

// A hunter page that has its own of everything a test reads.
const page = async () => {
	const built = await makeIndividualSim({ spec: Spec.SpecHunter });
	const { player, sim } = built;
	player.setRace(Race.RaceOrc);
	player.setReactionTime(150);
	player.setProfession1(Profession.Engineering);
	player.setProfession2(Profession.Blacksmithing);
	player.setTalentsString('-0053');
	player.setAplRotation(ROTATION);
	player.setSimpleCooldowns(Cooldowns.fromJson({ hpPercentForDefensives: 0.2, cooldowns: [{ id: { spellId: 26296 }, timings: [5] }] }));
	player.setConsumes({ ...player.getConsumes(), flaskId: 13510 });
	player.setEpRatios([1, 0.5, 0]);
	sim.setFaction(Faction.Horde);
	sim.setShowDamageMetrics(true);
	return built;
};

beforeAll(() => serveDatabase());

describe('a cooldowns: section', () => {
	it('sets the cooldowns it writes and leaves the rotation as it is', async () => {
		const { player, host } = await page();
		reads({ player: { class: 'ClassHunter', hunter: {}, cooldowns: { hpPercentForDefensives: 0.4 } } }, ['r']);

		await SIML_IMPORTER.onImport(host, 'the text');

		expect(player.getSimpleCooldowns().hpPercentForDefensives).toBe(0.4);
		expect(player.getSimpleCooldowns().cooldowns).toHaveLength(1);
		expect(APLRotation.equals(player.aplRotation, { ...ROTATION, simple: player.aplRotation.simple })).toBe(true);
		expect(player.getTalentsString()).toBe('-0053');
	});

	it('sets the cooldowns of the rotation the text holds beside it', async () => {
		const { player, host } = await page();
		reads({ player: { class: 'ClassHunter', hunter: {}, rotation: { type: 'TypeAPL' }, cooldowns: { hpPercentForDefensives: 0.4 } } }, ['r']);

		await SIML_IMPORTER.onImport(host, 'the text');

		expect(player.aplRotation.priorityList).toHaveLength(0);
		expect(player.getSimpleCooldowns().hpPercentForDefensives).toBe(0.4);
		expect(player.getSimpleCooldowns().cooldowns).toHaveLength(1);
	});

	it("keeps the page's cooldowns through a rotation without one", async () => {
		const { player, host } = await page();
		reads({ player: { class: 'ClassHunter', hunter: {}, rotation: { type: 'TypeAPL' } } }, ['r']);

		await SIML_IMPORTER.onImport(host, 'the text');

		expect(player.aplRotation.priorityList).toHaveLength(0);
		expect(player.getSimpleCooldowns().hpPercentForDefensives).toBe(0.2);
		expect(player.getSimpleCooldowns().cooldowns).toHaveLength(1);
	});
});

describe('a section that leaves fields out', () => {
	// builds/bm.siml without its use lines: distance and pet, but no race, reaction time or professions.
	it("keeps the page's race, reaction time and professions through a settings: section without them", async () => {
		const { player, host } = await page();
		reads({ player: { class: 'ClassHunter', distanceFromTarget: 12, hunter: { options: { classOptions: { petUptime: 1 } } } } }, ['m']);

		await SIML_IMPORTER.onImport(host, 'the text');

		expect(player.getDistanceFromTarget()).toBe(12);
		expect(player.getSpecOptions()).toMatchObject({ classOptions: { petUptime: 1 } });
		expect(player.getRace()).toBe(Race.RaceOrc);
		expect(player.getReactionTime()).toBe(150);
		expect(player.getProfession1()).toBe(Profession.Engineering);
		expect(player.getProfession2()).toBe(Profession.Blacksmithing);
	});

	it("keeps the page's shown metrics, faction and EP ratios through sim: iterations 10000", async () => {
		const { player, sim, host } = await page();
		reads({ settings: { iterations: 10000 }, player: { class: 'ClassHunter', hunter: {} } }, ['u']);

		await SIML_IMPORTER.onImport(host, 'the text');

		expect(sim.getIterations()).toBe(10000);
		expect(sim.getShowDamageMetrics()).toBe(true);
		expect(sim.getFaction()).toBe(Faction.Horde);
		expect(player.getEpRatios().slice(0, 3)).toEqual([1, 0.5, 0]);
	});

	it('empties a category whose section the text writes empty', async () => {
		const { player, host } = await page();
		reads({ player: { class: 'ClassHunter', hunter: {}, consumables: {} } }, ['c']);

		await SIML_IMPORTER.onImport(host, 'the text');

		expect(player.getConsumes().flaskId).toBe(0);
	});
});
