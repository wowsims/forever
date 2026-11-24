import i18n from '@i18n/config';
import type { BulkRankedResult, BulkResults } from '@sim/bulk/types';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { TalentsConfig } from '@sim/talents/config';
import { classTalentsConfig } from '@sim/talents/factory';
import { type DeltaText, formatDeltaText, formatSignificance } from '@sim/utils/format';
import { stDevToConf95, zTest } from '@sim/utils/math';
import { Button } from '@ui-kit/Button';
import { Pill } from '@ui-kit/Pill';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { gearSetupLabel } from '../../model/builds/gear_setups';
import { rotationLabel } from '../../model/builds/rotations';
import { settingsLabel } from '../../model/builds/settings_entries';
import { talentsLabel } from '../../model/builds/talents';
import { changedEnchantLabels } from '../../model/results/compare';
import { bulkResultKey, equipAndAnnounce } from '../../model/results/results';
import { BulkDpsCell } from '../BulkDpsCell';
import { BulkEnchantsDiffer } from '../BulkEnchantsDiffer';
import { BulkSetupLabel, type BulkSetupLabels, BulkSetupLines, bulkSetupText } from '../BulkSetupLabel';
import { BulkTalentsGrid, bulkTalentsText } from '../BulkTalentsGrid';
import { BulkResultSlots } from './BulkResultSlots';

export interface BulkResultRowViewProps {
	showDamage: boolean;
	avg: number;
	/** The 95% confidence interval of this row's own mean; zero leaves it out. */
	plusMinus: number;
	/** The current gear's row: named so, with no delta and no Equip. */
	isBase: boolean;
	/** The reference row: no delta. */
	isReference?: boolean;
	delta: DeltaText;
	significance: string;
	/** The row's gear setup as its pill label, "Current gear" on a Builds batch's current gear row; none in Gear mode. */
	gearSetup?: string | null;
	/** The gear setup's enchants that differ from its gear set's own, each as "Slot: Enchant"; empty shows no count. */
	enchants?: readonly string[];
	/** The row's talent setup, with its talents in the label's tooltip; none in Gear mode. */
	talents?: { label: string; talentsString: string; config: TalentsConfig<any> } | null;
	/** The row's rotation; none in Gear mode. */
	rotation?: { label: string } | null;
	/** The row's settings; none in Gear mode. */
	settings?: { label: string } | null;
	/** A named bundle's row: its name shows in place of the four parts, which its tooltip lists. */
	bundle?: string | null;
	/** The current gear with other talents, rotation or settings: says so in place of the gear cells. */
	gearUnchanged?: boolean;
	/** A tie group follows this row: its own border separates them, so the row draws no separator or padding. */
	beforeGroup?: boolean;
	/** A row stopped before the final stage: a note under its DPS. */
	lowPrecision?: { note: string; tooltip: string };
	/** Data attributes for the row's root; a `data-testid` among them replaces the row's own. */
	rootDataAttributes?: Record<string, string>;
	onEquip: () => void;
	children: ReactNode;
}

const NO_ENCHANTS: readonly string[] = [];

// The sr-only copies of the talents and of a bundle's parts describe their pills for screen readers, which the tooltips do
// not reach.
export const BulkResultRowView = ({
	showDamage,
	avg,
	plusMinus,
	isBase,
	isReference = false,
	delta,
	significance,
	gearSetup = null,
	enchants = NO_ENCHANTS,
	talents = null,
	rotation = null,
	settings = null,
	bundle = null,
	gearUnchanged = false,
	beforeGroup = false,
	lowPrecision,
	rootDataAttributes,
	onEquip,
	children,
}: BulkResultRowViewProps) => {
	const talentsTooltipId = useId();
	const bundleTooltipId = useId();
	const parts: BulkSetupLabels = { gear: gearSetup, talents: talents?.label ?? null, rotation: rotation?.label ?? null, settings: settings?.label ?? null };

	return (
		<div
			className={clsx(
				'flex items-center gap-4 not-last:not-only:mb-6 max-md:flex-wrap',
				!beforeGroup && 'not-last:not-only:border-b not-last:not-only:border-b-border not-last:not-only:pb-6',
			)}
			data-testid="bulk-sim-result-root"
			{...rootDataAttributes}>
			{/* Wide enough for "9,999.99 DPS(±1)" and its pill on one line each, so every row's gear starts at the same x. */}
			<div className="w-52 shrink-0" data-testid="results-sim">
				{showDamage && (
					<BulkDpsCell
						testId="results-sim-dps"
						label={isBase ? i18n.t('bulk_tab.results.current_gear') : undefined}
						avg={avg}
						plusMinus={plusMinus}
						diff={isBase || isReference ? null : { ...delta, significance }}
						lowPrecision={lowPrecision}
					/>
				)}
			</div>
			<div className="flex min-w-0 flex-1 flex-wrap gap-1 max-md:order-last max-md:basis-full" data-testid="bulk-gear-combo">
				{(gearSetup !== null || talents || rotation || settings || bundle) && (
					<div className="flex basis-full flex-wrap items-center gap-2 text-sm wrap-break-word" data-testid="bulk-result-talents">
						{bundle && (
							<>
								<Pill
									rounded="xs"
									tabIndex={0}
									className="cursor-default focus-visible:focus-ring"
									aria-describedby={`${bundleTooltipId}-text`}
									testId="bulk-result-bundle"
									{...tooltipAnchorProps(bundleTooltipId)}>
									{bundle}
								</Pill>
								<span id={`${bundleTooltipId}-text`} className="sr-only">
									{bulkSetupText(parts)}
								</span>
								<Tooltip id={bundleTooltipId} align="start" content={<BulkSetupLines labels={parts} />} />
							</>
						)}
						{!bundle && gearSetup !== null && (
							<Pill rounded="xs" testId="bulk-result-gear-set">
								<BulkSetupLabel kind="gear" label={gearSetup} />
							</Pill>
						)}
						{enchants.length > 0 && <BulkEnchantsDiffer enchants={enchants} />}
						{!bundle && talents && (
							<>
								<Pill
									rounded="xs"
									tabIndex={0}
									className="cursor-default focus-visible:focus-ring"
									aria-describedby={`${talentsTooltipId}-text`}
									testId="bulk-result-talents-anchor"
									{...tooltipAnchorProps(talentsTooltipId)}>
									<BulkSetupLabel kind="talents" label={talents.label} />
								</Pill>
								<span id={`${talentsTooltipId}-text`} className="sr-only">
									{bulkTalentsText(talents)}
								</span>
								<Tooltip
									id={talentsTooltipId}
									align="start"
									content={<BulkTalentsGrid config={talents.config} talentsString={talents.talentsString} />}
								/>
							</>
						)}
						{!bundle && rotation && (
							<Pill rounded="xs" testId="bulk-result-rotation">
								<BulkSetupLabel kind="rotation" label={rotation.label} />
							</Pill>
						)}
						{!bundle && settings && (
							<Pill rounded="xs" testId="bulk-result-settings">
								<BulkSetupLabel kind="settings" label={settings.label} />
							</Pill>
						)}
					</div>
				)}
				{gearUnchanged ? (
					<span className="text-sm text-muted" data-testid="bulk-result-gear-unchanged">
						{i18n.t('bulk_tab.results.gear_unchanged')}
					</span>
				) : (
					children
				)}
			</div>
			<div className="ml-auto flex shrink-0 justify-end">
				<Button className={clsx(isBase && 'hidden')} data-testid="bulk-equip-btn" data-hidden={isBase ? '' : undefined} onClick={onEquip}>
					{i18n.t('bulk_tab.results.equip_button')}
				</Button>
			</div>
		</div>
	);
};

export interface BulkResultRowProps {
	result: BulkRankedResult;
	results: BulkResults;
	beforeGroup?: boolean;
	onActivateTab: (id: string) => void;
}

export const BulkResultRow = ({ result, results, beforeGroup, onActivateTab }: BulkResultRowProps) => {
	const host = useSimHost();
	const { damage: showDamage } = useDisplayMetrics(host.sim);
	const { originalGearResults: baseResult, currentGear, iterations } = results;

	// Displayed per-row uncertainty (95% CI of the mean). Deliberately the UNPAIRED error:
	// it describes this row's own mean; tie grouping uses the much tighter paired error of
	// the difference between rows — both are correct, they answer different questions.
	const plusMinusDps = stDevToConf95(result.dpsMetrics.stdev, iterations);
	const isReference = bulkResultKey(result) === bulkResultKey(baseResult);
	const isBaseResult = isReference && results.referenceIsCurrent;
	const gearUnchanged = !isBaseResult && result.gear.equals(currentGear);

	const test = zTest(iterations, result.dpsMetrics.avg, result.dpsMetrics.stdev, iterations, baseResult.dpsMetrics.avg, baseResult.dpsMetrics.stdev);
	const delta = formatDeltaText(baseResult.dpsMetrics.avg, result.dpsMetrics.avg, 2, undefined, !test.isDiff, true);

	// Gear-mode rows carry no settings; a Builds row names all four parts of its build, shared with every other row or not.
	const builds = result.settings !== null;

	return (
		<BulkResultRowView
			showDamage={showDamage}
			avg={result.dpsMetrics.avg}
			plusMinus={plusMinusDps}
			isBase={isBaseResult}
			isReference={isReference}
			delta={delta}
			significance={formatSignificance(test)}
			gearSetup={result.gearSetup ? gearSetupLabel(result.gearSetup) : builds ? i18n.t('bulk_tab.results.current_gear') : null}
			enchants={changedEnchantLabels(result)}
			talents={
				builds
					? {
							label: talentsLabel(result.talents),
							talentsString: result.talents.talentsString,
							config: classTalentsConfig[host.player.getClass()],
						}
					: null
			}
			rotation={builds ? { label: rotationLabel(result.rotation) } : null}
			settings={result.settings ? { label: settingsLabel(result.settings) } : null}
			bundle={result.gearSetup?.bundle ?? null}
			gearUnchanged={gearUnchanged}
			beforeGroup={beforeGroup}
			onEquip={() => equipAndAnnounce(host, result, onActivateTab)}>
			{!isBaseResult && <BulkResultSlots gear={result.gear} currentGear={currentGear} />}
		</BulkResultRowView>
	);
};
