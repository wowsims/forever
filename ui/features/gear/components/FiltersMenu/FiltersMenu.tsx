import {
	type AreaType,
	type ArmorType,
	type CreatureFamily,
	Faction,
	ItemSlot,
	type MobType,
	type RangedWeaponType,
	type Stat,
	type WeaponType,
} from '@generated/proto/common';
import { type DatabaseFilters, FactionFilter } from '@generated/proto/ui';
import i18n from '@i18n/config';
import {
	translateAreaType,
	translateArmorType,
	translateCreatureFamily,
	translateMobType,
	translateRaidFilter,
	translateRangedWeaponType,
	translateSourceFilter,
	translateStat,
	translateWeaponType,
} from '@i18n/localization';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Player } from '@sim/player/player';
import { Sim } from '@sim/sim';
import { subscribePlayerField, subscribeSimField } from '@sim/state/subscriptions';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { MultiComboBox } from '@ui-kit/MultiComboBox';
import { NumberField } from '@ui-kit/TextField';
import { useMemo } from 'react';

import { filterOptionsOf } from '../../model/filter_options';
import { MenuSection } from './MenuSection';

const factionLabels: Record<Faction, string> = {
	[Faction.Unknown]: i18n.t('gear_tab.gear_picker.filters.faction_labels.none'),
	[Faction.Alliance]: i18n.t('gear_tab.gear_picker.filters.faction_labels.alliance_only'),
	[Faction.Horde]: i18n.t('gear_tab.gear_picker.filters.faction_labels.horde_only'),
};

type FilterListField = { [K in keyof DatabaseFilters]-?: DatabaseFilters[K] extends Array<number> ? K : never }[keyof DatabaseFilters];
// The fields one multi-pick combo box edits; a pick narrows the item list.
type MultiPickField = 'areaTypes' | 'mobTypes' | 'creatureFamilies' | 'itemStats';
type FilterNumberField = { [K in keyof DatabaseFilters]-?: DatabaseFilters[K] extends number ? K : never }[keyof DatabaseFilters];

const SPEED_OPTIONS = { float: true, positive: true };

export type FiltersMenuRow =
	| 'ilvl'
	| 'requiredLevel'
	| 'faction'
	| 'newlyDiscovered'
	| 'stats'
	| 'sources'
	| 'raids'
	| 'bonus'
	| 'armor'
	| 'weapons'
	| 'ranged';

const EVERY_SLOT_ROWS: ReadonlyArray<FiltersMenuRow> = ['ilvl', 'requiredLevel', 'faction', 'newlyDiscovered', 'stats', 'sources', 'raids', 'bonus'];

const slotRows = (slot: ItemSlot | undefined): ReadonlyArray<FiltersMenuRow> => {
	if (slot === undefined) return EVERY_SLOT_ROWS;
	if (Player.ARMOR_SLOTS.includes(slot)) return [...EVERY_SLOT_ROWS, 'armor'];
	if (Player.WEAPON_SLOTS.includes(slot)) return [...EVERY_SLOT_ROWS, 'weapons'];
	return slot === ItemSlot.ItemSlotRanged ? [...EVERY_SLOT_ROWS, 'ranged'] : EVERY_SLOT_ROWS;
};

export interface FiltersMenuViewProps {
	/** The slot whose list the menu narrows, which picks the armor, weapon or ranged rows. */
	slot?: ItemSlot;
	/** The rows to show, for a list that spans several slots. Without it every row shows but the armor, weapon and ranged ones, of which `slot` picks one. */
	rows?: ReadonlyArray<FiltersMenuRow>;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	filters: DatabaseFilters;
	/** Hands over an edit; the caller applies it to the current filters and stores the result. */
	onFiltersChange: (edit: (filters: DatabaseFilters) => void) => void;
	armorTypes: ReadonlyArray<ArmorType>;
	weaponTypes: ReadonlyArray<WeaponType>;
	rangedWeaponTypes: ReadonlyArray<RangedWeaponType>;
	/** The areas, mob types and creature groups some item has a bonus in or against; picking one narrows the list to those items. */
	areaTypes: ReadonlyArray<AreaType>;
	mobTypes: ReadonlyArray<MobType>;
	creatureFamilies: ReadonlyArray<CreatureFamily>;
	/** The stats some item carries, by itself or with a random suffix; picking several narrows the list to the items with all of them. */
	itemStats: ReadonlyArray<Stat>;
	canDualWield: boolean;
	/** The player's race's faction, which names the faction filter's default option. */
	raceFaction: Faction;
}

export const FiltersMenuView = ({
	slot,
	rows,
	open,
	onOpenChange,
	filters,
	onFiltersChange,
	armorTypes,
	weaponTypes,
	rangedWeaponTypes,
	areaTypes,
	mobTypes,
	creatureFamilies,
	itemStats,
	canDualWield,
	raceFaction,
}: FiltersMenuViewProps) => {
	const shown = new Set(rows ?? slotRows(slot));
	const listToggle = (field: FilterListField, value: number, id: string, label: string) => (
		<BooleanPicker
			key={id}
			id={id}
			label={label}
			layout="inline"
			size="xs"
			value={(filters as Record<FilterListField, Array<number>>)[field].includes(value)}
			onChange={checked =>
				onFiltersChange(next => {
					const lists = next as Record<FilterListField, Array<number>>;
					lists[field] = checked ? [...lists[field], value] : lists[field].filter(entry => entry != value);
				})
			}
		/>
	);

	const numberInput = (field: FilterNumberField, id: string, label: string, options: { float?: boolean; positive?: boolean; showZeroes?: boolean }) => (
		<NumberField
			id={id}
			label={label}
			{...options}
			value={filters[field]}
			onChange={value =>
				onFiltersChange(next => {
					next[field] = value;
				})
			}
		/>
	);

	const multiPicker = <T extends number>(field: MultiPickField, id: string, label: string, types: ReadonlyArray<T>, translate: (type: T) => string) => {
		// A stored pick no item carries any more stays on offer, so its chip is there to remove it.
		const offered = [...new Set([...types, ...(filters[field] as ReadonlyArray<number> as ReadonlyArray<T>)])].sort((a, b) => a - b);
		if (!offered.length) return null;
		return (
			<MultiComboBox
				id={id}
				label={label}
				className="min-w-0 flex-col items-stretch"
				placeholder={i18n.t('gear_tab.gear_picker.filters.bonus.any')}
				emptyText={i18n.t('gear_tab.gear_picker.filters.bonus.empty')}
				removeLabel={i18n.t('gear_tab.gear_picker.filters.bonus.remove')}
				openLabel={i18n.t('gear_tab.gear_picker.filters.bonus.open')}
				values={offered.map(type => ({ name: translate(type), value: type }))}
				value={filters[field]}
				onChange={picked =>
					onFiltersChange(next => {
						(next as Record<typeof field, Array<number>>)[field] = picked;
					})
				}
				testId={id}
			/>
		);
	};
	const multiCount = (field: MultiPickField, types: ReadonlyArray<number>) => types.length + filters[field].length;

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="filters-menu"
			size="md"
			elevated
			bodyClassName="gap-2"
			title={i18n.t('gear_tab.gear_picker.filters.title')}>
			{(shown.has('ilvl') || shown.has('requiredLevel') || shown.has('faction') || shown.has('newlyDiscovered')) && (
				<MenuSection name={i18n.t('gear_tab.gear_picker.filters.general')}>
					{shown.has('ilvl') && (
						<div className="relative grid grid-cols-2 gap-x-(--modal-padding)">
							{numberInput('minIlvl', 'filters-min-ilvl', i18n.t('gear_tab.gear_picker.filters.min_ilvl'), { showZeroes: false })}
							<span className="absolute bottom-0 left-1/2 -translate-x-1/2 py-2">-</span>
							{numberInput('maxIlvl', 'filters-max-ilvl', i18n.t('gear_tab.gear_picker.filters.max_ilvl'), { showZeroes: false })}
						</div>
					)}
					{(shown.has('requiredLevel') || shown.has('faction') || shown.has('newlyDiscovered')) && (
						<div className="grid grid-cols-2 gap-x-(--modal-padding) gap-y-2">
							{shown.has('requiredLevel') &&
								numberInput('maxRequiredLevel', 'filters-required-level', i18n.t('gear_tab.gear_picker.filters.required_level'), {
									showZeroes: false,
								})}
							{shown.has('faction') && (
								<ChoiceField
									id="filters-faction-restriction"
									label={i18n.t('gear_tab.gear_picker.filters.faction_restrictions')}
									options={[
										{ value: FactionFilter.FactionFilterRace, label: factionLabels[raceFaction] },
										{ value: FactionFilter.FactionFilterAny, label: factionLabels[Faction.Unknown] },
									]}
									value={filters.factionFilter}
									onChange={value =>
										onFiltersChange(next => {
											next.factionFilter = value;
										})
									}
								/>
							)}
							{shown.has('newlyDiscovered') && (
								<BooleanPicker
									id="filters-newly-discovered"
									label={i18n.t('gear_tab.gear_picker.filters.newly_discovered')}
									layout="inline"
									size="xs"
									value={filters.newlyDiscoveredOnly}
									onChange={checked =>
										onFiltersChange(next => {
											next.newlyDiscoveredOnly = checked;
										})
									}
								/>
							)}
						</div>
					)}
				</MenuSection>
			)}
			{shown.has('stats') && multiCount('itemStats', itemStats) > 0 && (
				<MenuSection name={i18n.t('gear_tab.gear_picker.filters.stats.title')}>
					{multiPicker('itemStats', 'filters-item-stats', i18n.t('gear_tab.gear_picker.filters.stats.label'), itemStats, translateStat)}
				</MenuSection>
			)}
			{shown.has('sources') && (
				<MenuSection name={i18n.t('gear_tab.gear_picker.filters.source')} className="ui-filters-menu-section-bool-list">
					{Sim.ALL_SOURCES.map(source => {
						const label = translateSourceFilter(source);
						return label ? listToggle('sources', source, `filters-source-${source}`, label) : null;
					})}
					<BooleanPicker
						id="filters-source-world-drop"
						label={i18n.t('common.sources.world_drop')}
						layout="inline"
						size="xs"
						value={!filters.hideWorldDrops}
						onChange={checked =>
							onFiltersChange(next => {
								next.hideWorldDrops = !checked;
							})
						}
					/>
				</MenuSection>
			)}
			{shown.has('raids') && (
				<MenuSection name={i18n.t('gear_tab.gear_picker.filters.raids')} className="ui-filters-menu-section-bool-list">
					{Sim.ALL_RAIDS.map(raid => listToggle('raids', raid, `filters-raid-${raid}`, translateRaidFilter(raid)))}
				</MenuSection>
			)}
			{shown.has('bonus') &&
				multiCount('areaTypes', areaTypes) + multiCount('mobTypes', mobTypes) + multiCount('creatureFamilies', creatureFamilies) > 0 && (
					<MenuSection name={i18n.t('gear_tab.gear_picker.filters.bonus.title')} className="grid grid-cols-3 items-start gap-x-2 max-sm:grid-cols-1">
						{multiPicker('areaTypes', 'filters-bonus-area', i18n.t('gear_tab.gear_picker.filters.bonus.area'), areaTypes, translateAreaType)}
						{multiPicker('mobTypes', 'filters-bonus-mob-type', i18n.t('gear_tab.gear_picker.filters.bonus.mob_type'), mobTypes, translateMobType)}
						{multiPicker(
							'creatureFamilies',
							'filters-bonus-creature',
							i18n.t('gear_tab.gear_picker.filters.bonus.creature'),
							creatureFamilies,
							translateCreatureFamily,
						)}
					</MenuSection>
				)}
			{shown.has('armor') && armorTypes.length > 1 && (
				<MenuSection name={i18n.t('gear_tab.gear_picker.armor_type')} className="ui-filters-menu-section-bool-list">
					{armorTypes.map(armorType => listToggle('armorTypes', armorType, `filters-armor-type-${armorType}`, translateArmorType(armorType)))}
				</MenuSection>
			)}
			{shown.has('weapons') && weaponTypes.length > 0 && (
				<>
					<MenuSection name={i18n.t('gear_tab.gear_picker.weapon_type')} className="ui-filters-menu-section-bool-list">
						{weaponTypes.map(weaponType =>
							listToggle('weaponTypes', weaponType, `filters-weapon-type-${weaponType}`, translateWeaponType(weaponType)),
						)}
					</MenuSection>
					<MenuSection name={i18n.t('gear_tab.gear_picker.weapon_speed')} className="ui-filters-menu-section-number-list">
						{numberInput('minMhWeaponSpeed', 'filters-min-weapon-speed', i18n.t('gear_tab.gear_picker.min_mh_speed'), SPEED_OPTIONS)}
						{numberInput('maxMhWeaponSpeed', 'filters-max-weapon-speed', i18n.t('gear_tab.gear_picker.max_mh_speed'), SPEED_OPTIONS)}
						{canDualWield && (
							<>
								{numberInput('minOhWeaponSpeed', 'filters-min-oh-weapon-speed', i18n.t('gear_tab.gear_picker.min_oh_speed'), SPEED_OPTIONS)}
								{numberInput('maxOhWeaponSpeed', 'filters-max-oh-weapon-speed', i18n.t('gear_tab.gear_picker.max_oh_speed'), SPEED_OPTIONS)}
							</>
						)}
					</MenuSection>
				</>
			)}
			{shown.has('ranged') && rangedWeaponTypes.length > 1 && (
				<>
					<MenuSection name={i18n.t('gear_tab.gear_picker.ranged_weapon_type')} className="ui-filters-menu-section-bool-list">
						{rangedWeaponTypes.map(rangedWeaponType =>
							listToggle(
								'rangedWeaponTypes',
								rangedWeaponType,
								`filter-ranged-weapon-type-${rangedWeaponType}`,
								translateRangedWeaponType(rangedWeaponType),
							),
						)}
					</MenuSection>
					<MenuSection name={i18n.t('gear_tab.gear_picker.ranged_weapon_speed')} className="ui-filters-menu-section-number-list">
						{numberInput('minRangedWeaponSpeed', 'filters-min-ranged-weapon-speed', i18n.t('gear_tab.gear_picker.min_ranged_speed'), SPEED_OPTIONS)}
						{numberInput('maxRangedWeaponSpeed', 'filters-max-ranged-weapon-speed', i18n.t('gear_tab.gear_picker.max_ranged_speed'), SPEED_OPTIONS)}
					</MenuSection>
				</>
			)}
		</Dialog>
	);
};

export interface FiltersMenuProps {
	slot: ItemSlot;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

export const FiltersMenu = ({ slot, open, onOpenChange }: FiltersMenuProps) => {
	const player = useSimHost().player;
	const sim = player.sim;
	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const playerClass = player.getPlayerClass();
	const raceFaction = useStoreSubscribe(subscribePlayerField(player, 'race'), () => player.getFaction());
	const { areaTypes, mobTypes, creatureFamilies, itemStats } = useMemo(
		() => filterOptionsOf(sim.db.getAllItems(), id => sim.db.getRandomSuffixById(id)),
		[sim.db],
	);

	return (
		<FiltersMenuView
			slot={slot}
			open={open}
			onOpenChange={onOpenChange}
			filters={filters}
			onFiltersChange={edit => {
				const next = sim.getFilters();
				edit(next);
				sim.setFilters(next);
			}}
			armorTypes={playerClass.armorTypes}
			weaponTypes={playerClass.weaponTypes.map(entry => entry.weaponType)}
			rangedWeaponTypes={playerClass.rangedWeaponTypes}
			areaTypes={areaTypes}
			mobTypes={mobTypes}
			creatureFamilies={creatureFamilies}
			itemStats={itemStats}
			canDualWield={player.getPlayerSpec().canDualWield}
			raceFaction={raceFaction}
		/>
	);
};
