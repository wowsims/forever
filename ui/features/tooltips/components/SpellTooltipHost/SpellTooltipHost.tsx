import { ItemSpec } from '@generated/proto/common';
import type { UIEnchant } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { ActionId } from '@sim/proto/action_id';
import { Database } from '@sim/proto/database';
import { Stats } from '@sim/proto/stats';
import {
	loadedEnchantName,
	loadedItemTooltip,
	loadedSetTooltip,
	loadedSpellTooltip,
	renderTooltipHeader,
	renderTooltipText,
	SPELL_TOOLTIP_ID,
	SPELL_TOOLTIP_SELECTOR,
	type SpellTooltipData,
	subscribeTooltips,
	tooltipFromLink,
	type TooltipPlayer,
	tooltipPlayer,
	type TooltipSegment,
	tooltipsVersion,
} from '@sim/proto/tooltips';
import { WOWHEAD_BASE_URL, wowheadIconUrl } from '@sim/proto/wowhead';
import { classTalentsConfig } from '@sim/talents/factory';
import { parseTalentsString } from '@sim/talents/talents_string';
import { GameTooltip, type GameTooltipLine, type GameTooltipProps } from '@ui-kit/GameTooltip';
import { Tooltip } from '@ui-kit/Tooltip';
import { useMemo, useSyncExternalStore } from 'react';

import { enchantText, itemTooltipLines, otherItemTooltipLines } from '../../model/item_tooltip';

/** The one tooltip every spell, item and enchant anchor opens (spellTooltipAnchor, itemTooltipAnchor, enchantTooltipAnchor), with the player's stats, talents and gear worked in. */
export const SpellTooltipHost = () => (
	<Tooltip
		id={SPELL_TOOLTIP_ID}
		anchorSelect={SPELL_TOOLTIP_SELECTOR}
		maxWidth="max-w-96"
		carriesLink
		render={({ activeAnchor }) => {
			const attribute = (name: string) => Number(activeAnchor?.getAttribute(name)) || 0;
			const flag = (name: string) => activeAnchor?.getAttribute(name) === 'true';
			// A marked anchor says what it shows; any other is a link to a spell's or an item's page.
			const link = activeAnchor?.hasAttribute('data-tooltip-id') ? {} : tooltipFromLink(activeAnchor?.getAttribute('href') ?? '');
			// A spell's or an enchant's tooltip links where its anchor does, so the two open one page: a talent's at its rank.
			const href = activeAnchor?.getAttribute('href') ?? '';
			const anchorPage = href.startsWith(`${WOWHEAD_BASE_URL}/`) ? href : undefined;
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
			if (effectId)
				return (
					<EnchantTooltip
						effectId={effectId}
						spellId={attribute('data-enchant-spell')}
						itemId={attribute('data-enchant-item')}
						pageUrl={anchorPage}
					/>
				);
			const spellId = attribute('data-spell-tooltip') || link.spellId;
			if (!spellId) return null;
			return (
				<SpellTooltip
					spellId={spellId}
					buff={flag('data-spell-buff')}
					talent={flag('data-spell-talent')}
					pageUrl={anchorPage ?? ActionId.makeSpellUrl(spellId)}
				/>
			);
		}}
	/>
);

/**
 * What a tooltip's numbers read off the player: the sim's final stats, done multipliers and starting form, the class, the
 * gear and the points in each talent. It also renders the tooltip again when a tooltip file arrives, since an id the
 * first file lacks waits for the second.
 */
const usePlayerReader = (): { reader: TooltipPlayer; talentPoints: ReadonlyMap<number, number> } => {
	useSyncExternalStore(subscribeTooltips, tooltipsVersion);
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
		() =>
			tooltipPlayer(currentStats.finalStats ? Stats.fromProto(currentStats.finalStats) : undefined, gear, talentPoints, {
				done: currentStats.doneMultipliers,
				playerClass: player.getClass(),
				form: currentStats.shapeshiftForm,
			}),
		[currentStats, gear, talentPoints, player],
	);
	return { reader, talentPoints };
};

const iconOf = (data: SpellTooltipData) => (data.i ? wowheadIconUrl(data.i) : undefined);

const wowheadLink = (href: string) => ({ href, label: i18n.t('tooltips.wowhead_link') });

const SpellTooltip = ({ spellId, buff, talent, pageUrl }: { spellId: number; buff: boolean; talent: boolean; pageUrl: string }) => {
	const { reader, talentPoints } = usePlayerReader();
	const data = loadedSpellTooltip(spellId);
	if (!data) return null;

	if (talent && data.m) return <TalentTooltip data={data} maxPoints={data.m} points={talentPoints.get(spellId) ?? 0} reader={reader} pageUrl={pageUrl} />;

	const own: Array<TooltipSegment> | undefined = buff && data.a ? data.a : data.d;
	const text = own ? renderTooltipText(own, reader) : '';
	const lines: Array<GameTooltipLine> = [
		...(buff ? [] : renderTooltipHeader(data.h ?? [], reader).map(([left, right]) => ({ left, right }))),
		...(text ? [{ left: text, tone: 'gold' as const }] : []),
	];
	return <GameTooltip name={data.n} rank={data.r} iconUrl={iconOf(data)} lines={lines} link={wowheadLink(pageUrl)} testId="spell-tooltip" />;
};

interface TalentTooltipProps {
	data: SpellTooltipData;
	maxPoints: number;
	points: number;
	reader: TooltipPlayer;
	pageUrl: string;
}

/** As the game shows a talent: the rank taken and its text, then the next rank's; at no points, the first rank's as the next. */
const TalentTooltip = ({ data, maxPoints, points, reader, pageUrl }: TalentTooltipProps) => {
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
			link={wowheadLink(pageUrl)}
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
	// The suffix's stats join the item's own only in its dynamic stats.
	const equipped = useMemo(
		() => db.lookupItemSpec(ItemSpec.create({ id: itemId, randomSuffix, enchant }))?.withDynamicStats(),
		[db, itemId, randomSuffix, enchant],
	);
	// The page shows the item with its random suffix and enchant.
	const link = wowheadLink(ActionId.makeItemUrl(itemId, randomSuffix, enchant));
	if (!equipped) return <OtherItemTooltip itemId={itemId} reader={reader} link={link} />;

	const worn = new Set(gear.asArray().flatMap(item => (item ? [item.id] : [])));
	const lines = itemTooltipLines({
		equipped,
		data: loadedItemTooltip(itemId),
		set: equipped.item.setId ? loadedSetTooltip(equipped.item.setId) : undefined,
		worn,
		enchantName: equipped.enchant ? loadedEnchantName(equipped.enchant.effectId) : undefined,
		enchantIcon: equipped.enchant ? enchantIconOf(equipped.enchant) : undefined,
		itemName: id => db.getItemById(id)?.name ?? '',
		spellText: spellId => {
			const segments = loadedSpellTooltip(spellId)?.d;
			return segments?.length ? renderTooltipText(segments, reader) : undefined;
		},
	});
	const name = equipped.randomSuffix ? `${equipped.item.name} ${equipped.randomSuffix.name}` : equipped.item.name;
	return (
		<GameTooltip name={name} quality={equipped.item.quality} iconUrl={wowheadIconUrl(equipped.item.icon)} lines={lines} link={link} testId="item-tooltip" />
	);
};

// The icon the game files give a spell without one of its own, as a scope's enchant spell.
const PLACEHOLDER_ICON = 'temp';

/** An enchant's icon: that of the spell or item that applies it, else the database's own. */
const enchantIconOf = (enchant: UIEnchant) => {
	const icons = [loadedSpellTooltip(enchant.spellId)?.i, enchant.itemId ? loadedItemTooltip(enchant.itemId)?.i : undefined, enchant.icon];
	const icon = icons.find(name => name && name !== PLACEHOLDER_ICON);
	return icon ? wowheadIconUrl(icon) : undefined;
};

const spellTextOf = (spellId: number, reader: TooltipPlayer) => {
	const segments = loadedSpellTooltip(spellId)?.d;
	return segments?.length ? renderTooltipText(segments, reader) : undefined;
};

/** An item the item database does not hold, a consumable or an enchant's item, from the tooltips alone. */
const OtherItemTooltip = ({ itemId, reader, link }: { itemId: number; reader: TooltipPlayer; link: GameTooltipProps['link'] }) => {
	const data = loadedItemTooltip(itemId);
	if (!data?.n) return null;
	return (
		<GameTooltip
			name={data.n}
			quality={data.q}
			iconUrl={data.i ? wowheadIconUrl(data.i) : undefined}
			lines={otherItemTooltipLines(data, spellId => spellTextOf(spellId, reader))}
			link={link}
			testId="item-tooltip"
		/>
	);
};

interface EnchantTooltipProps {
	effectId: number;
	spellId: number;
	itemId: number;
	/** The anchor's link; without one, the page of the spell or item that applies the enchant. */
	pageUrl?: string;
}

/** An enchant: its name, what it adds, and the text of the spell or item that applies it. */
const EnchantTooltip = ({ effectId, spellId, itemId, pageUrl }: EnchantTooltipProps) => {
	const { reader } = usePlayerReader();
	const rows = Database.getSync().getEnchantsByEffectId(effectId);
	const enchant = rows.find(row => row.spellId === spellId && row.itemId === itemId) ?? rows[0];
	if (!enchant) return null;
	const itemText = enchant.itemId ? loadedItemTooltip(enchant.itemId)?.e?.find(([trigger]) => trigger === 0) : undefined;
	const text = spellTextOf(enchant.spellId, reader) ?? (itemText && spellTextOf(itemText[1], reader));
	return (
		<GameTooltip
			name={enchant.name}
			quality={enchant.quality}
			iconUrl={enchantIconOf(enchant)}
			lines={[
				{ left: loadedEnchantName(enchant.effectId) ?? enchantText(enchant), tone: 'green' },
				...(text ? [{ left: text, tone: 'gold' as const }] : []),
			]}
			link={wowheadLink(pageUrl ?? (enchant.spellId ? ActionId.makeSpellUrl(enchant.spellId) : ActionId.makeItemUrl(enchant.itemId)))}
			testId="enchant-tooltip"
		/>
	);
};
