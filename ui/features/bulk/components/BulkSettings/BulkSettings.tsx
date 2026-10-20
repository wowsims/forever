import { EnchantPresets } from '@features/enchant-presets/components/EnchantPresets';
import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { BulkSimItemSlot, getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { usePlayer, useSimHost } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { useUiStore } from '@sim/hooks/useUiStore';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { Box } from '@ui-kit/Box';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceFieldProps } from '@ui-kit/ChoiceField';
import { RunCard } from '@ui-kit/RunCard';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { type ReactNode, useEffect } from 'react';

import { BULK_SETTINGS_GROUP, useBulkSettingsGroups } from '../../hooks/useBulkSettingsGroups';
import { useBulkState } from '../../hooks/useBulkState';
import { bulkCombinationsLimit } from '../../model/limits';
import { frozenItemSlot } from '../../model/picker_groups';
import { runBulkBatch } from '../../model/run';
import { canRunBatch } from '../../model/selectors';
import { setBulkFrozenItem, setBulkFrozenWeaponSlot } from '../../model/settings';
import { Encounters } from '../Encounters';
import { Talents } from '../Talents';
import { CombinationsCount } from './CombinationsCount';
import { FreezeWeaponTypes } from './FreezeWeaponTypes';

export interface BulkWeaponFreeze {
	/** The frozen weapon slot, or -1 for none. */
	value: number;
	weaponTypes: ReactNode;
}

export interface BulkSettingsViewProps {
	combinationsCount: ReactNode;
	canRun: boolean;
	onRun: () => void;
	openGroups: string[];
	onOpenGroupsChange: (groups: string[]) => void;
	/** Null hides the group, as with power tools off. */
	encounters: ReactNode;
	/** Null hides the group, as with power tools off. */
	talents: ReactNode;
	enchants: ReactNode;
	frozenItems: ReadonlyArray<ChoiceFieldProps<number>>;
	/** Null for a player who cannot dual wield. */
	weaponFreeze: BulkWeaponFreeze | null;
	onFrozenWeaponSlotChange: (slot: number) => void;
}

export const BulkSettingsView = ({
	combinationsCount,
	canRun,
	onRun,
	openGroups,
	onOpenGroupsChange,
	encounters,
	talents,
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
					action={
						<Button className="w-full" data-testid="bulk-settings-btn" disabled={!canRun} onClick={onRun}>
							{i18n.t('bulk_tab.actions.simulate_batch')}
						</Button>
					}>
					{combinationsCount}
				</RunCard>
				<Accordion value={openGroups} onValueChange={onOpenGroupsChange} variant="boxed">
					{encounters != null && (
						<AccordionItem
							value={BULK_SETTINGS_GROUP.encounters}
							title={i18n.t('bulk_tab.settings.groups.encounters')}
							testId="bulk-settings-group-encounters">
							{encounters}
						</AccordionItem>
					)}
					{talents != null && (
						<AccordionItem
							value={BULK_SETTINGS_GROUP.talents}
							title={i18n.t('bulk_tab.settings.groups.talents')}
							testId="bulk-settings-group-talents">
							{talents}
						</AccordionItem>
					)}
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

export const BulkSettings = () => {
	const host = useSimHost();
	const player = usePlayer();
	const frozenItems = useBulkState(slice => slice.frozenItems);
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);
	const canRun = useBulkState(slice => canRunBatch(slice, bulkCombinationsLimit(player.sim.isNative)));
	const gear = usePlayerStore('gear');
	const powerTools = useUiStore('showPowerTools');
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
			combinationsCount={<CombinationsCount />}
			canRun={canRun}
			onRun={() => void runBulkBatch(host)}
			openGroups={openGroups}
			onOpenGroupsChange={setOpenGroups}
			encounters={powerTools ? <Encounters /> : null}
			talents={powerTools ? <Talents /> : null}
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
