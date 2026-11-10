import { getEmptySlotIconUrl } from '@features/gear/model/empty_slot_icons';
import { byLvlDescending } from '@features/gear/model/item_data';
import { formatSearchQuery } from '@features/gear/model/item_search';
import { ItemSlot } from '@generated/proto/common';
import type { BulkEnchantPreset, UIEnchant as Enchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateBulkSlotName } from '@i18n/localization';
import { ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { useSimReady } from '@sim/hooks/useSimReady';
import { useSimStore } from '@sim/hooks/useSimStore';
import { isOffHandOnlyEnchant } from '@sim/proto/items';
import { Accordion, AccordionItem } from '@ui-kit/Accordion';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { IconSwatch } from '@ui-kit/IconSwatch';
import { Pill } from '@ui-kit/Pill';
import { SearchField } from '@ui-kit/TextField';
import { type ReactNode, type RefObject, useCallback, useId, useMemo, useState } from 'react';

import { useEnchantPresets } from '../../hooks/useEnchantPresets';
import { ENCHANT_PRESET_SLOTS, saveEnchantPreset, validateEnchantPresetName } from '../../model/enchant_presets';
import { EnchantPresetTile } from './EnchantPresetTile';

export interface EnchantPresetSlotGroup {
	slot: ItemSlot;
	label: string;
	iconUrl: string;
	checkedCount: number;
	tiles: ReactNode;
}

export interface EnchantPresetDialogViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	title: string;
	name: string;
	onNameChange: (name: string) => void;
	nameTaken: boolean;
	canSave: boolean;
	onSave: () => void;
	groups: readonly EnchantPresetSlotGroup[];
	openSlots: string[];
	onOpenSlotsChange: (slots: string[]) => void;
	search: string;
	onSearchChange: (search: string) => void;
	showAll: boolean;
	onShowAllChange: (showAll: boolean) => void;
	finalFocus?: RefObject<HTMLElement | null>;
}

export const EnchantPresetDialogView = ({
	open,
	onOpenChange,
	finalFocus,
	title,
	name,
	onNameChange,
	nameTaken,
	canSave,
	onSave,
	groups,
	openSlots,
	onOpenSlotsChange,
	search,
	onSearchChange,
	showAll,
	onShowAllChange,
}: EnchantPresetDialogViewProps) => {
	const nameId = useId();
	const showAllId = useId();

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			finalFocus={finalFocus}
			size="xl"
			scrollContents
			elevated
			title={title}
			testId="bulk-enchant-preset-dialog"
			footer={
				<div className="flex gap-2">
					<Button variant="outline-cancel" onClick={() => onOpenChange(false)}>
						{i18n.t('common.cancel')}
					</Button>
					<Button variant="primary" disabled={!canSave} onClick={onSave} data-testid="bulk-enchant-preset-save">
						{i18n.t('bulk_tab.settings.enchant_presets.save')}
					</Button>
				</div>
			}>
			<div className="grid gap-6">
				<Field
					id={nameId}
					className="sm:w-1/2"
					label={i18n.t('bulk_tab.settings.enchant_presets.name')}
					description={nameTaken ? i18n.t('bulk_tab.settings.enchant_presets.name_duplicate') : undefined}>
					<Input
						id={nameId}
						type="text"
						placeholder={i18n.t('bulk_tab.settings.enchant_presets.name_placeholder')}
						aria-invalid={nameTaken}
						value={name}
						onChange={event => onNameChange(event.target.value)}
						onKeyDown={event => {
							if (event.key === 'Enter' && canSave) onSave();
						}}
					/>
				</Field>
				<div className="flex flex-wrap items-center gap-4">
					<SearchField
						className="min-w-0 flex-1"
						value={search}
						onChange={onSearchChange}
						placeholder={i18n.t('common.search')}
						inputTestId="bulk-enchant-search"
					/>
					<BooleanPicker
						id={showAllId}
						label={i18n.t('bulk_tab.settings.enchant_presets.show_all')}
						labelTooltip={i18n.t('bulk_tab.settings.enchant_presets.show_all_tooltip')}
						layout="inline"
						value={showAll}
						onChange={onShowAllChange}
					/>
				</div>
				{search.trim() && !groups.length && <p className="mb-0 text-gray-400">{i18n.t('bulk_tab.search.no_results')}</p>}
				<Accordion value={openSlots} onValueChange={onOpenSlotsChange} variant="boxed">
					{groups.map(group => (
						<AccordionItem
							key={group.slot}
							value={String(group.slot)}
							testId="bulk-enchant-slot"
							panelClassName="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3"
							title={
								<span className="flex items-center gap-2">
									<IconSwatch icon={{ iconUrl: group.iconUrl, name: group.label }} className="size-8 min-w-8" aria-hidden />
									<span>{group.label}</span>
									{group.checkedCount > 0 && (
										<Pill size="xs" className="min-w-[1lh] justify-center" testId="bulk-enchant-slot-count">
											{group.checkedCount}
										</Pill>
									)}
								</span>
							}>
							{group.tiles}
						</AccordionItem>
					))}
				</Accordion>
			</div>
		</Dialog>
	);
};

export interface EnchantPresetDialogProps {
	open: boolean;
	/** Null for a new preset. */
	preset: BulkEnchantPreset | null;
	onClose: () => void;
	finalFocus?: EnchantPresetDialogViewProps['finalFocus'];
}

const slotEnchants = (slot: ItemSlot, enchants: Enchant[]): Enchant[] => {
	if (slot === ItemSlot.ItemSlotMainHand) return enchants.filter(enchant => !isOffHandOnlyEnchant(enchant));
	if (slot === ItemSlot.ItemSlotOffHand) return enchants.filter(isOffHandOnlyEnchant);
	return enchants;
};

const slotLabel = (slot: ItemSlot): string => {
	if (slot === ItemSlot.ItemSlotMainHand) return i18n.t('bulk_tab.settings.enchant_presets.slot_weapons');
	if (slot === ItemSlot.ItemSlotOffHand) return i18n.t('bulk_tab.settings.enchant_presets.slot_off_hand');
	return translateBulkSlotName(ITEM_SLOT_TO_BULK_SIM_ITEM_SLOT.get(slot)!);
};

export const EnchantPresetDialog = ({ open, preset, onClose, finalFocus }: EnchantPresetDialogProps) => {
	const player = usePlayer();
	const presets = useEnchantPresets();
	const [name, setName] = useState(preset?.name ?? '');
	const [selected, setSelected] = useState<ReadonlyMap<ItemSlot, readonly number[]>>(
		() => new Map(preset?.slots.map(options => [options.slot, options.effectIds]) ?? []),
	);

	const ready = useSimReady();
	const slots = useMemo(
		() =>
			ready
				? ENCHANT_PRESET_SLOTS.map(slot => ({
						slot,
						enchants: slotEnchants(slot, [...player.getEnchants(slot)].sort(byLvlDescending)).map(enchant => ({
							enchant,
							relevant: player.isEnchantRelevant(enchant),
							searchName: formatSearchQuery(enchant.name),
						})),
					})).filter(({ enchants }) => enchants.length)
				: [],
		[player, ready],
	);

	const [openSlots, setOpenSlots] = useState(() => preset?.slots.filter(options => options.effectIds.length).map(options => String(options.slot)) ?? []);
	const [search, setSearch] = useState('');
	// A search opens every slot it matches in, less those closed by hand since the search text last changed. The slots
	// opened by hand keep their own state, so clearing the search brings them back.
	const [searchClosedSlots, setSearchClosedSlots] = useState<ReadonlySet<string>>(() => new Set());
	const showAll = useSimStore('filters').showAllEnchants;
	const changeShowAll = (next: boolean) => {
		const filters = player.sim.getFilters();
		filters.showAllEnchants = next;
		player.sim.setFilters(filters);
	};
	const terms = useMemo(() => formatSearchQuery(search).split(' ').filter(Boolean), [search]);
	const shown = useMemo(
		() =>
			slots
				.map(({ slot, enchants }) => ({
					slot,
					enchants: enchants
						.filter(
							({ enchant, relevant, searchName }) =>
								(showAll || relevant || !!selected.get(slot)?.includes(enchant.effectId)) && terms.every(term => searchName.includes(term)),
						)
						.map(({ enchant }) => enchant),
				}))
				.filter(({ enchants }) => enchants.length),
		[slots, terms, showAll, selected],
	);
	const shownSlots = shown.map(({ slot }) => String(slot));
	const changeSearch = (next: string) => {
		setSearch(next);
		setSearchClosedSlots(new Set());
	};
	const changeSearchOpenSlots = (next: string[]) =>
		setSearchClosedSlots(current => new Set([...current, ...shownSlots].filter(slot => !next.includes(slot))));

	const problem = validateEnchantPresetName(presets, name, preset?.name);

	const toggle = useCallback(
		(slot: ItemSlot, effectId: number) =>
			setSelected(current => {
				const ids = current.get(slot) ?? [];
				return new Map(current).set(slot, ids.includes(effectId) ? ids.filter(id => id !== effectId) : [...ids, effectId]);
			}),
		[],
	);

	const save = () => {
		saveEnchantPreset(player, { name, slots: [...selected].map(([slot, effectIds]) => ({ slot, effectIds })) }, preset?.name);
		onClose();
	};

	return (
		<EnchantPresetDialogView
			open={open}
			onOpenChange={next => !next && onClose()}
			finalFocus={finalFocus}
			title={i18n.t(preset ? 'bulk_tab.settings.enchant_presets.edit_title' : 'bulk_tab.settings.enchant_presets.new_title')}
			name={name}
			onNameChange={setName}
			nameTaken={problem === 'duplicate'}
			canSave={problem === null}
			onSave={save}
			openSlots={terms.length ? shownSlots.filter(slot => !searchClosedSlots.has(slot)) : openSlots}
			onOpenSlotsChange={terms.length ? changeSearchOpenSlots : setOpenSlots}
			search={search}
			onSearchChange={changeSearch}
			showAll={showAll}
			onShowAllChange={changeShowAll}
			groups={shown.map(({ slot, enchants }) => ({
				slot,
				label: slotLabel(slot),
				iconUrl: getEmptySlotIconUrl(slot),
				checkedCount: selected.get(slot)?.length ?? 0,
				// One tile per enchant row; rows that share an effect id are one enchant to the sim, so they tick together.
				tiles: enchants.map(enchant => (
					<EnchantPresetTile
						key={`${enchant.effectId}-${enchant.name}`}
						slot={slot}
						enchant={enchant}
						checked={!!selected.get(slot)?.includes(enchant.effectId)}
						onToggle={toggle}
					/>
				)),
			}))}
		/>
	);
};
