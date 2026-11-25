import i18n from '@i18n/config';
import { translateCreatureFamily, translateMobType, translateStat } from '@i18n/localization';
import type { DamageLog, VersusLog, VersusLogEffect, VersusLogSource } from '@sim/proto/combat_log';
import { WowheadIcon } from '@ui-kit/WowheadIcon';

export interface VersusSourcesProps {
	/** The hits to list: a tick's own, or every hit of a cast. Hits that took no versus bonus add nothing. */
	logs: ReadonlyArray<DamageLog>;
}

const numberText = (value: number) => String(Number(value.toFixed(4)));

export const versusEffectText = (effect: VersusLogEffect): string => {
	if (effect.kind === 'stat') {
		return i18n.t('results_tab.details.timeline.versus.stat', {
			amount: `${effect.amount > 0 ? '+' : ''}${numberText(effect.amount)}`,
			stat: translateStat(effect.stat),
		});
	}
	return i18n.t(`results_tab.details.timeline.versus.${effect.kind}`, { multiplier: numberText(effect.multiplier) });
};

const effectKey = (effect: VersusLogEffect) => (effect.kind === 'stat' ? `stat:${effect.stat}:${effect.amount}` : `${effect.kind}:${effect.multiplier}`);

interface VersusGroup {
	key: string;
	versus: VersusLog;
	sources: Map<string, VersusLogSource>;
}

/** One group per target kind, with each distinct source and effect once, in the order the hits named them. */
const groupVersus = (logs: ReadonlyArray<DamageLog>): Array<VersusGroup> => {
	const groups = new Map<string, VersusGroup>();
	for (const { versus } of logs) {
		if (!versus) continue;
		const key = `${versus.mobType}:${versus.creatureFamily ?? ''}`;
		const group = groups.get(key) ?? { key, versus, sources: new Map<string, VersusLogSource>() };
		groups.set(key, group);
		for (const source of versus.sources) {
			const sourceKey = `${source.actionId.toString()}|${effectKey(source.effect)}`;
			if (!group.sources.has(sourceKey)) group.sources.set(sourceKey, source);
		}
	}
	return [...groups.values()];
};

/** The versus bonuses the hits took: "Versus Undead", then one line per source with its effect. */
export const VersusSources = ({ logs }: VersusSourcesProps) => {
	const groups = groupVersus(logs);
	if (groups.length === 0) return null;

	return (
		<>
			{groups.map(({ key, versus, sources }) => (
				<div key={key} data-testid="versus-sources" className="flex flex-col gap-1">
					<span className="font-bold text-warning">
						{versus.creatureFamily === undefined
							? i18n.t('results_tab.details.timeline.versus.header', { target: translateMobType(versus.mobType) })
							: i18n.t('results_tab.details.timeline.versus.header_family', {
									target: translateMobType(versus.mobType),
									family: translateCreatureFamily(versus.creatureFamily),
								})}
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
