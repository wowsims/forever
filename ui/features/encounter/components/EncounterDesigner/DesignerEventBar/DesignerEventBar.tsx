import { usePointerAxisDrag } from '@ui-kit/hooks/usePointerAxisDrag';
import { Icon } from '@ui-kit/Icon';
import { cssVars } from '@ui-kit/utils/css';
import type { KeyboardEvent as ReactKeyboardEvent, Ref } from 'react';
import { useRef, useState } from 'react';

import type { TimeWindow } from '../../../model/design/time';
import { fractionToSeconds, secondsToFraction, snap } from '../../../model/design/time';
import { designerText } from '../utils/configs';
import { barKeyAction, editWindow } from './utils';

export type DesignerEventBarShape = 'span' | 'dot' | 'diamond';

export interface DesignerEventBarProps {
	name: string;
	/** A point event has `start === end`. Fractions of the fight. */
	window: TimeWindow;
	/** The lane row, counted from 0, when the bar sits in a lane's grid. */
	row: number;
	shape: DesignerEventBarShape;
	/** The jitter around the start and the end, as fractions of the fight. */
	jitter?: { start: number; end: number };
	baseSeconds: number;
	/** The drag and keyboard step; 0 drags freely and steps by 1 s. */
	snapSeconds: number;
	pixelsPerSecond: () => number;
	valueText: (window: TimeWindow) => string;
	selected?: boolean;
	invalid?: boolean;
	tabIndex: 0 | -1;
	onChange: (window: TimeWindow) => void;
	onSelect: () => void;
	/** Enter, after `onSelect`; a click or a drag only selects. */
	onActivate: () => void;
	onRemove: () => void;
	onSibling: (direction: -1 | 1) => void;
	onFocus?: () => void;
	ref?: Ref<HTMLDivElement>;
}

const MIN_WIDTH_SECONDS = 1;

const SPAN_CLASSES =
	'top-1 bottom-1 flex w-timeline-segment min-w-2 items-center rounded-sm border border-transparent bg-(--bar) data-selected:border-white data-invalid:border-dashed data-invalid:not-data-selected:border-danger';
const POINT_CLASSES = 'top-1/2 size-(--designer-keyframe-size) -translate-1/2';

// The shape is its own element, so the diamond's rotation stays off the whisker beside it.
const POINT_SHAPE_CLASSES: Record<Exclude<DesignerEventBarShape, 'span'>, string> = {
	dot: 'rounded-full',
	diamond: 'rotate-45',
};

const Whisker = ({ seconds, at }: { seconds: number; at: 'start' | 'end' | 'center' }) =>
	seconds > 0 ? (
		<span
			aria-hidden="true"
			data-at={at}
			className="pointer-events-none absolute top-1/2 h-px w-timeline-segment -translate-x-1/2 bg-(--designer-jitter-whisker) data-[at=center]:left-1/2 data-[at=end]:left-full data-[at=start]:left-0"
			style={cssVars({ '--dur': String(2 * seconds) })}
		/>
	) : null;

export const DesignerEventBar = ({
	name,
	window,
	row,
	shape,
	jitter,
	baseSeconds,
	snapSeconds,
	pixelsPerSecond,
	valueText,
	selected,
	invalid,
	tabIndex,
	onChange,
	onSelect,
	onActivate,
	onRemove,
	onSibling,
	onFocus,
	ref,
}: DesignerEventBarProps) => {
	const [preview, setPreview] = useState<TimeWindow | null>(null);
	const origin = useRef(window);
	const latest = useRef<TimeWindow | null>(null);
	const minWidth = secondsToFraction(MIN_WIDTH_SECONDS, baseSeconds);
	const span = shape === 'span';

	const drag = usePointerAxisDrag({
		pixelsPerSecond,
		onStart: () => {
			origin.current = window;
			onSelect();
		},
		onDelta: (seconds, mode) => {
			const next = editWindow(origin.current, span ? mode : 'move', secondsToFraction(snap(seconds, snapSeconds), baseSeconds), minWidth);
			latest.current = next;
			setPreview(next);
		},
		onEnd: commit => {
			const next = latest.current;
			latest.current = null;
			setPreview(null);
			if (commit && next && (next.start !== window.start || next.end !== window.end)) onChange(next);
		},
	});

	const onKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
		const action = barKeyAction(event, !span);
		if (!action) return;
		event.preventDefault();
		event.stopPropagation();
		switch (action.type) {
			case 'edit': {
				const step = snapSeconds > 0 ? snapSeconds : 1;
				const next = editWindow(window, action.mode, secondsToFraction(step * action.steps, baseSeconds), minWidth);
				if (next.start !== window.start || next.end !== window.end) onChange(next);
				return;
			}
			case 'remove':
				onRemove();
				return;
			case 'activate':
				onSelect();
				onActivate();
				return;
			case 'sibling':
				onSibling(action.direction);
		}
	};

	const shown = preview ?? window;
	const start = fractionToSeconds(shown.start, baseSeconds);
	const startJitter = fractionToSeconds(jitter?.start ?? 0, baseSeconds);
	const endJitter = fractionToSeconds(jitter?.end ?? 0, baseSeconds);
	const state = { 'data-selected': selected ? '' : undefined, 'data-invalid': invalid ? '' : undefined };
	const valueTextWithState = [valueText(shown), selected && designerText('state_selected'), invalid && designerText('state_invalid')]
		.filter(Boolean)
		.join(', ');

	return (
		<div
			ref={ref}
			role="slider"
			tabIndex={tabIndex}
			aria-label={name}
			aria-orientation="horizontal"
			aria-valuemin={0}
			aria-valuemax={Math.round(baseSeconds)}
			aria-valuenow={Math.round(start)}
			aria-valuetext={valueTextWithState}
			data-shape={shape}
			data-dragging={preview ? '' : undefined}
			{...state}
			className={`absolute left-timeline-pos z-1 col-start-1 row-(--row) cursor-grab text-xs font-bold text-white outline-none select-none focus-visible:focus-ring data-dragging:z-2 data-dragging:cursor-grabbing data-selected:z-2 ${span ? SPAN_CLASSES : POINT_CLASSES}`}
			style={cssVars({
				// An absolutely placed grid item with an auto end line reaches the grid's edge, so the span is spelled out.
				'--row': `${row + 1} / span 1`,
				'--t': String(start),
				'--dur': String(fractionToSeconds(shown.end - shown.start, baseSeconds)),
			})}
			onKeyDown={onKeyDown}
			onClick={onSelect}
			onFocus={onFocus}
			{...drag}>
			{shape === 'span' ? (
				<>
					<Whisker seconds={startJitter} at="start" />
					<Whisker seconds={endJitter} at="end" />
					{invalid && <Icon name="triangle-exclamation" className="ml-1 shrink-0" />}
					<span className="min-w-0 truncate px-1">{name}</span>
					{/* At most a quarter of the bar each, so a short bar keeps a middle that moves it. */}
					<span aria-hidden="true" data-handle="start" className="absolute inset-y-0 left-0 w-1 max-w-1/4 cursor-ew-resize" />
					<span aria-hidden="true" data-handle="end" className="absolute inset-y-0 right-0 w-1 max-w-1/4 cursor-ew-resize" />
				</>
			) : (
				<>
					<span
						aria-hidden="true"
						{...state}
						className={`absolute inset-0 border border-transparent bg-(--bar) data-invalid:not-data-selected:border-danger data-selected:border-white ${POINT_SHAPE_CLASSES[shape]}`}
					/>
					<Whisker seconds={startJitter} at="center" />
				</>
			)}
		</div>
	);
};
