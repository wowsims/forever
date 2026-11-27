import { ItemSpec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { Database } from '@sim/proto/database';
import { Stats } from '@sim/proto/stats';
import {
	loadedItemTooltip,
	loadedSetTooltip,
	loadedSpellTooltip,
	renderTooltipText,
	SPELL_TOOLTIP_ID,
	SPELL_TOOLTIP_SELECTOR,
	type SpellTooltipData,
	tooltipFromLink,
	type TooltipPlayer,
	tooltipPlayer,
	type TooltipSegment,
} from '@sim/proto/tooltips';
import { wowheadIconUrl } from '@sim/proto/wowhead';
import { classTalentsConfig } from '@sim/talents/factory';
import { parseTalentsString } from '@sim/talents/talents_string';
import { GameTooltip, type GameTooltipLine } from '@ui-kit/GameTooltip';
import { Tooltip } from '@ui-kit/Tooltip';
import { useMemo } from 'react';

import { enchantText, itemTooltipLines, otherItemTooltipLines } from '../../model/item_tooltip';

/** The one tooltip every spell, item and enchant anchor opens (spellTooltipAnchor, itemTooltipAnchor, enchantTooltipAnchor), with the player's stats, talents and gear worked in. */
export const SpellTooltipHost = () => (
	<Tooltip
		id={SPELL_TOOLTIP_ID}
		anchorSelect={SPELL_TOOLTIP_SELECTOR}
		maxWidth="max-w-96"
		render={({ activeAnchor }) => {
			const attribute = (name: string) => Number(activeAnchor?.getAttribute(name)) || 0;
			const flag = (name: string) => activeAnchor?.getAttribute(name) === 'true';
			// A marked anchor says what it shows; any other is a link to a spell's or an item's page.
			const link = activeAnchor?.hasAttribute('data-tooltip-id') ? {} : tooltipFromLink(activeAnchor?.getAttribute('href') ?? '');
			const itemId = attribute('data-item-tooltip') || link.itemId;
			if (itemId)
				return (
					<ItemTooltip
						itemId={itemId}
						randomSuffix={attribute('data-item-suffix') || link.randomSuffix || 0}
						enchant={attribute('data-item-enchant')}
					/>
				);
			const effectId = attribute('data-enchant-tooltip');
			if (effectId) return <EnchantTooltip effectId={effectId} spellId={attribute('data-enchant-spell')} itemId={attribute('data-enchant-item')} />;
			const spellId = attribute('data-spell-tooltip') || link.spellId;
			if (!spellId) return null;
			return <SpellTooltip spellId={spellId} buff={flag('data-spell-buff')} talent={flag('data-spell-talent')} />;
		}}
	/>
);

/** What a tooltip's numbers read off the player: the sim's final stats, the gear and the points in each talent. */
const usePlayerReader = (): { reader: TooltipPlayer; talentPoints: ReadonlyMap<number, number> } => {
	const player = usePlayer();
	const currentStats = usePlayerStore('currentStats');
	const gear = usePlayerStore('gear');
	const talentsString = usePlayerStore('talentsString');

	const talentPoints = useMemo(() => {
		const config = classTalentsConfig[player.getClass()];
		const points = parseTalentsString(config, talentsString);
		return new Map(config.flatMap((tree, treeIdx) => tree.talents.map((talent, talentIdx) => [talent.spellId, points[treeIdx][talentIdx]] as const)));
	}, [player, talentsString]);

	const reader = useMemo(
		() => tooltipPlayer(currentStats.finalStats ? Stats.fromProto(currentStats.finalStats) : undefined, gear, talentPoints),
		[currentStats, gear, talentPoints],
	);
	return { reader, talentPoints };
};

const iconOf = (data: SpellTooltipData) => (data.i ? wowheadIconUrl(data.i) : undefined);

const SpellTooltip = ({ spellId, buff, talent }: { spellId: number; buff: boolean; talent: boolean }) => {
	const { reader, talentPoints } = usePlayerReader();
	const data = loadedSpellTooltip(spellId);
	if (!data) return null;

	if (talent && data.m) return <TalentTooltip data={data} maxPoints={data.m} points={talentPoints.get(spellId) ?? 0} reader={reader} />;

	const own: Array<TooltipSegment> | undefined = buff && data.a ? data.a : data.d;
	const text = own ? renderTooltipText(own, reader) : '';
	const lines: Array<GameTooltipLine> = [
		...(buff ? [] : (data.h ?? []).filter(([left, right]) => left || right).map(([left, right]) => ({ left, right }))),
		...(text ? [{ left: text, tone: 'gold' as const }] : []),
	];
	return <GameTooltip name={data.n} rank={data.r} iconUrl={iconOf(data)} lines={lines} testId="spell-tooltip" />;
};

interface TalentTooltipProps {
	data: SpellTooltipData;
	maxPoints: number;
	points: number;
	reader: TooltipPlayer;
}

/** As the game shows a talent: the rank taken and its text, then the next rank's; at no points, the first rank's as the next. */
const TalentTooltip = ({ data, maxPoints, points, reader }: TalentTooltipProps) => {
	const atRank = (rank: number) => {
		const segments = data.x ? data.x[rank - 1] : data.d;
		return segments ? renderTooltipText(segments, reader) : '';
	};
	const lines: Array<GameTooltipLine> = [];
	if (points > 0) lines.push({ left: atRank(points), tone: 'gold' });
	if (points < maxPoints && (points === 0 || data.x)) {
		lines.push({ left: i18n.t('talents_tab.tooltip.next_rank'), spaced: points > 0 }, { left: atRank(points + 1), tone: 'gold' });
	}
	return (
		<GameTooltip
			name={data.n}
			rank={i18n.t('talents_tab.tooltip.rank', { rank: points, max: maxPoints })}
			iconUrl={iconOf(data)}
			lines={lines}
			testId="talent-tooltip"
		/>
	);
};

interface ItemTooltipProps {
	itemId: number;
	randomSuffix: number;
	/** The enchant's effect id. */
	enchant: number;
}

const ItemTooltip = ({ itemId, randomSuffix, enchant }: ItemTooltipProps) => {
	const { reader } = usePlayerReader();
	const gear = usePlayerStore('gear');
	const db = Database.getSync();
	const equipped = useMemo(() => db.lookupItemSpec(ItemSpec.create({ id: itemId, randomSuffix, enchant })), [db, itemId, randomSuffix, enchant]);
	if (!equipped) return <OtherItemTooltip itemId={itemId} reader={reader} />;

	const worn = new Set(gear.asArray().flatMap(item => (item ? [item.id] : [])));
	const lines = itemTooltipLines({
		equipped,
		data: loadedItemTooltip(itemId),
		set: equipped.item.setId ? loadedSetTooltip(equipped.item.setId) : undefined,
		worn,
		itemName: id => db.getItemById(id)?.name ?? '',
		spellText: spellId => {
			const segments = loadedSpellTooltip(spellId)?.d;
			return segments?.length ? renderTooltipText(segments, reader) : undefined;
		},
	});
	const name = equipped.randomSuffix ? `${equipped.item.name} ${equipped.randomSuffix.name}` : equipped.item.name;
	return <GameTooltip name={name} quality={equipped.item.quality} iconUrl={wowheadIconUrl(equipped.item.icon)} lines={lines} testId="item-tooltip" />;
};

const spellTextOf = (spellId: number, reader: TooltipPlayer) => {
	const segments = loadedSpellTooltip(spellId)?.d;
	return segments?.length ? renderTooltipText(segments, reader) : undefined;
};

/** An item the item database does not hold, a consumable or an enchant's item, from the tooltips alone. */
const OtherItemTooltip = ({ itemId, reader }: { itemId: number; reader: TooltipPlayer }) => {
	const data = loadedItemTooltip(itemId);
	if (!data?.n) return null;
	return (
		<GameTooltip
			name={data.n}
			quality={data.q}
			iconUrl={data.i ? wowheadIconUrl(data.i) : undefined}
			lines={otherItemTooltipLines(data, spellId => spellTextOf(spellId, reader))}
			testId="item-tooltip"
		/>
	);
};

interface EnchantTooltipProps {
	effectId: number;
	spellId: number;
	itemId: number;
}

/** An enchant: its name, what it adds, and the text of the spell or item that applies it. */
const EnchantTooltip = ({ effectId, spellId, itemId }: EnchantTooltipProps) => {
	const { reader } = usePlayerReader();
	const rows = Database.getSync().getEnchantsByEffectId(effectId);
	const enchant = rows.find(row => row.spellId === spellId && row.itemId === itemId) ?? rows[0];
	if (!enchant) return null;
	const itemText = enchant.itemId ? loadedItemTooltip(enchant.itemId)?.e?.find(([trigger]) => trigger === 0) : undefined;
	const text = spellTextOf(enchant.spellId, reader) ?? (itemText && spellTextOf(itemText[1], reader));
	const icon = loadedSpellTooltip(enchant.spellId)?.i ?? (enchant.itemId ? loadedItemTooltip(enchant.itemId)?.i : undefined) ?? enchant.icon;
	return (
		<GameTooltip
			name={enchant.name}
			quality={enchant.quality}
			iconUrl={icon ? wowheadIconUrl(icon) : undefined}
			lines={[{ left: enchantText(enchant), tone: 'green' }, ...(text ? [{ left: text, tone: 'gold' as const }] : [])]}
			testId="enchant-tooltip"
		/>
	);
};
