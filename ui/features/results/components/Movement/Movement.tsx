import { Slider } from '@base-ui/react/slider';
import i18n from '@i18n/config';
import { MAX_MELEE_RANGE, MIN_RANGED_RANGE } from '@sim/proto/sim_tables_auto_gen';
import { MAX_PARTY_SIZE } from '@sim/raid/party';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { ToggleGroup } from '@ui-kit/ToggleGroup';
import { type PointerEvent, useCallback, useEffect, useMemo, useState } from 'react';

import { type CanvasPaint, useCanvasPainter } from '../../hooks/useCanvasPainter';
import { useSimResult } from '../../hooks/useSimResult';
import { buildMovementRun, type MovementRun, positionAt, type RangeZone, rangeZone } from '../../model/movement/movement';
import type { SimResultData } from '../../model/result_data';
import { resultKey } from '../Timeline/utils';
import { paintField, paintStrip, readPalette } from './paint';

const t = (key: string, options?: Record<string, unknown>) => i18n.t(`results_tab.details.movement.${key}`, options);

const SPEEDS = ['0.5', '1', '2', '4'] as const;
type Speed = (typeof SPEEDS)[number];

const ZONE_KEYS: Record<RangeZone, string> = { melee: 'melee', 'dead-zone': 'dead_zone', ranged: 'ranged' };

// Seconds per slider step, and per step with Shift or Page Up and Down.
const SLIDER_STEP = 0.1;
const SLIDER_LARGE_STEP = 1;

// A ring's label drops a trailing zero: 5, not 5.0.
const ringYards = (yards: number) => String(Number(yards.toFixed(1)));

interface MovementPlaybackProps {
	run: MovementRun;
	targetName: string;
	active: boolean;
}

const MovementPlayback = ({ run, targetName, active }: MovementPlaybackProps) => {
	const [time, setTime] = useState(run.start);
	const [playing, setPlaying] = useState(false);
	const [speed, setSpeed] = useState<Speed>('1');

	// Plays on while the tab is open; a closed tab holds the time where it was.
	useEffect(() => {
		if (!playing || !active) return;
		let last: number | null = null;
		let frame = requestAnimationFrame(function step(now: number) {
			if (last !== null) {
				const elapsed = ((now - last) / 1000) * Number(speed);
				setTime(current => Math.min(run.end, current + elapsed));
			}
			last = now;
			frame = requestAnimationFrame(step);
		});
		return () => cancelAnimationFrame(frame);
	}, [playing, active, speed, run.end]);

	useEffect(() => {
		if (playing && time >= run.end) setPlaying(false);
	}, [playing, time, run.end]);

	const togglePlaying = () => {
		if (playing) {
			setPlaying(false);
			return;
		}
		if (time >= run.end) setTime(run.start);
		setPlaying(true);
	};

	const yards = useMemo(() => run.tracks.map(track => positionAt(track, time)), [run, time]);
	const player = run.tracks[0];
	const zone = rangeZone(yards[0]);

	const fieldPaint = useCallback<CanvasPaint>(
		(ctx, width, height) =>
			paintField(ctx, width, height, readPalette(ctx.canvas), {
				run,
				yards,
				unitLabels: run.tracks.map((track, index) => t('unit', { name: track.name, yards: yards[index].toFixed(1) })),
				targetLabel: targetName,
				meleeLabel: t('rings.melee', { yards: ringYards(MAX_MELEE_RANGE) }),
				rangedLabel: t('rings.ranged', { yards: ringYards(MIN_RANGED_RANGE) }),
			}),
		[run, yards, targetName],
	);
	const stripPaint = useCallback<CanvasPaint>((ctx, width, height) => paintStrip(ctx, width, height, readPalette(ctx.canvas), run, time), [run, time]);
	const fieldRef = useCanvasPainter(fieldPaint);
	const stripRef = useCanvasPainter(stripPaint);

	const scrub = (event: PointerEvent<HTMLCanvasElement>) => {
		const box = event.currentTarget.getBoundingClientRect();
		if (box.width === 0) return;
		const share = Math.min(1, Math.max(0, (event.clientX - box.left) / box.width));
		setTime(run.start + share * (run.end - run.start));
	};

	const timeText = t('time', { time: time.toFixed(2) });

	return (
		<div className="flex flex-col gap-2">
			<p>{t('note')}</p>
			<div className="relative h-128">
				<canvas ref={fieldRef} role="img" aria-label={t('field_label')} className="absolute inset-0 size-full text-xs" />
			</div>
			<div className="flex flex-wrap items-center gap-2">
				<Button variant="outline-primary" size="xs" aria-label={playing ? t('pause') : t('play')} onClick={togglePlaying}>
					<Icon name={playing ? 'pause' : 'play'} fixedWidth />
				</Button>
				<ToggleGroup
					size="xs"
					aria-label={t('speed_label')}
					value={speed}
					onChange={setSpeed}
					options={SPEEDS.map(value => ({ value, label: t('speed', { speed: value }) }))}
				/>
				<p className="ml-2 flex flex-wrap items-center gap-x-4 tabular-nums">
					<span>{timeText}</span>
					<span>{t('unit', { name: player.name, yards: yards[0].toFixed(2) })}</span>
					<span data-zone={zone} className="data-[zone=dead-zone]:text-danger">
						{t(`zones.${ZONE_KEYS[zone]}`)}
					</span>
				</p>
			</div>
			<Slider.Root
				value={time}
				min={run.start}
				max={run.end}
				step={SLIDER_STEP}
				largeStep={SLIDER_LARGE_STEP}
				onValueChange={value => setTime(value)}
				className="w-full">
				<Slider.Control className="flex h-6 w-full touch-none items-center select-none">
					<Slider.Track className="h-1 w-full rounded-full bg-surface-border">
						<Slider.Indicator className="rounded-full bg-primary" />
						<Slider.Thumb
							getAriaLabel={() => t('time_label')}
							getAriaValueText={() => timeText}
							className="size-4 rounded-full bg-foreground focus-ring"
						/>
					</Slider.Track>
				</Slider.Control>
			</Slider.Root>
			<div className="relative h-16">
				<canvas
					ref={stripRef}
					aria-hidden
					className="absolute inset-0 size-full cursor-pointer touch-none text-xs"
					onPointerDown={event => {
						event.currentTarget.setPointerCapture(event.pointerId);
						scrub(event);
					}}
					onPointerMove={event => {
						if (event.currentTarget.hasPointerCapture(event.pointerId)) scrub(event);
					}}
				/>
			</div>
		</div>
	);
};

export interface MovementViewProps {
	/** The run's movement, or null when no unit walked in it. */
	run: MovementRun | null;
	/** The name drawn under the target in the middle. */
	targetName: string;
	/** The movement tab is open; playback holds while it is not. */
	active: boolean;
}

export const MovementView = ({ run, targetName, active }: MovementViewProps) =>
	run ? <MovementPlayback run={run} targetName={targetName} active={active} /> : <p className="p-6 text-center text-base">{t('no_walks')}</p>;

const movementOf = ({ result, filter }: SimResultData): { run: MovementRun; targetName: string } | null => {
	const player = result.getRaidIndexedPlayers(filter)[0];
	if (!player) return null;
	const settings = result.request.raid?.parties[Math.floor(player.index / MAX_PARTY_SIZE)]?.players[player.index % MAX_PARTY_SIZE];
	const run = buildMovementRun({ player, playerStartYards: settings?.distanceFromTarget ?? 0, logs: result.logs, duration: result.firstIterationDuration });
	return run && { run, targetName: result.getTargets()[0]?.name ?? '' };
};

export interface MovementProps {
	/** The movement tab is open. A run that lands while it is closed waits for it to open. */
	active: boolean;
}

/** Where the player and its pets stand around the target over the run the timeline shows, with a slider and playback. */
export const Movement = ({ active }: MovementProps) => {
	const live = useSimResult();
	const [shown, setShown] = useState<SimResultData | null>(null);

	useEffect(() => {
		if (!active || !live) return;
		setShown(prev => (resultKey(prev) === resultKey(live) ? prev : live));
	}, [active, live]);

	const movement = useMemo(() => (shown ? movementOf(shown) : null), [shown]);
	if (!shown) return null;
	// A new run starts its playback over from the prepull.
	return <MovementView key={resultKey(shown)} run={movement?.run ?? null} targetName={movement?.targetName ?? ''} active={active} />;
};
