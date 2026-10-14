import { ResourceType } from '@generated/proto/spell';
import { describe, expect, it } from 'vitest';

import { loadAttackRows, loadRageResult, loadTopline, resultsHostParts, warriorRotationModel } from './results-views';

describe('results view fixtures', () => {
	it('builds the topline and its deltas against the weaker reference without the item database', async () => {
		const { metrics, referenceDiffs } = await loadTopline();

		expect(metrics.map(metric => metric.metric)).toContain('dps');
		expect(referenceDiffs.dps?.tone).toBe('positive');
	});

	it('merges the two hands into one attack parent', async () => {
		const { rows, maxDamage } = await loadAttackRows();

		expect(rows).toHaveLength(1);
		expect(rows[0].subRows!.map(row => row.metric.name)).toEqual([
			'common.action_names.attack (common.action_tags.main_hand)',
			'common.action_names.attack (common.action_tags.off_hand)',
		]);
		expect(rows[0].metric.damage).toBeCloseTo(
			rows[0].subRows!.reduce((total, row) => total + row.metric.damage, 0),
			6,
		);
		expect(maxDamage).toBe(rows[0].metric.damage);
		expect(rows[0].metric.totalDamagePercent).toBeCloseTo(100, 6);
	});

	it('carries two rage sources and nothing else', async () => {
		const { resultData } = await loadRageResult();
		const player = resultData.result.getFirstPlayer()!;

		expect(player.getResourceMetrics(ResourceType.ResourceTypeRage).map(metric => metric.name)).toEqual([
			'common.action_names.attack (common.action_tags.main_hand)',
			'common.action_names.rage_gain',
		]);
		expect(player.getResourceMetrics(ResourceType.ResourceTypeMana)).toEqual([]);
	});

	it('emits the run it is handed on the result channel', async () => {
		const { resultData } = await loadRageResult();
		const parts = resultsHostParts(resultData.result);

		expect((parts.resultChannel as { getSnapshot: () => unknown }).getSnapshot()).toMatchObject({ result: resultData.result, filter: { target: null } });
	});

	it('indexes every row of the rotation model under its key', () => {
		const model = warriorRotationModel();

		expect(model.rows.map(row => model.byKey.get(row.key))).toEqual(model.rows.map((_, index) => index));
		expect(model.sections[0].rowKeys).toEqual(model.rows.slice(1).map(row => row.key));
	});
});
