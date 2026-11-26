import type { VersusEffect } from '@generated/proto/api';
import { type AreaType, CreatureFamily, type Stat, type Target } from '@generated/proto/common';
import type { UIItem } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { translateAreaType, translateCreatureFamily, translateMobType, translateStat } from '@i18n/localization';
import { useSim } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { signedStatAmount } from '@sim/proto/stats';
import { Pill, type PillTone } from '@ui-kit/Pill';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { Fragment, useId } from 'react';
import { useStore } from 'zustand';

import { itemVersusChips, namesVersus, type VersusChip, versusEffectFactor, type VersusState } from '../../model/versus_state';

const TONES: Record<VersusState, PillTone> = { on: 'positive', partial: 'warning', off: 'neutral' };

const chipName = (chip: VersusChip) =>
	chip.kind === 'mobType' ? translateMobType(chip.value) : chip.kind === 'family' ? translateCreatureFamily(chip.value) : translateAreaType(chip.value);

// A family chip names a target by its family, or by its mob type when it has none.
const targetName = (chip: VersusChip, target: Target) =>
	chip.kind === 'family' && target.creatureFamily !== CreatureFamily.CreatureFamilyUnknown
		? translateCreatureFamily(target.creatureFamily)
		: translateMobType(target.mobType);

const statLines = (stats: { [key: number]: number }) =>
	Object.entries(stats).map(([key, amount]) => {
		const stat = Number(key) as Stat;
		return `${signedStatAmount(stat, amount)} ${translateStat(stat)}`;
	});

const valueLines = (chip: VersusChip, item: UIItem, versusEffects: ReadonlyArray<VersusEffect>): Array<string> => {
	const scaling = item.scalingOptions[0];
	const stats =
		chip.kind === 'mobType'
			? (scaling?.mobTypeStats ?? []).filter(entry => entry.mobTypes.includes(chip.value)).map(entry => entry.stats)
			: chip.kind === 'area'
				? (scaling?.areaStats ?? []).filter(entry => entry.areaType === chip.value).map(entry => entry.stats)
				: [];
	const effects = versusEffects.filter(
		effect =>
			effect.itemId === item.id &&
			(chip.kind === 'mobType' ? effect.mobTypes.includes(chip.value) : chip.kind === 'family' && effect.creatureFamilies.includes(chip.value)),
	);
	return [...stats.flatMap(statLines), ...effects.map(versusEffectFactor)];
};

const reasonLine = (chip: VersusChip, targets: ReadonlyArray<Target>): string => {
	if (chip.kind === 'area') return i18n.t(`gear_tab.versus.${chip.state === 'on' ? 'active_area' : 'inactive_area'}`, { area: chipName(chip) });
	if (chip.state === 'partial') return i18n.t('gear_tab.versus.active_some', { matched: chip.matched, total: chip.total });
	if (!targets.length) return i18n.t('gear_tab.versus.no_targets');
	const listed =
		targets.length === 1
			? targetName(chip, targets[0])
			: targets.map((target, index) => i18n.t('gear_tab.versus.target', { name: targetName(chip, target), number: index + 1 })).join(', ');
	return i18n.t(`gear_tab.versus.${chip.state === 'on' ? 'active_targets' : 'inactive_targets'}`, { count: targets.length, targets: listed });
};

export interface VersusChipsViewProps {
	item: UIItem;
	targets: ReadonlyArray<Target>;
	areaTypes: ReadonlyArray<AreaType>;
	/** The sim's factors for the worn items, matched to `item` by id; an item not worn has none. */
	versusEffects: ReadonlyArray<VersusEffect>;
}

export const VersusChipsView = ({ item, targets, areaTypes, versusEffects }: VersusChipsViewProps) => {
	const tooltipId = useId();
	const chips = itemVersusChips(item, targets, areaTypes);
	if (!chips.length) return null;

	return (
		<span className="relative z-1 flex flex-wrap gap-1">
			{chips.map((chip, index) => {
				const id = `${tooltipId}-${index}`;
				const name = chipName(chip);
				const values = valueLines(chip, item, versusEffects);
				const reason = reasonLine(chip, targets);
				const title =
					chip.kind === 'area'
						? i18n.t('sidebar.character_stats.versus.in_area', { area: name })
						: i18n.t('sidebar.character_stats.versus.against', { target: name });
				return (
					<Fragment key={`${chip.kind}-${chip.value}`}>
						<Pill
							tone={TONES[chip.state]}
							size="xs"
							rounded="xs"
							tabIndex={0}
							className="ui-versus-chip cursor-pointer focus-visible:focus-ring"
							data-state={chip.state}
							aria-describedby={`${id}-text`}
							testId="gear-versus-chip"
							{...tooltipAnchorProps(id)}>
							{chip.state === 'partial' ? i18n.t('gear_tab.versus.partial', { name, matched: chip.matched, total: chip.total }) : name}
						</Pill>
						<span id={`${id}-text`} className="sr-only">
							{[...values, reason].join('. ')}
						</span>
						<Tooltip
							id={id}
							align="start"
							content={
								<div className="flex flex-col gap-1">
									<span className="font-semibold">{title}</span>
									{values.map((line, at) => (
										<span key={at}>{line}</span>
									))}
									<span className="text-muted">{reason}</span>
								</div>
							}
						/>
					</Fragment>
				);
			})}
		</span>
	);
};

export interface VersusChipsProps {
	item: UIItem;
}

const ConnectedVersusChips = ({ item }: VersusChipsProps) => {
	const sim = useSim();
	const targets = useStore(sim.store, state => state.encounter.targets);
	const areaTypes = useStore(sim.store, state => state.encounter.areaTypes);
	const { versusEffects } = usePlayerStore('currentStats');
	return <VersusChipsView item={item} targets={targets} areaTypes={areaTypes} versusEffects={versusEffects} />;
};

// The check comes before the store reads, so an item naming nothing adds no subscription to a gear cell or list row.
export const VersusChips = ({ item }: VersusChipsProps) => (namesVersus(item) ? <ConnectedVersusChips item={item} /> : null);
