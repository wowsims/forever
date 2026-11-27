import type { ItemQuality } from '@generated/proto/common';
import clsx from 'clsx';

import { QUALITY_TEXT } from '../utils/colors';

/** The game's tooltip colours: white for facts, gold for a spell's text, green for what gear adds, grey for what is not active. */
export type GameTooltipTone = 'white' | 'gold' | 'green' | 'grey';

export interface GameTooltipLine {
	left?: string;
	/** Right-aligned on the same line, such as a cooldown or a weapon's speed. */
	right?: string;
	tone?: GameTooltipTone;
	/** A gap above the line, between the tooltip's blocks. */
	spaced?: boolean;
	/** Set in from the left, as a set's pieces are under its name. */
	indent?: boolean;
}

export interface GameTooltipProps {
	name: string;
	/** The name's colour for an item; white without one. */
	quality?: ItemQuality;
	/** Beside the name, such as "Rank 5" or a talent's "Rank 2/5". */
	rank?: string;
	iconUrl?: string;
	lines?: ReadonlyArray<GameTooltipLine>;
	testId?: string;
}

const TONE: Record<GameTooltipTone, string> = {
	white: 'text-white',
	gold: 'text-brand',
	green: 'text-success',
	grey: 'text-muted',
};

/** A spell's or an item's tooltip laid out the way the game lays one out, with its icon beside it. */
export const GameTooltip = ({ name, quality, rank, iconUrl, lines, testId }: GameTooltipProps) => (
	<div className="flex items-start gap-2 p-2 text-left text-sm" data-testid={testId}>
		{iconUrl && <span className="size-10 shrink-0 border border-surface-border bg-cover bg-center" style={{ backgroundImage: `url('${iconUrl}')` }} />}
		<div className="flex min-w-0 flex-col">
			<div className="flex items-baseline justify-between gap-4">
				<span className={clsx('font-semibold', quality === undefined ? 'text-white' : QUALITY_TEXT[quality])}>{name}</span>
				{rank && <span className="shrink-0 text-muted">{rank}</span>}
			</div>
			{lines?.map((line, index) => (
				<div
					key={index}
					className={clsx(
						'flex justify-between gap-4 whitespace-pre-line',
						TONE[line.tone ?? 'white'],
						line.spaced && 'mt-1',
						line.indent && 'pl-4',
					)}>
					<span>{line.left}</span>
					{line.right && <span className="shrink-0">{line.right}</span>}
				</div>
			))}
		</div>
	</div>
);
