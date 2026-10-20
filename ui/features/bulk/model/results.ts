import type { BulkRankedResult, BulkResults, BulkTalentSetup } from '@sim/bulk/types';
import type { Player } from '@sim/player/player';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import { batch } from '@sim/state/batch';
import { bulkSimHighStageFloor } from '@sim/wasm/bulk_sim/estimate';

import type { CoreBulkSimResult } from './core_sim';
import { buildTieChains } from './tie_chains';

export interface BulkTalentRun {
	talents: BulkTalentSetup;
	result: CoreBulkSimResult;
}

/** The ranked rows per encounter besides the current gear with the current talents. */
export const BULK_RANKED_SETS = 5;

const gearKey = (gear: Gear) => getGearIdentityKey(gear.asSpec());

export const bulkResultKey = (result: BulkRankedResult): string => `${gearKey(result.gear)}|${result.talents.talentsString}`;

// `runs[0]` is the current talents' run, whose reference DPS is the current gear's row every delta is measured against.
// The backend's paired errors are against that run's own baseline, so rows from any other run keep only their gear and
// DPS, and their ties fall back to the unpaired test.
export const rankEncounterResults = (runs: readonly BulkTalentRun[], baseGear: Gear): BulkResults => {
	const baseKey = gearKey(baseGear);
	const reference: BulkRankedResult = { gear: baseGear, dpsMetrics: runs[0].result.referenceDpsMetrics, talents: runs[0].talents };
	const pool = runs.flatMap(({ talents, result }, index): BulkRankedResult[] => {
		const ranked = result.topGearResults.filter(row => gearKey(row.gear) !== baseKey);
		if (index === 0) return ranked.map(row => ({ ...row, talents }));
		return [{ gear: baseGear, dpsMetrics: result.referenceDpsMetrics }, ...ranked].map(row => ({ gear: row.gear, dpsMetrics: row.dpsMetrics, talents }));
	});
	const best = pool.sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg).slice(0, BULK_RANKED_SETS);
	const ranked = [...best, reference].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg);
	const iterations = Math.max(1, Math.min(...runs.map(({ result }) => result.iterations || bulkSimHighStageFloor())));
	return { chains: buildTieChains(ranked, reference, iterations), originalGearResults: reference, iterations, talentSetups: runs.length };
};

/** Wears the row's gear and talents; true when the talents changed. */
export const equipBulkResult = (player: Player<any>, result: BulkRankedResult): boolean => {
	const talentsChanged = player.getTalentsString() !== result.talents.talentsString;
	batch(() => {
		player.setGear(result.gear);
		player.setTalentsString(result.talents.talentsString);
	});
	return talentsChanged;
};
