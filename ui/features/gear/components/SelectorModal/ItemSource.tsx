import { Faction } from '@generated/proto/common';
import { type DropSource, DungeonDifficulty, UIItem as Item, UIItem_FactionRestriction, UIItemSource as ItemSourceEntry } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { difficultyNames, professionNames, REP_LEVEL_NAMES } from '@sim/proto/names';
import type { Sim } from '@sim/sim';
import { Button } from '@ui-kit/Button';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { type ReactNode, useId } from 'react';

export interface ItemSourceProps {
	/** An enchant passes its own item's id and its sources, and has no faction restriction. */
	item: Pick<Item, 'id' | 'sources'> & Partial<Pick<Item, 'factionRestriction'>>;
	sim: Sim;
	/** The character's side, whose quests and vendors show; omit it to show both sides'. */
	faction?: Faction;
}

// The cell shows the first source by kind, whatever order the sources come in, and the tooltip the rest in this order.
const KIND_ORDER = ['crafted', 'drop', 'quest', 'rep', 'pvp', 'soldBy'] as const;

const FactionIcon = ({ restriction }: { restriction?: UIItem_FactionRestriction }) => {
	if (restriction === UIItem_FactionRestriction.ALLIANCE_ONLY)
		return <img src="/forever/assets/img/alliance.png" className="ml-1" width="15" height="15" alt="" />;
	if (restriction === UIItem_FactionRestriction.HORDE_ONLY) return <img src="/forever/assets/img/horde.png" className="ml-1" width="15" height="15" alt="" />;
	return null;
};

// A world drop has a name and no zone.
const isWorldDrop = ({ zoneId, otherName }: DropSource): boolean => !zoneId && !!otherName;

const SourceLink = ({ href, children }: { href: string; children: ReactNode }) => (
	<WowheadAnchor href={href} target="_blank" className="relative z-1">
		<small>{children}</small>
	</WowheadAnchor>
);

// False for a source it cannot name: a quest with no name, or a drop in a zone the database lacks.
const canShowSource = ({ source }: ItemSourceEntry, { item, sim }: ItemSourceProps): boolean => {
	if (source.oneofKind === 'quest') return !!source.quest.name;
	if (source.oneofKind !== 'drop' || isWorldDrop(source.drop) || sim.db.getZone(source.drop.zoneId)) return true;
	console.error('No zone found for item:', item);
	return false;
};

const renderSource = ({ source }: ItemSourceEntry, { item, sim }: ItemSourceProps): ReactNode => {
	switch (source.oneofKind) {
		case 'crafted': {
			const src = source.crafted;
			const href = src.spellId ? ActionId.makeSpellUrl(src.spellId) : ActionId.makeItemUrl(item.id);
			return <SourceLink href={href}>{professionNames.get(src.profession) ?? i18n.t('common.sources.unknown')}</SourceLink>;
		}
		case 'drop': {
			const src = source.drop;
			if (isWorldDrop(src)) return <SourceLink href={ActionId.makeItemUrl(item.id)}>{src.otherName}</SourceLink>;

			const zone = sim.db.getZone(src.zoneId)!;
			const npc = sim.db.getNpc(src.npcId);
			const zoneText = (
				<>
					{zone.name}
					{src.difficulty !== DungeonDifficulty.DifficultyUnknown && ` (${difficultyNames.get(src.difficulty) ?? i18n.t('common.sources.unknown')})`}
				</>
			);

			const category = src.category ? ` - ${src.category}` : '';
			if (npc) {
				return (
					<SourceLink href={ActionId.makeNpcUrl(npc.id)}>
						<span>
							{zoneText}
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
							{zoneText}
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

const OtherSources = ({ entries, ...props }: ItemSourceProps & { entries: ItemSourceEntry[] }) => (
	<ul className="m-0 flex list-none flex-col gap-2 p-0">
		{entries.map((entry, index) => (
			<li key={index}>{renderSource(entry, props)}</li>
		))}
	</ul>
);

const isForOtherFaction = ({ source }: ItemSourceEntry, faction: Faction): boolean => {
	const sourceFaction = source.oneofKind === 'quest' ? source.quest.faction : source.oneofKind === 'soldBy' ? source.soldBy.faction : Faction.Unknown;
	return faction !== Faction.Unknown && sourceFaction !== Faction.Unknown && sourceFaction !== faction;
};

export const ItemSource = ({ item, sim, faction = Faction.Unknown }: ItemSourceProps): ReactNode => {
	const tooltipId = useId();
	if (!item.sources?.length) return null;

	const named = KIND_ORDER.flatMap(kind => item.sources.filter(entry => entry.source.oneofKind === kind)).filter(entry =>
		canShowSource(entry, { item, sim }),
	);
	// PvP is for the filter, so it shows only when the item has nothing else. The other side's quests and vendors stay
	// out, unless they are all the item has: an empty cell would say it comes from nowhere.
	const withoutPvp = named.filter(entry => entry.source.oneofKind !== 'pvp');
	const kept = withoutPvp.length ? withoutPvp : named;
	const own = kept.filter(entry => !isForOtherFaction(entry, faction));
	const [first, ...others] = own.length ? own : kept;
	if (!first) return null;
	if (!others.length) return renderSource(first, { item, sim });

	// The list is a component, so it is built when the tooltip opens: react-tooltip calls a render prop even while
	// hidden, and a row with over a hundred vendors re-renders with the list. The button sits above the row's click layer,
	// as the link beside it does, so a press equips nothing.
	return (
		<div className="flex items-start gap-2">
			{renderSource(first, { item, sim })}
			<Button
				variant="outline-secondary"
				size="none"
				className="relative z-1 shrink-0 px-1 text-xs"
				aria-label={i18n.t('gear_tab.gear_picker.more_sources', { count: others.length })}
				data-testid="item-source-more"
				{...tooltipAnchorProps(tooltipId)}>
				{`+${others.length}`}
			</Button>
			<Tooltip
				id={tooltipId}
				clickable
				place="bottom"
				align="start"
				className="max-h-[70dvh] overflow-y-auto"
				content={<OtherSources entries={others} item={item} sim={sim} />}
			/>
		</div>
	);
};
