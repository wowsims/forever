import { GearChangeIcon } from '@features/gear/components/GearChangeIcon';
import { ItemDetailCell } from '@features/gear/components/ItemCell';
import { ResultReferenceDiff } from '@features/results/components/ResultMetricList';
import { ResultMetricValue } from '@features/results/components/ResultMetricValue';
import { ItemSpec } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { BulkRankedResult, BulkResults } from '@sim/bulk/types';
import { getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import type { TalentsConfig } from '@sim/talents/config';
import { classTalentsConfig } from '@sim/talents/factory';
import { type DeltaText, formatDeltaText, formatSignificance, formatToNumber } from '@sim/utils/format';
import { stDevToConf95, zTest } from '@sim/utils/math';
import { Button } from '@ui-kit/Button';
import { Pill } from '@ui-kit/Pill';
import { toastManager } from '@ui-kit/Toast';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { gearSetupLabel } from '../../model/builds/gear_setups';
import { rotationLabel } from '../../model/builds/rotations';
import { settingsLabel } from '../../model/builds/settings_entries';
import { talentsLabel } from '../../model/builds/talents';
import { getSwappableItemSlotPair, itemSpecPairsEqualUnordered } from '../../model/gear/swappable_slots';
import { changedEnchantLabels } from '../../model/results/compare';
import { formatBulkDps } from '../../model/results/format';
import { type BulkEquipped, bulkResultKey, equipBulkResult } from '../../model/results/results';
import { BulkEnchantsDiffer } from '../BulkEnchantsDiffer';
import { BulkSetupLabel } from '../BulkSetupLabel';
import { BulkTalentsGrid, bulkTalentsText } from '../BulkTalentsGrid';

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
	/** The current gear with other talents, rotation or settings: says so in place of the gear cells. */
	gearUnchanged?: boolean;
	/** A tie group follows this row: its own border separates them, so the row draws no separator or padding. */
	beforeGroup?: boolean;
	onEquip: () => void;
	children: ReactNode;
}

const NO_ENCHANTS: readonly string[] = [];

// The sr-only copy of the talents describes the talent pill for screen readers, which the tooltip does not reach.
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
	gearUnchanged = false,
	beforeGroup = false,
	onEquip,
	children,
}: BulkResultRowViewProps) => {
	const marginTooltipId = useId();
	const deltaTooltipId = useId();
	const talentsTooltipId = useId();

	return (
		<div
			className={clsx(
				'flex items-center gap-4 not-last:not-only:mb-6 max-md:flex-wrap',
				!beforeGroup && 'not-last:not-only:border-b not-last:not-only:border-b-border not-last:not-only:pb-6',
			)}
			data-testid="bulk-sim-result-root">
			{/* Wide enough for "9,999.99 DPS(±1)" and its pill on one line each, so every row's gear starts at the same x. */}
			<div className="w-52 shrink-0" data-testid="results-sim">
				{showDamage && (
					<div className="flex flex-col items-start gap-2 text-left font-bold" data-testid="results-sim-dps">
						{isBase && <span className="text-sm font-normal">{i18n.t('bulk_tab.results.current_gear')}</span>}
						<div>
							<ResultMetricValue
								average={formatBulkDps(avg)}
								unit={i18n.t('sidebar.results.metrics.dps.label')}
								margin={plusMinus > 0 ? formatToNumber(plusMinus, { maximumFractionDigits: 0 }) : undefined}
								marginTooltipId={marginTooltipId}
							/>
						</div>
						{!isBase && !isReference && <ResultReferenceDiff diff={{ ...delta, significance }} tooltipId={deltaTooltipId} hideLabel />}
						<Tooltip id={marginTooltipId} content={i18n.t('bulk_tab.results.margin_of_error')} />
						<Tooltip id={deltaTooltipId} />
					</div>
				)}
			</div>
			<div className="flex min-w-0 flex-1 flex-wrap gap-1 max-md:order-last max-md:basis-full" data-testid="bulk-gear-combo">
				{(gearSetup !== null || talents || rotation || settings) && (
					<div className="flex basis-full flex-wrap items-center gap-2 text-sm wrap-break-word" data-testid="bulk-result-talents">
						{gearSetup !== null && (
							<Pill rounded="xs" testId="bulk-result-gear-set">
								<BulkSetupLabel kind="gear" label={gearSetup} />
							</Pill>
						)}
						{enchants.length > 0 && <BulkEnchantsDiffer enchants={enchants} />}
						{talents && (
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
						{rotation && (
							<Pill rounded="xs" testId="bulk-result-rotation">
								<BulkSetupLabel kind="rotation" label={rotation.label} />
							</Pill>
						)}
						{settings && (
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

const equippedToast = (equipped: BulkEquipped): string => {
	switch (equipped) {
		case 'build':
			return i18n.t('bulk_tab.results.build_equipped');
		case 'gear':
			return i18n.t('bulk_tab.results.gear_equipped');
	}
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

	const canDualWield = getBulkPlayerCanDualWield(host.player);
	const resultAsSpec = result.gear.asSpec();
	const originalEquipmentSpec = currentGear.asSpec();
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
			gearUnchanged={gearUnchanged}
			beforeGroup={beforeGroup}
			onEquip={() => {
				const equipped = equipBulkResult(host, result);
				onActivateTab('gear-tab');
				toastManager.add({ variant: 'success', body: equippedToast(equipped) });
			}}>
			{!isBaseResult &&
				resultAsSpec.items.map((spec, idx) => {
					const swappableItemSlotPair = getSwappableItemSlotPair(idx, canDualWield);
					const itemChanged = swappableItemSlotPair
						? !itemSpecPairsEqualUnordered(resultAsSpec.items, originalEquipmentSpec.items, swappableItemSlotPair)
						: !ItemSpec.equals(spec, originalEquipmentSpec.items[idx]);

					// Three display states per slot:
					// 1. Unchanged -  slots stay empty
					// 2. Same - Item with a different enchant shows the compact change icon (enchant marker);
					// 3. New - The item appearing at all already says the slot changed.
					if (itemChanged && spec.id !== 0 && spec.id === originalEquipmentSpec.items[idx]?.id) {
						return (
							<div key={idx} className="ui-bulk-result-item ui-bulk-item-cell" data-testid="bulk-result-item">
								<GearChangeIcon
									slot={idx}
									item={host.sim.db.lookupItemSpec(spec) ?? undefined}
									previousItem={currentGear.getEquippedItem(idx) ?? undefined}
									showEnchant
								/>
							</div>
						);
					}

					return (
						<ItemDetailCell
							key={idx}
							testId="bulk-result-item"
							className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
							slot={idx}
							item={itemChanged && spec.id !== 0 ? host.sim.db.lookupItemSpec(spec) : null}
						/>
					);
				})}
		</BulkResultRowView>
	);
};
