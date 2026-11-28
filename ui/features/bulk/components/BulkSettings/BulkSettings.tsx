import { EnchantPresets } from '@features/enchant-presets/components/EnchantPresets';
import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BulkSimItemSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { usePlayer, useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useSimReady } from '@sim/hooks/useSimReady';
import { formatToNumber } from '@sim/utils/format';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceFieldProps } from '@ui-kit/ChoiceField';
import { RunCard } from '@ui-kit/RunCard';
import { Spinner } from '@ui-kit/Spinner';
import { type ReactNode, useEffect } from 'react';
import { createPortal } from 'react-dom';

import { useBulkBuildsPlan } from '../../context/BulkBuildsContext';
import { useBulkMode } from '../../hooks/useBulkMode';
import { BULK_SETTINGS_GROUP, useBulkSettingsGroups } from '../../hooks/useBulkSettingsGroups';
import { useBulkState } from '../../hooks/useBulkState';
import { useDropsSetup } from '../../hooks/useDropsSetup';
import type { BulkBuildsCount } from '../../model/builds/builds';
import { dropsPickedItemIds, dropsTwoHanderSkipped } from '../../model/drops/items';
import { frozenItemSlot } from '../../model/gear/picker_groups';
import { setBulkFrozenItem, setBulkFrozenWeaponSlot } from '../../model/gear/settings';
import { runBulkBatch } from '../../model/run/run';
import { bulkTopResults } from '../../model/run/top_results';
import { bulkRunLimits, canRunBatch, canRunBuildsBatch, canRunDropsBatch } from '../../model/selectors';
import { BatchEncounter } from '../BatchEncounter';
import { DropsRunCount } from '../DropsRunCount';
import { BulkTopResultsField } from './BulkTopResultsField';
import { CombinationsCount } from './CombinationsCount';
import { FreezeWeaponTypes } from './FreezeWeaponTypes';

export interface BulkWeaponFreeze {
	/** The frozen weapon slot, or -1 for none. */
	value: number;
	weaponTypes: ReactNode;
}

export interface BulkSettingsViewProps {
	/** The top results control under the counts; null with power tools off and in Drops mode. */
	topResults: ReactNode;
	/** The Batch encounter block above the run button; null in Builds mode. */
	batchEncounter: ReactNode;
	combinationsCount: ReactNode;
	canRun: boolean;
	onRun: () => void;
	openGroups: string[];
	onOpenGroupsChange: (groups: string[]) => void;
	/** The setting groups under the run card: all of them in Gear mode, the enchant presets alone in Drops mode, none in Builds mode. */
	groups: 'all' | 'enchants' | 'none';
	runLabel: string;
	enchants: ReactNode;
	frozenItems: ReadonlyArray<ChoiceFieldProps<number>>;
	/** Null for a player who cannot dual wield. */
	weaponFreeze: BulkWeaponFreeze | null;
	onFrozenWeaponSlotChange: (slot: number) => void;
	/** The phone dock's bar: the run button and a one-line count move there, out of the sheet the panel sits in. */
	dock?: BulkSettingsDock | null;
	/** The count the dock's bar shows. */
	summary: { label: string; value: number; pending: boolean };
}

export interface BulkSettingsDock {
	summary: HTMLElement | null;
	action: HTMLElement | null;
}

export const BulkSettingsView = ({
	topResults,
	batchEncounter,
	combinationsCount,
	canRun,
	onRun,
	openGroups,
	onOpenGroupsChange,
	groups,
	runLabel,
	enchants,
	frozenItems,
	weaponFreeze,
	onFrozenWeaponSlotChange,
	dock,
	summary,
}: BulkSettingsViewProps) => {
	const runButton = (
		<Button className={dock ? undefined : 'w-full'} data-testid="bulk-settings-btn" disabled={!canRun} onClick={onRun}>
			{runLabel}
		</Button>
	);
	return (
		<div className="grid gap-6" data-testid="bulk-settings-container">
			{dock?.action && createPortal(runButton, dock.action)}
			{dock?.summary &&
				createPortal(
					<span className="flex items-center gap-2 whitespace-nowrap" data-testid="bulk-dock-summary">
						{summary.label}
						<b className="tabular-nums">{summary.pending ? <Spinner size="xs" /> : formatToNumber(summary.value)}</b>
					</span>,
					dock.summary,
				)}
			<RunCard testId="bulk-run-card" controls={batchEncounter} action={dock ? undefined : runButton}>
				{combinationsCount}
				{topResults}
			</RunCard>
			{groups !== 'none' && (
				<Accordion value={openGroups} onValueChange={onOpenGroupsChange} variant="boxed">
					<AccordionItem
						value={BULK_SETTINGS_GROUP.enchants}
						title={i18n.t('bulk_tab.settings.groups.enchants')}
						testId="bulk-settings-group-enchants">
						{enchants}
					</AccordionItem>
					{groups === 'all' && (
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
					)}
				</Accordion>
			)}
		</div>
	);
};

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

interface BulkSettingsContainerProps {
	builds: BulkBuildsCount | null;
	/** The items Drops mode picks and how many of them the worn two-hander skips; null outside Drops mode. */
	drops: { picked: number; twoHanderSkipped: number } | null;
	dock?: BulkSettingsDock | null;
}

const BulkSettingsContainer = ({ builds, drops, dock }: BulkSettingsContainerProps) => {
	const host = useSimHost();
	const player = usePlayer();
	const ready = useSimReady();
	const frozenItems = useBulkState(slice => slice.frozenItems);
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);
	const { powerTools } = useBulkMode();
	const { combinations: combinationsLimit, iterations: iterationsLimit } = bulkRunLimits(player.sim.isNative, powerTools);
	const canRunGear = useBulkState(slice => canRunBatch(slice, combinationsLimit, iterationsLimit, bulkTopResults(powerTools, slice)));
	const canRunDrops = useBulkState(slice => drops !== null && canRunDropsBatch(slice, drops.picked, combinationsLimit, iterationsLimit));
	const isRunning = useBulkState(slice => slice.isRunning);
	const combinations = useBulkState(slice => slice.combinations);
	const combinationsPending = useBulkState(slice => slice.combinationsPending);
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
			dock={dock}
			summary={
				builds
					? { label: i18n.t('bulk_tab.settings.gear_setups'), value: builds.setups, pending: !ready }
					: {
							label: i18n.t(drops === null ? 'bulk_tab.settings.combinations' : 'bulk_tab.drops.run.candidates'),
							value: combinations,
							pending: combinationsPending || !ready,
						}
			}
			topResults={powerTools && drops === null ? <BulkTopResultsField /> : null}
			batchEncounter={builds ? null : <BatchEncounter />}
			combinationsCount={
				drops === null ? (
					<CombinationsCount builds={builds} ready={ready} />
				) : (
					<DropsRunCount ready={ready} picked={drops.picked} twoHanderSkipped={drops.twoHanderSkipped} />
				)
			}
			canRun={ready && (builds ? canRunBuildsBatch(isRunning, builds, combinationsLimit, iterationsLimit) : drops === null ? canRunGear : canRunDrops)}
			onRun={() => void runBulkBatch(host)}
			openGroups={openGroups}
			onOpenGroupsChange={setOpenGroups}
			groups={builds ? 'none' : drops === null ? 'all' : 'enchants'}
			runLabel={i18n.t(
				builds ? 'bulk_tab.actions.simulate_builds' : drops === null ? 'bulk_tab.actions.simulate_batch' : 'bulk_tab.actions.simulate_drops',
			)}
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

// The picked count reads the database and the live gear, so only Drops mode subscribes to it.
const DropsBulkSettings = (props: BulkSettingsProps) => {
	const player = usePlayer();
	const gear = usePlayerStore('gear');
	const { list } = useDropsSetup();
	return (
		<BulkSettingsContainer
			builds={null}
			drops={{ picked: dropsPickedItemIds(list).length, twoHanderSkipped: dropsTwoHanderSkipped(list, gear, player.getPlayerSpec()) }}
			{...props}
		/>
	);
};

export interface BulkSettingsProps {
	dock?: BulkSettingsDock | null;
}

// In Builds mode the count comes off the plan a `BulkBuildsProvider` above builds.
export const BulkSettings = (props: BulkSettingsProps) => {
	const builds = useBulkBuildsPlan()?.count ?? null;
	const { drops } = useBulkMode();
	return drops ? <DropsBulkSettings {...props} /> : <BulkSettingsContainer builds={builds} drops={null} {...props} />;
};
