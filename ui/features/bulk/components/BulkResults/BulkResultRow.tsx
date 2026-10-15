import { GearChangeIcon } from '@features/gear/components/GearChangeIcon';
import { ItemDetailCell } from '@features/gear/components/ItemCell';
import { ResultReferenceDiff } from '@features/results/components/ResultMetricList';
import { ResultMetricValue } from '@features/results/components/ResultMetricValue';
import { ItemSlot, ItemSpec } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { TopGearResult } from '@sim/bulk/types';
import { BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS, getBulkItemSlotFromSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { type DeltaText, formatDeltaText, formatSignificance, formatToNumber } from '@sim/utils/format';
import { stDevToConf95, zTest } from '@sim/utils/math';
import { Button } from '@ui-kit/Button';
import { toastManager } from '@ui-kit/Toast';
import { Tooltip } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { formatBulkDps } from '../../model/format';

export interface BulkResultRowViewProps {
	showDamage: boolean;
	avg: number;
	/** The 95% confidence interval of this row's own mean; zero leaves it out. */
	plusMinus: number;
	isBase: boolean;
	delta: DeltaText;
	significance: string;
	onEquip: () => void;
	children: ReactNode;
}

export const BulkResultRowView = ({ showDamage, avg, plusMinus, isBase, delta, significance, onEquip, children }: BulkResultRowViewProps) => {
	const marginTooltipId = useId();
	const deltaTooltipId = useId();

	return (
		<div
			className="flex items-center gap-4 not-last:not-only:mb-6 not-last:not-only:border-b not-last:not-only:border-b-border not-last:not-only:pb-6"
			data-testid="bulk-sim-result-root">
			<div className="flex-3 text-center" data-testid="results-sim">
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
						{!isBase && <ResultReferenceDiff diff={{ ...delta, significance }} tooltipId={deltaTooltipId} label={null} />}
						<Tooltip id={marginTooltipId} content={i18n.t('bulk_tab.results.margin_of_error')} />
						<Tooltip id={deltaTooltipId} />
					</div>
				)}
			</div>
			<div className="flex flex-5 flex-wrap gap-1" data-testid="bulk-gear-combo">
				{children}
			</div>
			<div className="flex flex-1 justify-end">
				<Button className={clsx(isBase && 'hidden')} data-testid="bulk-equip-btn" data-hidden={isBase ? '' : undefined} onClick={onEquip}>
					{i18n.t('bulk_tab.results.equip_button')}
				</Button>
			</div>
		</div>
	);
};

export interface BulkResultRowProps {
	result: TopGearResult;
	baseResult: TopGearResult;
	iterations: number;
	onActivateTab: (id: string) => void;
}

const getSwappableItemSlotPair = (slot: number, canDualWield: boolean): [ItemSlot, ItemSlot] | undefined =>
	BULK_SIM_ITEM_SLOT_TO_ITEM_SLOT_PAIRS.get(getBulkItemSlotFromSlot(slot, canDualWield));

const itemSpecPairsEqualUnordered = (resultItems: ItemSpec[], originalItems: ItemSpec[], [slot1, slot2]: [ItemSlot, ItemSlot]): boolean =>
	(ItemSpec.equals(resultItems[slot1], originalItems[slot1]) && ItemSpec.equals(resultItems[slot2], originalItems[slot2])) ||
	(ItemSpec.equals(resultItems[slot1], originalItems[slot2]) && ItemSpec.equals(resultItems[slot2], originalItems[slot1]));

export const BulkResultRow = ({ result, baseResult, iterations, onActivateTab }: BulkResultRowProps) => {
	const host = useSimHost();
	const { damage: showDamage } = useDisplayMetrics(host.sim);

	// Displayed per-row uncertainty (95% CI of the mean). Deliberately the UNPAIRED error:
	// it describes this row's own mean; tie grouping uses the much tighter paired error of
	// the difference between rows — both are correct, they answer different questions.
	const plusMinusDps = stDevToConf95(result.dpsMetrics.stdev, iterations);
	const isBaseResult = result.gear.equals(baseResult.gear);

	const test = zTest(iterations, result.dpsMetrics.avg, result.dpsMetrics.stdev, iterations, baseResult.dpsMetrics.avg, baseResult.dpsMetrics.stdev);
	const delta = formatDeltaText(baseResult.dpsMetrics.avg, result.dpsMetrics.avg, 2, undefined, !test.isDiff, true);

	const canDualWield = getBulkPlayerCanDualWield(host.player);
	const resultAsSpec = result.gear.asSpec();
	const originalEquipmentSpec = baseResult.gear.asSpec();

	return (
		<BulkResultRowView
			showDamage={showDamage}
			avg={result.dpsMetrics.avg}
			plusMinus={plusMinusDps}
			isBase={isBaseResult}
			delta={delta}
			significance={formatSignificance(test)}
			onEquip={() => {
				host.player.setGear(result.gear);
				onActivateTab('gear-tab');
				toastManager.add({ variant: 'success', body: i18n.t('bulk_tab.results.gear_equipped') });
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
									previousItem={baseResult.gear.getEquippedItem(idx) ?? undefined}
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
