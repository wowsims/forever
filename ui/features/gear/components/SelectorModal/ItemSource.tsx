import { DungeonDifficulty, UIItem as Item, UIItem_FactionRestriction, UIItemSource as ItemSourceEntry } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { difficultyNames, professionNames, REP_LEVEL_NAMES } from '@sim/proto/names';
import type { Sim } from '@sim/sim';
import { Button } from '@ui-kit/Button';
import { Popover } from '@ui-kit/Popover';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import type { ReactNode } from 'react';

export interface ItemSourceProps {
	/** An enchant passes its own item's id and its sources, and has no faction restriction. */
	item: Pick<Item, 'id' | 'sources'> & Partial<Pick<Item, 'factionRestriction'>>;
	sim: Sim;
}

// The cell shows the first source by kind, whatever order the sources come in, and the popup the rest in this order.
const KIND_ORDER = ['crafted', 'drop', 'quest', 'rep', 'pvp', 'soldBy'] as const;

const FactionIcon = ({ restriction }: { restriction?: UIItem_FactionRestriction }) => {
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

// Null for a source it cannot name: a quest with no name, or a drop in a zone the database lacks.
const renderSource = ({ source }: ItemSourceEntry, { item, sim }: ItemSourceProps): ReactNode => {
	switch (source.oneofKind) {
		case 'crafted': {
			const src = source.crafted;
			const href = src.spellId ? ActionId.makeSpellUrl(src.spellId) : ActionId.makeItemUrl(item.id);
			return <SourceLink href={href}>{professionNames.get(src.profession) ?? i18n.t('common.sources.unknown')}</SourceLink>;
		}
		case 'drop': {
			const src = source.drop;
			// A world drop has a name and no zone.
			if (!src.zoneId && src.otherName) return <SourceLink href={ActionId.makeItemUrl(item.id)}>{src.otherName}</SourceLink>;

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
							{zone.name}
							{src.difficulty !== DungeonDifficulty.DifficultyUnknown &&
								` (${difficultyNames.get(src.difficulty) ?? i18n.t('common.sources.unknown')})`}
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
		case 'quest': {
			const src = source.quest;
			if (!src.name) return null;
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
		case 'rep': {
			const src = source.rep;
			return (
				<SourceLink href={ActionId.makeFactionUrl(src.repFactionId)}>
					<span>
						{sim.db.getRepFaction(src.repFactionId)?.name || i18n.t('common.sources.unknown')} - {REP_LEVEL_NAMES[src.repLevel]}
						<FactionIcon restriction={item.factionRestriction} />
					</span>
				</SourceLink>
			);
		}
		case 'pvp':
			return <SourceLink href={ActionId.makeItemUrl(item.id)}>{i18n.t('common.sources.pvp')}</SourceLink>;
		case 'soldBy': {
			const src = source.soldBy;
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
		default:
			return null;
	}
};

export const ItemSource = ({ item, sim }: ItemSourceProps): ReactNode => {
	if (!item.sources?.length) return null;

	// The database holds some crafted sources twice, and a repeat is not another source to count.
	const [first, ...others] = KIND_ORDER.flatMap(kind => item.sources.filter(entry => entry.source.oneofKind === kind))
		.filter((entry, index, all) => all.findIndex(other => ItemSourceEntry.equals(other, entry)) === index)
		.map(entry => renderSource(entry, { item, sim }))
		.filter(Boolean);
	if (!others.length) return first ?? null;

	// A popover and not the hover tooltip other counts use: its links take focus and clicks, and its portal leaves the
	// virtualised list. The button sits above the row's click layer, as the link beside it does, so a press equips nothing.
	return (
		<div className="flex items-start gap-2">
			{first}
			<Popover
				trigger={`+${others.length}`}
				triggerRender={<Button variant="outline-secondary" size="none" className="relative z-1 shrink-0 px-1 text-xs" />}
				triggerTestId="item-source-more"
				triggerProps={{ 'aria-label': i18n.t('gear_tab.gear_picker.more_sources', { count: others.length }) }}
				title={i18n.t('gear_tab.gear_picker.other_sources')}
				testId="item-source-more-popup">
				<ul className="ui-popover-body m-0 flex list-none flex-col gap-2 p-0">
					{others.map((source, index) => (
						<li key={index}>{source}</li>
					))}
				</ul>
			</Popover>
		</div>
	);
};
