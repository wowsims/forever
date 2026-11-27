import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { usePlayerStore } from '@sim/hooks/usePlayerStore';
import { Stats } from '@sim/proto/stats';
import { loadedSpellTooltip, renderTooltipText, SPELL_TOOLTIP_ID, type TooltipPlayer, tooltipPlayer, type TooltipSegment } from '@sim/proto/tooltips';
import { classTalentsConfig } from '@sim/talents/factory';
import { parseTalentsString } from '@sim/talents/talents_string';
import { GameTooltip } from '@ui-kit/GameTooltip';
import { Tooltip } from '@ui-kit/Tooltip';
import { useMemo } from 'react';

/** The one tooltip every spell anchor opens (spellTooltipAnchor): the spell's text with the player's stats and talents worked in. */
export const SpellTooltipHost = () => (
	<Tooltip
		id={SPELL_TOOLTIP_ID}
		maxWidth="max-w-80"
		render={({ activeAnchor }) => {
			const spellId = Number(activeAnchor?.getAttribute('data-spell-tooltip'));
			if (!spellId) return null;
			return <SpellTooltip spellId={spellId} buff={activeAnchor?.getAttribute('data-spell-buff') === 'true'} />;
		}}
	/>
);

const SpellTooltip = ({ spellId, buff }: { spellId: number; buff: boolean }) => {
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

	const data = loadedSpellTooltip(spellId);
	if (!data) return null;
	const text = (segments: Array<TooltipSegment> | undefined) => (segments ? renderTooltipText(segments, reader) : undefined);

	if (data.m) return <TalentTooltip name={data.n} maxPoints={data.m} points={talentPoints.get(spellId) ?? 0} ranks={data.x} text={data.d} reader={reader} />;

	const own = buff && data.a ? data.a : data.d;
	return <GameTooltip name={data.n} rank={data.r} lines={buff ? undefined : data.h} text={text(own)} testId="spell-tooltip" />;
};

interface TalentTooltipProps {
	name: string;
	maxPoints: number;
	points: number;
	/** The text at each rank, where the ranks read differently; text is every rank's otherwise. */
	ranks?: Array<Array<TooltipSegment>>;
	text?: Array<TooltipSegment>;
	reader: TooltipPlayer;
}

/** As the game shows a talent: the rank taken and its text, then the next rank's; at no points, the first rank's as the next. */
const TalentTooltip = ({ name, maxPoints, points, ranks, text, reader }: TalentTooltipProps) => {
	const atRank = (rank: number) => {
		const segments = ranks ? ranks[rank - 1] : text;
		return segments ? renderTooltipText(segments, reader) : '';
	};
	const nextRank = points < maxPoints && (points === 0 || ranks) ? atRank(points + 1) : undefined;
	return (
		<GameTooltip
			name={name}
			rank={i18n.t('talents_tab.tooltip.rank', { rank: points, max: maxPoints })}
			text={points > 0 ? atRank(points) : undefined}
			next={nextRank ? { label: i18n.t('talents_tab.tooltip.next_rank'), text: nextRank } : undefined}
			testId="talent-tooltip"
		/>
	);
};
