export interface GameTooltipProps {
	name: string;
	/** Beside the name, such as "Rank 5" or a talent's "Rank 2/5". */
	rank?: string;
	/** The cost and range, then the cast time and cooldown, as [left, right]. */
	lines?: ReadonlyArray<readonly [string, string]>;
	/** The description, with its line breaks. */
	text?: string;
	/** A talent's next rank, under its own heading. */
	next?: { label: string; text: string };
	testId?: string;
}

/** A spell's tooltip laid out the way the game lays one out: name and rank, the cost and timing lines, then the text in gold. */
export const GameTooltip = ({ name, rank, lines, text, next, testId }: GameTooltipProps) => (
	<div className="flex max-w-80 flex-col gap-1 p-2 text-left text-sm" data-testid={testId}>
		<div className="flex items-baseline justify-between gap-4">
			<span className="font-semibold text-white">{name}</span>
			{rank && <span className="text-muted">{rank}</span>}
		</div>
		{lines
			?.filter(([left, right]) => left || right)
			.map(([left, right]) => (
				<div key={`${left}|${right}`} className="flex justify-between gap-4 text-white">
					<span>{left}</span>
					<span>{right}</span>
				</div>
			))}
		{text && <p className="m-0 whitespace-pre-line text-brand">{text}</p>}
		{next && (
			<>
				<p className="m-0 mt-1 text-white">{next.label}</p>
				<p className="m-0 whitespace-pre-line text-brand">{next.text}</p>
			</>
		)}
	</div>
);
