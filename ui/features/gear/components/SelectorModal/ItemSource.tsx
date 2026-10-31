import { UIItem as Item, UIItem_FactionRestriction } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { difficultyNames, professionNames, REP_LEVEL_NAMES } from '@sim/proto/names';
import type { Sim } from '@sim/sim';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import type { ReactNode } from 'react';

export interface ItemSourceProps {
	item: Item;
	sim: Sim;
}

const FactionIcon = ({ restriction }: { restriction: UIItem_FactionRestriction }) => {
	if (restriction === UIItem_FactionRestriction.ALLIANCE_ONLY)
		return <img src="/forever/assets/img/alliance.png" className="ml-1" width="15" height="15" alt="" />;
	if (restriction === UIItem_FactionRestriction.HORDE_ONLY) return <img src="/forever/assets/img/horde.png" className="ml-1" width="15" height="15" alt="" />;
	return null;
};

const SourceLink = ({ href, children }: { href: string; children: ReactNode }) => (
	<WowheadAnchor href={href} target="_blank" className="relative z-1">
		<small>{children}</small>
	</WowheadAnchor>
);

export const ItemSource = ({ item, sim }: ItemSourceProps): ReactNode => {
	const pvp = () => <SourceLink href={ActionId.makeItemUrl(item.id)}>{i18n.t('common.sources.pvp')}</SourceLink>;

	if (!item.sources?.length) return null;

	let source = item.sources[0];
	if (source.source.oneofKind === 'crafted') {
		const src = source.source.crafted;
		const href = src.spellId ? ActionId.makeSpellUrl(src.spellId) : ActionId.makeItemUrl(item.id);
		return <SourceLink href={href}>{professionNames.get(src.profession) ?? i18n.t('common.sources.unknown')}</SourceLink>;
	}

	if (source.source.oneofKind === 'drop') {
		const src = source.source.drop;
		const zone = sim.db.getZone(src.zoneId);
		const npc = sim.db.getNpc(src.npcId);
		if (!zone) {
			console.error('No zone found for item:', item);
			return null;
		}

		const category = src.category ? ` - ${src.category}` : '';
		if (npc) {
			return (
				<SourceLink href={ActionId.makeNpcUrl(npc.id)}>
					<span>
						{zone.name} ({difficultyNames.get(src.difficulty) ?? i18n.t('common.sources.unknown')})
						<br />
						{npc.name + category}
					</span>
				</SourceLink>
			);
		}
		if (src.otherName) {
			return (
				<SourceLink href={ActionId.makeZoneUrl(zone.id)}>
					<span>
						{zone.name}
						<br />
						{src.otherName}
					</span>
				</SourceLink>
			);
		}
		return <SourceLink href={ActionId.makeZoneUrl(zone.id)}>{zone.name}</SourceLink>;
	}

	if (source.source.oneofKind === 'quest' && source.source.quest.name) {
		const src = source.source.quest;
		return (
			<SourceLink href={ActionId.makeQuestUrl(src.id)}>
				<span>
					{i18n.t('common.sources.quest')}
					<FactionIcon restriction={item.factionRestriction} />
					<br />
					{src.name}
				</span>
			</SourceLink>
		);
	}

	const reps = item.sources.flatMap(entry => (entry.source.oneofKind === 'rep' ? [entry.source.rep] : []));
	if (reps.length) {
		return (
			<SourceLink href={ActionId.makeFactionUrl(reps[0].repFactionId)}>
				{reps.map((rep, index) => (
					<span key={`${rep.repFactionId}-${index}`}>
						{sim.db.getRepFaction(rep.repFactionId)?.name || i18n.t('common.sources.unknown')} - {REP_LEVEL_NAMES[rep.repLevel]}
						<FactionIcon restriction={item.factionRestriction} />
						<br />
					</span>
				))}
			</SourceLink>
		);
	}

	if (item.sources.some(entry => entry.source.oneofKind === 'pvp')) return pvp();

	if (source.source.oneofKind === 'soldBy') {
		const src = source.source.soldBy;
		return (
			<SourceLink href={ActionId.makeNpcUrl(src.npcId)}>
				<span>
					{i18n.t('common.sources.sold_by')}
					<br />
					{src.npcName}
				</span>
			</SourceLink>
		);
	}

	return null;
};
