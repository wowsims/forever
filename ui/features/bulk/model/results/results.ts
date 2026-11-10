import { applySavedSettings } from '@features/settings/model/saved_settings';
import type { BulkGearSetup, BulkPlayerSettingsSetup, BulkRankedResult, BulkResults, BulkRotationSetup, BulkTalentSetup, TopGearResult } from '@sim/bulk/types';
import type { Gear } from '@sim/proto/gear';
import { getGearIdentityKey } from '@sim/proto/items';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import { bulkSimHighStageFloor } from '@sim/wasm/bulk_sim/estimate';

import { bonusStatsKey } from '../builds/gear_setups';
import { rotationKey } from '../builds/rotations';
import { settingsContentKey } from '../builds/settings_entries';
import type { CoreBulkSimResult } from '../run/core_sim';
import { buildTieChains } from './tie_chains';

export interface BulkFinishedRequest {
	/** Sims the live gear as its baseline: the Gear-mode request, and in Builds the ticked Current bundle's first request. */
	reference: boolean;
	talents: BulkTalentSetup;
	rotation: BulkRotationSetup;
	/** Null in Gear mode. */
	settings: BulkPlayerSettingsSetup | null;
	/** Null in Gear mode. In Builds, the sent gear setups by gear identity, and the one the baseline stands for, if any. */
	setups: { byGearKey: ReadonlyMap<string, BulkGearSetup>; baseline: BulkGearSetup | null } | null;
	result: CoreBulkSimResult;
}

export type BulkEquipped = 'gear' | 'build';

const gearKey = (gear: Gear) => getGearIdentityKey(gear.asSpec());

// Empty for a Gear-mode row and for one under the current settings, which send none.
const settingsKey = ({ settings }: Pick<BulkRankedResult, 'settings'>): string => (settings?.settings ? settingsContentKey(settings.settings) : '');

export const bulkResultKey = (result: BulkRankedResult): string =>
	[
		gearKey(result.gear),
		result.gearSetup ? bonusStatsKey(result.gearSetup.bonusStats) : '',
		result.talents.talentsString,
		rotationKey(result.rotation),
		settingsKey(result),
	].join('|');

interface NamedRow {
	row: TopGearResult;
	gearSetup: BulkGearSetup | null;
}

// A reference request comes first, and its reference DPS is the current gear's row every delta is measured against. The
// backend's paired errors are against that request's own baseline, so rows from any other request keep only their gear,
// DPS and names, and their ties fall back to the unpaired test. With no reference request every row is such a row, and
// the best one is the reference. With gear setups, a row no setup stands for is dropped, and a row a setup stands for
// carries the gear the setup holds, so it keys, compares and equips alike under every encounter whatever ring enchants
// the request's professions sent. Any other request's baseline is a row only when a gear setup stands for it. The
// ranking keeps the `topResults` best rows besides the current gear's, every row at 0.
export const rankEncounterResults = (requests: readonly BulkFinishedRequest[], currentGear: Gear, topResults: number): BulkResults => {
	const currentKey = gearKey(currentGear);
	const [first] = requests;
	const current: BulkRankedResult | null = first.reference
		? {
				gear: currentGear,
				dpsMetrics: first.result.referenceDpsMetrics,
				talents: first.talents,
				rotation: first.rotation,
				gearSetup: null,
				settings: first.settings,
			}
		: null;
	const pool = requests.flatMap(({ reference, talents, rotation, settings, setups, result }): BulkRankedResult[] => {
		const candidates = result.topGearResults.flatMap((row): NamedRow[] => {
			const key = gearKey(row.gear);
			if (!setups) return key === currentKey ? [] : [{ row, gearSetup: null }];
			const gearSetup = setups.byGearKey.get(key);
			return gearSetup ? [{ row: { ...row, gear: gearSetup.ownGear ?? row.gear }, gearSetup }] : [];
		});
		if (reference) return candidates.map(({ row, gearSetup }) => ({ ...row, talents, rotation, gearSetup, settings }));
		const baselineSetup = setups?.baseline ?? null;
		const baseline: NamedRow[] = baselineSetup
			? [{ row: { gear: baselineSetup.ownGear ?? baselineSetup.gear, dpsMetrics: result.referenceDpsMetrics }, gearSetup: baselineSetup }]
			: [];
		return [...baseline, ...candidates].map(({ row, gearSetup }) => ({
			gear: row.gear,
			dpsMetrics: row.dpsMetrics,
			talents,
			rotation,
			gearSetup,
			settings,
		}));
	});
	const best = pool.sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg).slice(0, topResults || pool.length);
	const ranked = current ? [...best, current].sort((a, b) => b.dpsMetrics.avg - a.dpsMetrics.avg) : best;
	const iterations = Math.max(1, Math.min(...requests.map(({ result }) => result.iterations || bulkSimHighStageFloor())));
	return {
		chains: buildTieChains(ranked, current, iterations),
		originalGearResults: current ?? ranked[0],
		referenceIsCurrent: current !== null,
		currentGear,
		iterations,
	};
};

// A row no gear setup stands for applies the gear only: a Gear-mode row, and the Builds reference row, which is the live
// setup. Any other Builds row applies its talents, rotation and its gear setup's bonus stats too; a row run under the
// current settings leaves the settings alone.
export const equipBulkResult = (host: IndividualSimHost<any>, result: BulkRankedResult): BulkEquipped => {
	const { player } = host;
	const { gearSetup, talents, settings } = result;
	if (!gearSetup) {
		player.setGear(result.gear);
		return 'gear';
	}
	batch(() => {
		player.setGear(result.gear);
		player.setTalentsString(talents.talentsString);
		player.setAplRotation(result.rotation.rotation);
		player.setBonusStats(gearSetup.bonusStats);
		if (settings?.settings) applySavedSettings(host, settings.settings);
	});
	return 'build';
};
