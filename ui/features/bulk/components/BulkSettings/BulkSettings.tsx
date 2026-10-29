import { EnchantPresets } from '@features/enchant-presets/components/EnchantPresets';
import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BulkSimItemSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { usePlayer, useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimReady } from '@sim/hooks/useSimReady';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Box } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceFieldProps } from '@ui-kit/ChoiceField';
import { RunCard } from '@ui-kit/RunCard';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { type ReactNode, useEffect } from 'react';

import { useBulkBuildsPlan } from '../../context/BulkBuildsContext';
import { useBulkMode } from '../../hooks/useBulkMode';
import { BULK_SETTINGS_GROUP, useBulkSettingsGroups } from '../../hooks/useBulkSettingsGroups';
import { useBulkState } from '../../hooks/useBulkState';
import type { BulkBuildsCount } from '../../model/builds';
import { bulkCombinationsLimit, bulkIterationsLimit } from '../../model/limits';
import { frozenItemSlot } from '../../model/picker_groups';
import { runBulkBatch } from '../../model/run';
import { canRunBatch, canRunBuildsBatch } from '../../model/selectors';
import { setBulkFrozenItem, setBulkFrozenWeaponSlot } from '../../model/settings';
import { bulkTopResults } from '../../model/top_results';
import { BulkTopResultsField } from './BulkTopResultsField';
import { CombinationsCount } from './CombinationsCount';
import { FreezeWeaponTypes } from './FreezeWeaponTypes';

export interface BulkWeaponFreeze {
	/** The frozen weapon slot, or -1 for none. */
	value: number;
	weaponTypes: ReactNode;
}

export interface BulkSettingsViewProps {
	/** The top results control above Simulate Batch; null with power tools off. */
	topResults: ReactNode;
	combinationsCount: ReactNode;
	canRun: boolean;
	onRun: () => void;
	openGroups: string[];
	onOpenGroupsChange: (groups: string[]) => void;
	/** False leaves the run card alone, as in Builds mode. */
	groups: boolean;
	enchants: ReactNode;
	frozenItems: ReadonlyArray<ChoiceFieldProps<number>>;
	/** Null for a player who cannot dual wield. */
	weaponFreeze: BulkWeaponFreeze | null;
	onFrozenWeaponSlotChange: (slot: number) => void;
}

export const BulkSettingsView = ({
	topResults,
	combinationsCount,
	canRun,
	onRun,
	openGroups,
	onOpenGroupsChange,
	groups,
	enchants,
	frozenItems,
	weaponFreeze,
	onFrozenWeaponSlotChange,
}: BulkSettingsViewProps) => (
	<TabPanelColumns.Right>
		<div className="sticky top-sim-header lg:max-h-bulk-settings-max-h lg:overflow-y-auto">
			<Box padding="md" className="grid gap-6 bg-background" testId="bulk-settings-container">
				<RunCard
					testId="bulk-run-card"
					controls={topResults}
					action={
						<Button className="w-full" data-testid="bulk-settings-btn" disabled={!canRun} onClick={onRun}>
							{i18n.t('bulk_tab.actions.simulate_batch')}
						</Button>
					}>
					{combinationsCount}
				</RunCard>
				{groups && (
					<Accordion value={openGroups} onValueChange={onOpenGroupsChange} variant="boxed">
						<AccordionItem
							value={BULK_SETTINGS_GROUP.enchants}
							title={i18n.t('bulk_tab.settings.groups.enchants')}
							testId="bulk-settings-group-enchants">
							{enchants}
						</AccordionItem>
						<AccordionItem
							value={BULK_SETTINGS_GROUP.freezes}
							title={i18n.t('bulk_tab.settings.groups.frozen_slots')}
							panelClassName="grid gap-4"
							testId="bulk-settings-group-freezes">
							{frozenItems.map(props => (
								<ChoiceField key={props.id} {...props} />
							))}
							{weaponFreeze && (
								<>
									<ChoiceField
										id="freeze-weapon"
										label={i18n.t('bulk_tab.settings.freeze_weapon.label')}
										labelTooltip={i18n.t('bulk_tab.settings.freeze_weapon.tooltip')}
										options={[
											{ value: -1, label: i18n.t('common.none') },
											{ value: ItemSlot.ItemSlotMainHand, label: i18n.t('slots.main_hand', { ns: 'character' }) },
											{ value: ItemSlot.ItemSlotOffHand, label: i18n.t('slots.off_hand', { ns: 'character' }) },
										]}
										value={weaponFreeze.value}
										onChange={onFrozenWeaponSlotChange}
									/>
									{weaponFreeze.weaponTypes}
								</>
							)}
						</AccordionItem>
					</Accordion>
				)}
			</Box>
		</div>
	</TabPanelColumns.Right>
);

interface FrozenPair {
	bulkSlot: BulkSimItemSlot.ItemSlotFinger | BulkSimItemSlot.ItemSlotTrinket;
	slots: [ItemSlot, ItemSlot];
	id: string;
	labelKey: string;
	slotKeys: [string, string];
}

const FROZEN_PAIRS: readonly FrozenPair[] = [
	{
		bulkSlot: BulkSimItemSlot.ItemSlotFinger,
		slots: [ItemSlot.ItemSlotFinger1, ItemSlot.ItemSlotFinger2],
		id: 'freeze-ring',
		labelKey: 'bulk_tab.settings.freeze_ring',
		slotKeys: ['slots.finger_1', 'slots.finger_2'],
	},
	{
		bulkSlot: BulkSimItemSlot.ItemSlotTrinket,
		slots: [ItemSlot.ItemSlotTrinket1, ItemSlot.ItemSlotTrinket2],
		id: 'freeze-trinket',
		labelKey: 'bulk_tab.settings.freeze_trinket',
		slotKeys: ['slots.trinket_1', 'slots.trinket_2'],
	},
];

const BulkSettingsContainer = ({ builds }: { builds: BulkBuildsCount | null }) => {
	const host = useSimHost();
	const player = usePlayer();
	const ready = useSimReady();
	const frozenItems = useBulkState(slice => slice.frozenItems);
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);
	const combinationsLimit = bulkCombinationsLimit(player.sim.isNative);
	const { powerTools } = useBulkMode();
	// The iterations gate needs power tools, which hold the Top results control that lowers it. Without them Gear gates on
	// the combinations limit alone, and past the iterations limit the run card only warns.
	const iterationsLimit = powerTools ? bulkIterationsLimit(player.sim.isNative) : Number.POSITIVE_INFINITY;
	const canRunGear = useBulkState(slice => canRunBatch(slice, combinationsLimit, iterationsLimit, bulkTopResults(powerTools, slice)));
	const isRunning = useBulkState(slice => slice.isRunning);
	const gear = usePlayerStore('gear');
	const [openGroups, setOpenGroups] = useBulkSettingsGroups();

	// Clearing a frozen item whose slot no longer holds it writes to the store; that is not safe
	// during render, so it runs in an effect after render instead.
	useEffect(() => {
		for (const pair of FROZEN_PAIRS) {
			const frozenItem = frozenItems.get(pair.bulkSlot);
			if (frozenItem && frozenItemSlot(gear, pair.slots, frozenItem) === null) setBulkFrozenItem(player, pair.bulkSlot, null);
		}
	}, [player, gear, frozenItems]);

	const freezeItemProps = (pair: FrozenPair): ChoiceFieldProps<number> => ({
		id: pair.id,
		label: i18n.t(`${pair.labelKey}.label`),
		labelTooltip: i18n.t(`${pair.labelKey}.tooltip`),
		options: [
			{ value: -1, label: i18n.t('common.none') },
			{ value: pair.slots[0], label: i18n.t(pair.slotKeys[0], { ns: 'character' }) },
			{ value: pair.slots[1], label: i18n.t(pair.slotKeys[1], { ns: 'character' }) },
		],
		value: frozenItemSlot(gear, pair.slots, frozenItems.get(pair.bulkSlot)) ?? -1,
		onChange: (newValue: number) => setBulkFrozenItem(player, pair.bulkSlot, newValue === -1 ? null : player.getGear().getEquippedItem(newValue)),
	});

	return (
		<BulkSettingsView
			topResults={powerTools ? <BulkTopResultsField /> : null}
			combinationsCount={<CombinationsCount builds={builds} ready={ready} />}
			canRun={ready && (builds ? canRunBuildsBatch(isRunning, builds, combinationsLimit, iterationsLimit) : canRunGear)}
			onRun={() => void runBulkBatch(host)}
			openGroups={openGroups}
			onOpenGroupsChange={setOpenGroups}
			groups={!builds}
			enchants={<EnchantPresets />}
			frozenItems={FROZEN_PAIRS.map(freezeItemProps)}
			weaponFreeze={
				getBulkPlayerCanDualWield(player)
					? {
							value: frozenWeaponSlot ?? -1,
							weaponTypes: (
								<>
									<FreezeWeaponTypes slot={ItemSlot.ItemSlotMainHand} />
									<FreezeWeaponTypes slot={ItemSlot.ItemSlotOffHand} />
								</>
							),
						}
					: null
			}
			onFrozenWeaponSlotChange={newValue => setBulkFrozenWeaponSlot(player, newValue === -1 ? null : newValue)}
		/>
	);
};

// In Builds mode the count comes off the plan a `BulkBuildsProvider` above builds.
export const BulkSettings = () => <BulkSettingsContainer builds={useBulkBuildsPlan()?.count ?? null} />;
