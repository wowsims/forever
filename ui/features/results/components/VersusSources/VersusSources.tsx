import i18n from '@i18n/config';
import { translateCreatureFamily, translateMobType, translateStat } from '@i18n/localization';
import type { DamageLog, Entity, VersusLog, VersusLogEffect, VersusLogSource } from '@sim/proto/combat_log';
import { signedStatAmount } from '@sim/proto/stats';
import { Pill } from '@ui-kit/Pill';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import { Fragment } from 'react';

import { EntityLabel } from '../EntityLabel';

export interface VersusSourcesProps {
	/** The hits to list: a tick's own, or every hit of a cast. Hits that took no versus bonus add nothing. */
	logs: ReadonlyArray<DamageLog>;
}

const signedAmount = (amount: number) => `${amount > 0 ? '+' : ''}${Number(amount.toFixed(4))}`;

const versusEffectText = (effect: VersusLogEffect): string => {
	switch (effect.kind) {
		case 'stat':
			return i18n.t('results_tab.details.timeline.versus.stat', {
				amount: signedStatAmount(effect.stat, effect.amount),
				stat: translateStat(effect.stat),
			});
		case 'crit_chance':
			return i18n.t('results_tab.details.timeline.versus.crit_chance', { amount: signedAmount(effect.amount) });
		case 'proc_only':
			return i18n.t('results_tab.details.timeline.versus.proc_only');
		default:
			return i18n.t(`results_tab.details.timeline.versus.${effect.kind}`, { multiplier: Number(effect.multiplier.toFixed(4)) });
	}
};

interface VersusGroup {
	key: string;
	versus: VersusLog;
	sources: Map<string, VersusLogSource>;
	/** The targets whose hits took this group, in the order they were first hit. */
	targets: Map<string, Entity>;
}

const targetKey = (entity: Entity) => `${entity.isTarget}:${entity.isPet}:${entity.index}:${entity.name}`;

/** One group per target kind, with each distinct source and effect once, in the order the hits named them. */
const groupVersus = (logs: ReadonlyArray<DamageLog>): Array<VersusGroup> => {
	const groups = new Map<string, VersusGroup>();
	for (const { versus, target } of logs) {
		if (!versus) continue;
		const key = `${versus.mobType}:${versus.creatureFamily ?? ''}`;
		const group = groups.get(key) ?? { key, versus, sources: new Map<string, VersusLogSource>(), targets: new Map<string, Entity>() };
		groups.set(key, group);
		if (target && !group.targets.has(targetKey(target))) group.targets.set(targetKey(target), target);
		for (const source of versus.sources) {
			const sourceKey = `${source.actionId.toString()}|${JSON.stringify(source.effect)}`;
			if (!group.sources.has(sourceKey)) group.sources.set(sourceKey, source);
		}
	}
	return [...groups.values()];
};

/** The target kind a hit that took a versus bonus is tagged with: "Humanoid", or "Humanoid (Orc)" with a creature family. */
export const VersusTag = ({ versus }: { versus: VersusLog }) => (
	<Pill tone="warning" size="xs" rounded="xs" testId="versus-tag">
		{versus.creatureFamily === undefined
			? translateMobType(versus.mobType)
			: i18n.t('results_tab.details.timeline.versus.tag_family', {
					target: translateMobType(versus.mobType),
					family: translateCreatureFamily(versus.creatureFamily),
				})}
	</Pill>
);

/** The versus bonuses the hits took: "Versus Undead", then one line per source with its effect. */
export const VersusSources = ({ logs }: VersusSourcesProps) => {
	const groups = groupVersus(logs);
	if (groups.length === 0) return null;
	// Names appear only when the hits landed on more than one target, so a lone target's header stays short.
	const nameTargets = new Set(logs.flatMap(({ target }) => (target ? [targetKey(target)] : []))).size > 1;

	return (
		<>
			{groups.map(({ key, versus, sources, targets }) => (
				<div key={key} data-testid="versus-sources" className="flex flex-col gap-1">
					<span className="font-bold text-warning">
						{versus.creatureFamily === undefined
							? i18n.t('results_tab.details.timeline.versus.header', { target: translateMobType(versus.mobType) })
							: i18n.t('results_tab.details.timeline.versus.header_family', {
									target: translateMobType(versus.mobType),
									family: translateCreatureFamily(versus.creatureFamily),
								})}
						{nameTargets &&
							[...targets.values()].map((target, index) => (
								<Fragment key={targetKey(target)}>
									{index === 0 ? ' ' : ', '}
									<EntityLabel entity={target} />
								</Fragment>
							))}
					</span>
					{[...sources].map(([sourceKey, source]) => (
						<span key={sourceKey} className="flex items-center gap-2">
							<WowheadIcon as="div" className="h-4 w-4 flex-none" iconUrl={source.actionId.iconUrl} />
							{source.actionId.name} {versusEffectText(source.effect)}
						</span>
					))}
				</div>
			))}
		</>
	);
};
