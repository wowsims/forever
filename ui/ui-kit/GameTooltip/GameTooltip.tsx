import type { ItemQuality } from '@generated/proto/common';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';

import { QUALITY_TEXT } from '../utils/colors';

/**
 * The game's tooltip colours: white for facts, gold for a spell's text and a set's name, green for what gear adds,
 * grey for what is not active, and light yellow for the set pieces the player wears.
 */
export type GameTooltipTone = 'white' | 'gold' | 'green' | 'grey' | 'yellow';

export interface GameTooltipLine {
	left?: string;
	/** Right-aligned on the same line, such as a cooldown or a weapon's speed. */
	right?: string;
	tone?: GameTooltipTone;
	/** A blank line above the line, between the tooltip's blocks. */
	spaced?: boolean;
	/** Set in from the left, as a set's pieces are under its name. */
	indent?: boolean;
	/** A small icon before the text, such as an enchant's. */
	icon?: string;
	/** An amount of money in copper after the text, in gold, silver and copper coins. */
	money?: number;
}

export interface GameTooltipProps {
	name: string;
	/** The name's colour for an item; white without one. */
	quality?: ItemQuality;
	/** Beside the name, such as "Rank 5" or a talent's "Rank 2/5". */
	rank?: string;
	iconUrl?: string;
	lines?: ReadonlyArray<GameTooltipLine>;
	/** The page of what the tooltip shows, opened in a new tab from a line on the edge that faces the anchor. */
	link?: { href: string; label: string };
	testId?: string;
}

const TONE: Record<GameTooltipTone, string> = {
	white: 'text-white',
	gold: 'text-gold',
	green: 'text-success',
	grey: 'text-grey',
	yellow: 'text-light-yellow',
};

const COINS: ReadonlyArray<[copper: number, className: string]> = [
	[10000, 'bg-gold'],
	[100, 'bg-silver'],
	[1, 'bg-copper'],
];

/** Each coin with any, as the game writes a price: 7 gold 30 silver 36 copper, or 15 silver. */
const Money = ({ copper }: { copper: number }) => (
	<span className="ml-1 inline-flex gap-2">
		{COINS.map(([unit, className], index) => {
			const count = index === 0 ? Math.floor(copper / unit) : Math.floor(copper / unit) % 100;
			return count > 0 ? (
				<span key={unit} className="inline-flex items-center gap-1">
					{count}
					<span className={clsx('size-2 rounded-full', className)} />
				</span>
			) : null;
		})}
	</span>
);

/** A spell's or an item's tooltip laid out the way the game lays one out, with its icon beside it. */
export const GameTooltip = ({ name, quality, rank, iconUrl, lines, link, testId }: GameTooltipProps) => (
	<div className={clsx('grid items-start gap-2 p-2 text-left text-sm', iconUrl && 'grid-cols-[auto_1fr]')} data-testid={testId}>
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
						line.spaced && 'mt-4',
						line.indent && 'pl-4',
					)}>
					<span>
						{line.icon && (
							<span className="mr-1 inline-block size-4 bg-cover bg-center align-middle" style={{ backgroundImage: `url('${line.icon}')` }} />
						)}
						{line.left}
						{line.money !== undefined && <Money copper={line.money} />}
					</span>
					{line.right && <span className="shrink-0">{line.right}</span>}
				</div>
			))}
		</div>
		{/*
		 * Centred on the edge that faces the anchor, with its box reaching that edge. The tooltip is centred on its anchor and
		 * touches it, so the pointer moves straight from the anchor onto the link without crossing another anchor. The link stays
		 * short, so a pointer moving on to the centre of an anchor the tooltip covers passes it by.
		 */}
		{link && (
			<Button
				variant="unstyled"
				href={link.href}
				target="_blank"
				className="col-span-full -mb-2 flex items-center gap-1 justify-self-center pb-1 text-xs text-muted hover:text-white in-[.react-tooltip\_\_place-bottom]:order-first in-[.react-tooltip\_\_place-bottom]:-mt-2 in-[.react-tooltip\_\_place-bottom]:mb-0 in-[.react-tooltip\_\_place-bottom]:pt-1 in-[.react-tooltip\_\_place-bottom]:pb-0">
				{link.label}
				<Icon name="arrow-up-right-from-square" />
			</Button>
		)}
	</div>
);
