import { ConsumableType } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { Database } from '@sim/proto/database';
import { subscribeSimField } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Popover } from '@ui-kit/Popover';
import { NumberField } from '@ui-kit/TextField';
import { useMemo } from 'react';

const TYPE_KEYS: ReadonlyArray<[ConsumableType, string]> = [
	[ConsumableType.ConsumableTypePotion, 'potion'],
	[ConsumableType.ConsumableTypeConjured, 'conjured'],
	[ConsumableType.ConsumableTypeFlask, 'flask'],
	[ConsumableType.ConsumableTypeElixir, 'elixir'],
	[ConsumableType.ConsumableTypeScroll, 'scroll'],
	[ConsumableType.ConsumableTypeFood, 'food'],
	[ConsumableType.ConsumableTypeExplosive, 'explosive'],
	[ConsumableType.ConsumableTypeImbue, 'imbue'],
	[ConsumableType.ConsumableTypeWeaponImbue, 'weaponImbue'],
	[ConsumableType.ConsumableTypeDevice, 'device'],
	[ConsumableType.ConsumableTypeOther, 'other'],
];

export const ConsumesFilters = () => {
	const { sim } = useSimHost();
	const filters = useStoreSubscribe(subscribeSimField(sim, 'filters'), () => sim.getFilters());
	const types = useMemo(() => {
		const ranges = new Map<ConsumableType, [number, number]>();
		for (const { type, lvl } of Database.getSync().getConsumables()) {
			const range = ranges.get(type);
			ranges.set(type, range ? [Math.min(range[0], lvl), Math.max(range[1], lvl)] : [lvl, lvl]);
		}
		return TYPE_KEYS.flatMap(([type, key]) => {
			const range = ranges.get(type);
			return range ? [{ type, key, range: `${range[0]}-${range[1]}` }] : [];
		});
	}, []);
	const setMinLvls = (minConsumableLvl: { [type: number]: number }) => sim.setFilters({ ...sim.getFilters(), minConsumableLvl });

	return (
		<Popover
			trigger={<Icon name="filter" />}
			triggerRender={<Button variant="link" size="sm" />}
			triggerClassName="ml-auto"
			triggerTestId="consumes-filters-button"
			triggerProps={{ 'aria-label': i18n.t('settings_tab.consumables.filters.title') }}
			side="bottom"
			align="end"
			backdrop
			testId="consumes-filters-popover">
			<h6>{i18n.t('settings_tab.consumables.filters.min_lvl')}</h6>
			<div className="grid grid-cols-2 gap-2">
				{types.map(({ type, key, range }) => (
					<NumberField
						key={type}
						id={`consumes-filters-min-lvl-${key}`}
						label={i18n.t(`settings_tab.consumables.filters.types.${key}`)}
						placeholder={range}
						positive
						showZeroes={false}
						value={sim.getMinConsumableLvl(type)}
						onChange={value => setMinLvls({ ...filters.minConsumableLvl, [type]: value })}
					/>
				))}
			</div>
			<Button
				variant="outline-primary"
				size="sm"
				className="mt-2 ml-auto flex"
				data-testid="consumes-filters-reset"
				disabled={Object.keys(filters.minConsumableLvl).length === 0}
				onClick={() => setMinLvls({})}>
				{i18n.t('settings_tab.consumables.filters.reset')}
			</Button>
		</Popover>
	);
};
