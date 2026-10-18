import { HitKind } from '@generated/proto/api';
import { ActionMetrics, SimResult } from '@sim/proto/sim_result';
import { RANKED_SPELLS, type SpellActionFixture, spellActionsRun, stubRankedSpellIcons, stubSpellIcons } from '@testing/results-metrics';
import { beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@i18n/localization', async () => (await import('@testing/sim/localization')).localizationStub);

import type { MetricRow } from './grouping';
import { buildActionRows } from './rank_groups';

const { FIREBALL_1, FIREBALL_2, FROSTBOLT } = RANKED_SPELLS;
const [PENANCE_HEAL_1, PENANCE_HEAL_2, PENANCE_1, PENANCE_2] = [1001, 1002, 1003, 1004];

const groupsOf = async (spells: Array<SpellActionFixture>) => {
	const { request, result } = spellActionsRun(spells);
	const player = (await SimResult.makeNew(request, result)).getFirstPlayer()!;
	return ActionMetrics.groupById(player.getDamageActions());
};

interface RowShape {
	spellId: number;
	rank: number;
	tag: number;
	subRows?: Array<RowShape>;
}

const shape = (rows: Array<MetricRow<ActionMetrics>>): Array<RowShape> =>
	rows.map(({ metric: { actionId }, subRows }) => ({
		spellId: actionId.spellId,
		rank: actionId.rank,
		tag: actionId.tag,
		...(subRows && { subRows: shape(subRows) }),
	}));

describe('buildActionRows', () => {
	beforeEach(() => {
		stubRankedSpellIcons();
	});

	it('puts the ranks of a spell under one row, and the tags of a rank under that rank', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FIREBALL_1, tag: 1, damage: 10 },
			{ spellId: FROSTBOLT, damage: 50 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);
		const rows = buildActionRows(groups, true);

		expect(shape(rows)).toEqual([
			{
				spellId: FIREBALL_2,
				rank: 0,
				tag: 0,
				subRows: [
					{
						spellId: FIREBALL_1,
						rank: 1,
						tag: 0,
						subRows: [
							{ spellId: FIREBALL_1, rank: 1, tag: 0 },
							{ spellId: FIREBALL_1, rank: 1, tag: 1 },
						],
					},
					{
						spellId: FIREBALL_2,
						rank: 2,
						tag: 0,
						subRows: [
							{ spellId: FIREBALL_2, rank: 2, tag: 0 },
							{ spellId: FIREBALL_2, rank: 2, tag: 1 },
						],
					},
				],
			},
			{ spellId: FROSTBOLT, rank: 0, tag: 0 },
		]);
		expect(rows[0].subRows![0].subRows![0].metric).toBe(groups[0][0]);
		expect(rows[0].subRows![1].subRows![1].metric).toBe(groups[2][1]);
		expect(rows[1].metric).toBe(groups[1][0]);
	});

	it('names the row of joined ranks after the highest rank without its rank, and merges the counts and ranges of every rank', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FIREBALL_1, tag: 1, damage: 10 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);
		const members = groups.flat();
		const [{ metric: parent, subRows }] = buildActionRows(groups, true);

		expect(parent.name).toBe('Fireball');
		expect(parent.actionId.spellId).toBe(FIREBALL_2);
		expect(parent.actionId.hasRank).toBe(false);
		expect(parent.damage).toBe(440);
		expect(subRows![1].metric.damage).toBe(330);
		expect(parent.hits).toBeCloseTo(members.reduce((sum, member) => sum + member.hits, 0));
		expect(parent.range(HitKind.HitKindHit)!.aggregatorData!.n).toBe(4 * members[0].range(HitKind.HitKindHit)!.aggregatorData!.n);
	});

	it('builds a row per spell id, ranks apart, while ranks are not grouped', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);
		const rows = buildActionRows(groups, false);

		expect(shape(rows)).toEqual([
			{ spellId: FIREBALL_1, rank: 1, tag: 0 },
			{
				spellId: FIREBALL_2,
				rank: 2,
				tag: 0,
				subRows: [
					{ spellId: FIREBALL_2, rank: 2, tag: 0 },
					{ spellId: FIREBALL_2, rank: 2, tag: 1 },
				],
			},
		]);
		expect(rows[1].metric.actionId.equals(groups[1][0].actionId.withoutTag())).toBe(true);
	});

	it('leaves a spell with one rank, and a group holding other spells at other ranks, as they are', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FROSTBOLT, damage: 50 },
		]);
		const [rider] = await groupsOf([{ spellId: FIREBALL_1, damage: 100 }]);
		const withRider = [[...groups[0], ...rider]];

		for (const sample of [groups, withRider]) {
			expect(shape(buildActionRows(sample, true))).toEqual(shape(buildActionRows(sample, false)));
		}
		expect(buildActionRows(withRider, true)[0].metric.actionId.rank).toBe(2);
	});

	it('keeps an unranked spell apart from the ranks of a spell of the same name', async () => {
		stubSpellIcons({
			[FIREBALL_1]: { name: 'Fireball', rank: 1 },
			[FIREBALL_2]: { name: 'Fireball', rank: 2 },
			[FROSTBOLT]: { name: 'Fireball' },
		});
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FROSTBOLT, damage: 50 },
			{ spellId: FIREBALL_2, damage: 300 },
		]);

		expect(shape(buildActionRows(groups, true))).toEqual([
			{
				spellId: FIREBALL_2,
				rank: 0,
				tag: 0,
				subRows: [
					{ spellId: FIREBALL_1, rank: 1, tag: 0 },
					{ spellId: FIREBALL_2, rank: 2, tag: 0 },
				],
			},
			{ spellId: FROSTBOLT, rank: 0, tag: 0 },
		]);
	});

	it('joins a rank that landed no tag 1 values with a rank that did', async () => {
		const groups = await groupsOf([
			{ spellId: FIREBALL_1, damage: 100 },
			{ spellId: FIREBALL_2, damage: 300 },
			{ spellId: FIREBALL_2, tag: 1, damage: 30 },
		]);
		const rows = buildActionRows(groups, true);

		expect(shape(rows)).toEqual([
			{
				spellId: FIREBALL_2,
				rank: 0,
				tag: 0,
				subRows: [
					{ spellId: FIREBALL_1, rank: 1, tag: 0 },
					{
						spellId: FIREBALL_2,
						rank: 2,
						tag: 0,
						subRows: [
							{ spellId: FIREBALL_2, rank: 2, tag: 0 },
							{ spellId: FIREBALL_2, rank: 2, tag: 1 },
						],
					},
				],
			},
		]);
		expect(rows[0].metric.name).toBe('Fireball');
		expect(rows[0].subRows![0].metric).toBe(groups[0][0]);
	});

	it('keeps apart two spells of one name that differ by tag, and names each after its tag', async () => {
		stubSpellIcons({
			[PENANCE_HEAL_1]: { name: 'Penance', rank: 1 },
			[PENANCE_HEAL_2]: { name: 'Penance', rank: 2 },
			[PENANCE_1]: { name: 'Penance', rank: 1 },
			[PENANCE_2]: { name: 'Penance', rank: 2 },
		});
		const groups = await groupsOf([
			{ spellId: PENANCE_HEAL_1, tag: 1, damage: 100 },
			{ spellId: PENANCE_1, damage: 100 },
			{ spellId: PENANCE_HEAL_2, tag: 1, damage: 200 },
			{ spellId: PENANCE_2, damage: 200 },
		]);
		const rows = buildActionRows(groups, true);

		expect(shape(rows)).toEqual([
			{
				spellId: PENANCE_HEAL_2,
				rank: 0,
				tag: 1,
				subRows: [
					{ spellId: PENANCE_HEAL_1, rank: 1, tag: 1 },
					{ spellId: PENANCE_HEAL_2, rank: 2, tag: 1 },
				],
			},
			{
				spellId: PENANCE_2,
				rank: 0,
				tag: 0,
				subRows: [
					{ spellId: PENANCE_1, rank: 1, tag: 0 },
					{ spellId: PENANCE_2, rank: 2, tag: 0 },
				],
			},
		]);
		expect(rows.map(row => row.metric.name)).toEqual(['Penance (Heal)', 'Penance']);
		expect(rows[0].subRows![1].metric).toBe(groups[2][0]);
	});
});
