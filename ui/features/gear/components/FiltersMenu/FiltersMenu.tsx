import {
	type AreaType,
	type ArmorType,
	type CreatureFamily,
	Faction,
	ItemSlot,
	type MobType,
	type RangedWeaponType,
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

import { MenuSection } from './MenuSection';

const factionLabels: Record<Faction, string> = {
	[Faction.Unknown]: i18n.t('gear_tab.gear_picker.filters.faction_labels.none'),
	[Faction.Alliance]: i18n.t('gear_tab.gear_picker.filters.faction_labels.alliance_only'),
	[Faction.Horde]: i18n.t('gear_tab.gear_picker.filters.faction_labels.horde_only'),
};

type FilterListField = { [K in keyof DatabaseFilters]-?: DatabaseFilters[K] extends Array<number> ? K : never }[keyof DatabaseFilters];
type FilterNumberField = { [K in keyof DatabaseFilters]-?: DatabaseFilters[K] extends number ? K : never }[keyof DatabaseFilters];

const SPEED_OPTIONS = { float: true, positive: true };

export interface FiltersMenuViewProps {
	slot: ItemSlot;
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
	canDualWield: boolean;
	/** The player's race's faction, which names the faction filter's default option. */
	raceFaction: Faction;
}

export const FiltersMenuView = ({
	slot,
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
	canDualWield,
	raceFaction,
}: FiltersMenuViewProps) => {
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

	const bonusPicker = <T extends number>(
		field: 'areaTypes' | 'mobTypes' | 'creatureFamilies',
		id: string,
		label: string,
		types: ReadonlyArray<T>,
		translate: (type: T) => string,
	) => {
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
	const bonusCount = (field: 'areaTypes' | 'mobTypes' | 'creatureFamilies', types: ReadonlyArray<number>) => types.length + filters[field].length;

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			testId="filters-menu"
			size="md"
			elevated
			bodyClassName="gap-2"
			title={i18n.t('gear_tab.gear_picker.filters.title')}>
			<MenuSection name={i18n.t('gear_tab.gear_picker.filters.general')}>
				<div className="relative grid grid-cols-2 gap-x-(--modal-padding)">
					{numberInput('minIlvl', 'filters-min-ilvl', i18n.t('gear_tab.gear_picker.filters.min_ilvl'), { showZeroes: false })}
					<span className="absolute bottom-0 left-1/2 -translate-x-1/2 py-2">-</span>
					{numberInput('maxIlvl', 'filters-max-ilvl', i18n.t('gear_tab.gear_picker.filters.max_ilvl'), { showZeroes: false })}
				</div>
				<div className="grid grid-cols-2 gap-x-(--modal-padding) gap-y-2">
					{numberInput('maxRequiredLevel', 'filters-required-level', i18n.t('gear_tab.gear_picker.filters.required_level'), { showZeroes: false })}
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
				</div>
			</MenuSection>
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
			<MenuSection name={i18n.t('gear_tab.gear_picker.filters.raids')} className="ui-filters-menu-section-bool-list">
				{Sim.ALL_RAIDS.map(raid => listToggle('raids', raid, `filters-raid-${raid}`, translateRaidFilter(raid)))}
			</MenuSection>
			{bonusCount('areaTypes', areaTypes) + bonusCount('mobTypes', mobTypes) + bonusCount('creatureFamilies', creatureFamilies) > 0 && (
				<MenuSection name={i18n.t('gear_tab.gear_picker.filters.bonus.title')} className="grid grid-cols-3 items-start gap-x-2 max-sm:grid-cols-1">
					{bonusPicker('areaTypes', 'filters-bonus-area', i18n.t('gear_tab.gear_picker.filters.bonus.area'), areaTypes, translateAreaType)}
					{bonusPicker('mobTypes', 'filters-bonus-mob-type', i18n.t('gear_tab.gear_picker.filters.bonus.mob_type'), mobTypes, translateMobType)}
					{bonusPicker(
						'creatureFamilies',
						'filters-bonus-creature',
						i18n.t('gear_tab.gear_picker.filters.bonus.creature'),
						creatureFamilies,
						translateCreatureFamily,
					)}
				</MenuSection>
			)}
			{Player.ARMOR_SLOTS.includes(slot)
				? armorTypes.length > 1 && (
						<MenuSection name={i18n.t('gear_tab.gear_picker.armor_type')} className="ui-filters-menu-section-bool-list">
							{armorTypes.map(armorType => listToggle('armorTypes', armorType, `filters-armor-type-${armorType}`, translateArmorType(armorType)))}
						</MenuSection>
					)
				: Player.WEAPON_SLOTS.includes(slot)
					? weaponTypes.length > 0 && (
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
											{numberInput(
												'minOhWeaponSpeed',
												'filters-min-oh-weapon-speed',
												i18n.t('gear_tab.gear_picker.min_oh_speed'),
												SPEED_OPTIONS,
											)}
											{numberInput(
												'maxOhWeaponSpeed',
												'filters-max-oh-weapon-speed',
												i18n.t('gear_tab.gear_picker.max_oh_speed'),
												SPEED_OPTIONS,
											)}
										</>
									)}
								</MenuSection>
							</>
						)
					: slot === ItemSlot.ItemSlotRanged &&
						rangedWeaponTypes.length > 1 && (
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
									{numberInput(
										'minRangedWeaponSpeed',
										'filters-min-ranged-weapon-speed',
										i18n.t('gear_tab.gear_picker.min_ranged_speed'),
										SPEED_OPTIONS,
									)}
									{numberInput(
										'maxRangedWeaponSpeed',
										'filters-max-ranged-weapon-speed',
										i18n.t('gear_tab.gear_picker.max_ranged_speed'),
										SPEED_OPTIONS,
									)}
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
	const { areaTypes, mobTypes, creatureFamilies } = useMemo(() => {
		const items = sim.db.getAllItems();
		const listed = <T extends number>(typesOf: (item: (typeof items)[number]) => Array<T>) => [...new Set(items.flatMap(typesOf))].sort((a, b) => a - b);
		return {
			areaTypes: listed(item => item.areaTypes),
			mobTypes: listed(item => item.mobTypes),
			creatureFamilies: listed(item => item.creatureFamilies),
		};
	}, [sim.db]);

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
			canDualWield={player.getPlayerSpec().canDualWield}
			raceFaction={raceFaction}
		/>
	);
};
