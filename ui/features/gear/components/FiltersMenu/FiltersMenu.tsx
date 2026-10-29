import { type ArmorType, Faction, ItemSlot, type RangedWeaponType, type WeaponType } from '@generated/proto/common';
import { type DatabaseFilters, FactionFilter, SourceFilterOption } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { sourceFilterI18nKeys } from '@i18n/entity_mapping';
import { translateArmorType, translateRaidFilter, translateRangedWeaponType, translateSourceFilter, translateWeaponType } from '@i18n/localization';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Player } from '@sim/player/player';
import { Sim } from '@sim/sim';
import { subscribePlayerField, subscribeSimField } from '@sim/state/subscriptions';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { NumberField } from '@ui-kit/TextField';

import { MenuSection } from './MenuSection';

const factionLabels: Record<Faction, string> = {
	[Faction.Unknown]: i18n.t('gear_tab.gear_picker.filters.faction_labels.none'),
	[Faction.Alliance]: i18n.t('gear_tab.gear_picker.filters.faction_labels.alliance_only'),
	[Faction.Horde]: i18n.t('gear_tab.gear_picker.filters.faction_labels.horde_only'),
};

// A sorted copy, since `Sim.ALL_SOURCES` is shared. Both inputs are static, so the order is decided
// once here instead of on every render.
const sourceKeys = Object.keys(sourceFilterI18nKeys) as unknown as SourceFilterOption[];
const SOURCES = [...Sim.ALL_SOURCES].sort((a, b) => sourceKeys.indexOf(a) - sourceKeys.indexOf(b));

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
	canDualWield,
	raceFaction,
}: FiltersMenuViewProps) => {
	const listToggle = (field: FilterListField, value: number, id: string, label: string) => (
		<BooleanPicker
			key={id}
			id={id}
			label={label}
			layout="inline"
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
				<div className="grid grid-cols-2 gap-x-(--modal-padding)">
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
				</div>
			</MenuSection>
			<MenuSection name={i18n.t('gear_tab.gear_picker.filters.source')} className="ui-filters-menu-section-bool-list">
				{SOURCES.map(source => {
					const label = translateSourceFilter(source);
					return label ? listToggle('sources', source, `filters-source-${source}`, label) : null;
				})}
			</MenuSection>
			<MenuSection name={i18n.t('gear_tab.gear_picker.filters.raids')} className="ui-filters-menu-section-bool-list">
				{Sim.ALL_RAIDS.map(raid => listToggle('raids', raid, `filters-raid-${raid}`, translateRaidFilter(raid)))}
			</MenuSection>
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
			canDualWield={player.getPlayerSpec().canDualWield}
			raceFaction={raceFaction}
		/>
	);
};
