import { Profession } from '@generated/proto/common';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { createContext, type ReactNode, useContext, useMemo } from 'react';

import { useBulkBundles } from '../hooks/useBulkBundles';
import { useBulkEnchantPicks } from '../hooks/useBulkEnchantPicks';
import { useBulkEncounters } from '../hooks/useBulkEncounters';
import { useBulkMode } from '../hooks/useBulkMode';
import { useBulkState } from '../hooks/useBulkState';
import {
	bulkBuildPlan,
	type BulkBuildsCount,
	bulkBuildsCount,
	bulkBundlePlans,
	bulkGearSetSetups,
	bulkPlanRuns,
	bulkRunningPlans,
	type ResolveBuildRotation,
} from '../model/builds';
import { bulkBundleRuns } from '../model/bundles';
import { bulkEnchantPickRuns } from '../model/enchant_picks';
import { bulkEncounterRuns } from '../model/encounters';

interface BulkBuildsPlan {
	/** What the bundles are read from, as the batch resolves them when it starts. */
	sources: ReturnType<typeof useBulkBundles>;
	count: BulkBuildsCount;
	/** Each bundle that sims like an earlier one and does not run, with that one, which the Builds column names; -1 is the Current bundle. */
	sameAs: ReadonlyMap<number, number>;
}

const BulkBuildsContext = createContext<BulkBuildsPlan | null>(null);

// The plans a Builds batch starts with, built once for the Builds column and the run card. The enchants come from the
// database, so there is no plan until the sim is ready.
const usePlan = (): BulkBuildsPlan => {
	const player = usePlayer();
	const { ready, current, parts, bundles, currentChecked } = useBulkBundles();
	const sources = useMemo(() => ({ ready, current, parts, bundles, currentChecked }), [ready, current, parts, bundles, currentChecked]);
	const { entries: enchantPickEntries, checked: checkedEnchantPicks } = useBulkEnchantPicks();
	const { current: currentEncounter, entries: encounterEntries, checked: checkedEncounters } = useBulkEncounters();
	const profession1 = usePlayerStore('profession1');
	const profession2 = usePlayerStore('profession2');

	// Kept while the enchant picks stay, so an edit to the live talents, rotation or settings or a ticked encounter builds
	// no gear setup again: a plan builds only the gear sets the last plan did not hold.
	const setupsOf = useMemo(
		() => bulkGearSetSetups(player, bulkEnchantPickRuns(enchantPickEntries, checkedEnchantPicks)),
		[player, enchantPickEntries, checkedEnchantPicks],
	);
	const plans = useMemo(
		() => (ready ? bulkBundlePlans(bulkBundleRuns(current, parts, bundles, currentChecked), setupsOf) : []),
		[ready, current, parts, bundles, currentChecked, setupsOf],
	);
	// A rotation is resolved once per talents, pick and items, the build it is given. The rest it reads off the live
	// player, which the live settings stand for, so the resolver is kept until those change, and with it every rotation
	// the last plan resolved.
	const { settings } = current;
	const resolve = useMemo(
		(): ResolveBuildRotation => build => player.getResolvedAplRotationForBuild(build),
		// eslint-disable-next-line react-hooks/exhaustive-deps -- `settings` is the invalidation key, not an input to the resolver.
		[player, settings],
	);
	const reference = useMemo(
		() => ({
			gear: current.gear,
			bonusStats: current.bonusStats,
			isEnchanter: profession1 === Profession.Enchanting || profession2 === Profession.Enchanting,
			settings,
		}),
		[current.gear, current.bonusStats, profession1, profession2, settings],
	);
	const plan = useMemo(() => bulkBuildPlan(plans, reference, resolve), [plans, reference, resolve]);

	const topResults = useBulkState(slice => slice.buildsTopResults);
	const count = useMemo(() => {
		const encounters = bulkEncounterRuns(currentEncounter, encounterEntries, checkedEncounters);
		return bulkBuildsCount(bulkRunningPlans(plans, plan), bulkPlanRuns(plan, encounters), encounters.length, topResults);
	}, [plans, plan, currentEncounter, encounterEntries, checkedEncounters, topResults]);

	return useMemo(() => ({ sources, count, sameAs: plan.sameAs }), [sources, count, plan]);
};

const BuildsPlanProvider = ({ children }: { children?: ReactNode }) => <BulkBuildsContext value={usePlan()}>{children}</BulkBuildsContext>;

// The plan is built only in Builds mode: it looks the gear sets up in the database, which Gear mode never needs.
export const BulkBuildsProvider = ({ children }: { children?: ReactNode }) =>
	useBulkMode().builds ? <BuildsPlanProvider>{children}</BuildsPlanProvider> : children;

/** Null outside Builds mode. */
export const useBulkBuildsPlan = (): BulkBuildsPlan | null => useContext(BulkBuildsContext);
