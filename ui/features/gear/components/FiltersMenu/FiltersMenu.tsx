import { type ArmorType, ItemSlot, type RangedWeaponType, type WeaponType } from '@generated/proto/common';
import { type DatabaseFilters, SourceFilterOption, UIItem_FactionRestriction } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { sourceFilterI18nKeys } from '@i18n/entity_mapping';
import { translateArmorType, translateRaidFilter, translateRangedWeaponType, translateSourceFilter, translateWeaponType } from '@i18n/localization';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Player } from '@sim/player/player';
import { Sim } from '@sim/sim';
import { subscribeSimField } from '@sim/state/subscriptions';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { ChoiceField } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { NumberField } from '@ui-kit/TextField';

import { MenuSection } from './MenuSection';

const factionRestrictionsToLabels: Record<UIItem_FactionRestriction, string> = {
	[UIItem_FactionRestriction.UNSPECIFIED]: i18n.t('gear_tab.gear_picker.filters.faction_labels.none'),
	[UIItem_FactionRestriction.ALLIANCE_ONLY]: i18n.t('gear_tab.gear_picker.filters.faction_labels.alliance_only'),
	[UIItem_FactionRestriction.HORDE_ONLY]: i18n.t('gear_tab.gear_picker.filters.faction_labels.horde_only'),
};

// A sorted copy, since `Sim.ALL_SOURCES` is shared. Both inputs are static, so the order is decided
// once here instead of on every render.
const sourceKeys = Object.keys(sourceFilterI18nKeys) as unknown as SourceFilterOption[];
const SOURCES = [...Sim.ALL_SOURCES].sort((a, b) => sourceKeys.indexOf(a) - sourceKeys.indexOf(b));

type FilterListField = { [K in keyof DatabaseFilters]: DatabaseFilters[K] extends Array<number> ? K : never }[keyof DatabaseFilters];
type FilterNumberField = { [K in keyof DatabaseFilters]: DatabaseFilters[K] extends number ? K : never }[keyof DatabaseFilters];

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
				<div className="grid grid-cols-split gap-x-(--modal-padding)">
					{numberInput('minIlvl', 'filters-min-ilvl', i18n.t('gear_tab.gear_picker.filters.min_ilvl'), { showZeroes: false })}
					<span className="flex justify-center self-end py-2">-</span>
					{numberInput('maxIlvl', 'filters-max-ilvl', i18n.t('gear_tab.gear_picker.filters.max_ilvl'), { showZeroes: false })}
				</div>
				<ChoiceField
					id="filters-faction-restriction"
					label={i18n.t('gear_tab.gear_picker.filters.faction_restrictions')}
					options={[UIItem_FactionRestriction.UNSPECIFIED, UIItem_FactionRestriction.ALLIANCE_ONLY, UIItem_FactionRestriction.HORDE_ONLY].map(
						restriction => ({ value: restriction, label: factionRestrictionsToLabels[restriction] }),
					)}
					value={filters.factionRestriction}
					onChange={value =>
						onFiltersChange(next => {
							next.factionRestriction = value;
						})
					}
				/>
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
		/>
	);
};
