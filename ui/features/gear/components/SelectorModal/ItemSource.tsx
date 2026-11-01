import { UIItem as Item, UIItem_FactionRestriction, UIItemSource as ItemSourceEntry } from '@generated/proto/ui';
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
	if (!item.sources?.length) return null;

	// One source shows, by kind: crafted, drop, rep, quest, pvp, then vendor, whatever order the sources come in.
	const first = <K extends NonNullable<ItemSourceEntry['source']['oneofKind']>>(kind: K) =>
		item.sources.find(entry => entry.source.oneofKind === kind)?.source as Extract<ItemSourceEntry['source'], { oneofKind: K }> | undefined;

	const crafted = first('crafted');
	if (crafted) {
		const src = crafted.crafted;
		const href = src.spellId ? ActionId.makeSpellUrl(src.spellId) : ActionId.makeItemUrl(item.id);
		return <SourceLink href={href}>{professionNames.get(src.profession) ?? i18n.t('common.sources.unknown')}</SourceLink>;
	}

	const drop = first('drop');
	if (drop) {
		const src = drop.drop;
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

	const quest = item.sources.find(entry => entry.source.oneofKind === 'quest' && entry.source.quest.name)?.source;
	if (quest?.oneofKind === 'quest') {
		const src = quest.quest;
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

	if (item.sources.some(entry => entry.source.oneofKind === 'pvp'))
		return <SourceLink href={ActionId.makeItemUrl(item.id)}>{i18n.t('common.sources.pvp')}</SourceLink>;

	const soldBy = first('soldBy');
	if (soldBy) {
		const src = soldBy.soldBy;
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
