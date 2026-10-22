import { pathChoices, type PathOption } from '@features/apl/model/submenus';
import type { UNIT_SET } from '@features/apl/model/unit_sets';
import { unitSets } from '@features/apl/model/unit_sets';
import { refToValue, unitOptionModels, type UnitValue, withIconUrl } from '@features/apl/model/unit_values';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import type { InputConfig } from '@features/inputs/input';
import { UnitReference } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useActionIds } from '@sim/hooks/useActionIds';
import { useUnitMetadataVersion } from '@sim/hooks/useUnitMetadataVersion';
import type { Player } from '@sim/player/player';
import { type ChoiceOption, UnitIcon } from '@ui-kit/ChoiceField';
import { CLASS_TEXT } from '@ui-kit/utils/colors';
import { useMemo } from 'react';
import { useStore } from 'zustand';

export interface UnitFieldProps {
	player: Player<any>;
	config: InputConfig<Player<any>, UnitReference | undefined> & { id: string };
	unitSet: UNIT_SET;
}

/** Two `UnitValue`s are the same unit; the display fields around the reference are deliberately not compared. */
const sameUnit = (a: UnitValue, b: UnitValue) => UnitReference.equals(a.value || UnitReference.create(), b.value || UnitReference.create());

/** A `UnitValue`'s icon, class colour and text as an option; `submenu` files it under a category or owner. */
const unitOption = (unit: UnitValue, submenu?: Array<UnitValue | string>): PathOption<UnitValue> => ({
	value: unit,
	label: unit.text,
	icon: unit.icon ?? (unit.iconUrl ? { iconUrl: unit.iconUrl, name: unit.text ?? '' } : undefined),
	className: unit.color && CLASS_TEXT[unit.color],
	submenu,
});

/** The default unit shows its icon alone on the trigger. */
const renderUnit = ({ value: unit, label }: ChoiceOption<UnitValue>) => (
	<>
		{(unit.icon || unit.iconUrl) && <UnitIcon icon={unit.icon} iconUrl={unit.iconUrl} />}
		{unit.value && label}
	</>
);

/**
 * The unit a field points at.
 *
 * The option list follows unit metadata, which is where pets and targets come from; the selected
 * label is re-resolved on every render, so a pet that gains a name updates without a store write.
 */
export const UnitField = ({ player, config, unitSet }: UnitFieldProps) => {
	const targetUI = unitSets[unitSet].targetUI;
	useUnitMetadataVersion(player.sim);
	useStore(player.sim.store, state => state.raid.targetDummies);
	useStore(player.sim.store, state => state.raid.numActiveParties);
	useStore(player.sim.store, state => state.raid.composition);
	useStore(player.sim.store, state => state.raid.healingSettings);
	const models = unitOptionModels(unitSet, player);
	const icons = useActionIds(models.map(({ unit }) => (typeof unit.iconUrl === 'string' ? undefined : unit.iconUrl)));
	const choices = useMemo(() => {
		const built = pathChoices(
			models.map((model, index) =>
				unitOption(
					withIconUrl(model.unit, icons[index]),
					model.submenu?.map(unit => (typeof unit === 'string' ? unit : withIconUrl(unit))),
				),
			),
			sameUnit,
		);
		// A pet, and with it the only submenu, can arrive after mount; an optionTree from the first render keeps the trigger on the menu path instead of swapping it then.
		return { options: built.options, optionTree: built.optionTree ?? (() => built.options) };
	}, [models, icons]);

	const boundConfig = {
		...config,
		extraClassNames: ['my-auto mr-2 ml-0', ...(config.extraClassNames || [])],
		sourceToValue: (source: UnitReference | undefined) => withIconUrl(refToValue(source, player, targetUI)),
		valueToSource: (unit: UnitValue) => unit.value,
	};

	return (
		<ConnectedChoiceField<Player<any>, UnitReference | undefined, UnitValue>
			modObject={player}
			config={boundConfig}
			options={choices.options}
			optionTree={choices.optionTree}
			equals={sameUnit}
			variant="link"
			placeholder={i18n.t('rotation_tab.apl.helpers.unit_placeholder')}
			renderValue={renderUnit}
		/>
	);
};
