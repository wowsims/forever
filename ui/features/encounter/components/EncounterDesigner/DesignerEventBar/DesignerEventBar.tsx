import { Input } from '@ui-kit/FormControl';
import { usePointerAxisDrag } from '@ui-kit/hooks/usePointerAxisDrag';
import { Icon } from '@ui-kit/Icon';
import { cssVars } from '@ui-kit/utils/css';
import type { KeyboardEvent as ReactKeyboardEvent, Ref } from 'react';
import { useRef, useState } from 'react';

import type { TimeWindow } from '../../../model/design/time';
import { fractionToSeconds, secondsToFraction, snap } from '../../../model/design/time';
import { designerText } from '../utils/configs';
import { barKeyAction, editWindow } from './utils';

export interface DesignerEventBarProps {
	name: string;
	/** The text on the bar when it is not the name. */
	label?: string;
	/** Marks the bar for the lane's context menu, which finds the event it opened on. */
	dataKey?: string;
	/** The label is a text field. */
	renaming?: boolean;
	/** A point event has `start === end`. Fractions of the fight. */
	window: TimeWindow;
	/** The lane row, counted from 0, when the bar sits in a lane's grid. */
	row: number;
	/** The start edge is fixed, as the boss health curve's implied first point is; it gets no start handle. */
	startFixed?: boolean;
	/** The end may be dragged past the fight's end, and a move stops only the start there. */
	openEnd?: boolean;
	/** The jitter around the start and the end, as fractions of the fight. */
	jitter?: { start: number; end: number };
	baseSeconds: number;
	/** The drag and keyboard step; 0 drags freely and steps by 1 s. */
	snapSeconds: number;
	pixelsPerSecond: () => number;
	valueText: (window: TimeWindow) => string;
	/** What the event does, for the hover tooltip that `tooltipId` names. */
	tooltipText?: (window: TimeWindow) => string;
	tooltipId?: string;
	selected?: boolean;
	invalid?: boolean;
	tabIndex: 0 | -1;
	onChange: (window: TimeWindow) => void;
	onSelect: () => void;
	/** Enter, after `onSelect`; a click or a drag only selects. */
	onActivate: () => void;
	onRemove: () => void;
	/** Ctrl+D, or an Alt drag that ended: a copy of the event shifted by the fraction of the fight. */
	onDuplicate: (deltaFraction: number) => void;
	/** Double click on the label, or F2. */
	onRenameStart: () => void;
	/** Enter or blur with the text; Escape with undefined. */
	onRename: (name: string | undefined) => void;
	onSibling: (direction: -1 | 1) => void;
	onFocus?: () => void;
	ref?: Ref<HTMLDivElement>;
}

const MIN_WIDTH_SECONDS = 1;

const BAR_CLASSES =
	'top-1 bottom-1 flex w-timeline-segment min-w-2 items-center rounded-sm border border-transparent bg-(--bar) data-selected:border-white data-invalid:border-dashed data-invalid:not-data-selected:border-danger';

const Whisker = ({ seconds, at }: { seconds: number; at: 'start' | 'end' | 'center' }) =>
	seconds > 0 ? (
		<span
			aria-hidden="true"
			data-at={at}
			className="pointer-events-none absolute top-1/2 h-px w-timeline-segment -translate-x-1/2 bg-(--designer-jitter-whisker) data-[at=center]:left-1/2 data-[at=end]:left-full data-[at=start]:left-0"
			style={cssVars({ '--dur': String(2 * seconds) })}
		/>
	) : null;

// Enter or leaving the field keeps the text, Escape drops it. Its keys and pointer stay off the bar, which would move or remove it.
const RenameInput = ({ name, onRename }: { name: string; onRename: (name: string | undefined) => void }) => {
	const done = useRef(false);
	const finish = (value: string | undefined) => {
		if (done.current) return;
		done.current = true;
		onRename(value);
	};
	return (
		<Input
			type="text"
			autoFocus
			defaultValue={name}
			aria-label={designerText('menu.rename_label')}
			className="min-w-0 flex-1 px-1 text-xs"
			onFocus={event => event.currentTarget.select()}
			onPointerDown={event => event.stopPropagation()}
			onClick={event => event.stopPropagation()}
			onKeyDown={event => {
				event.stopPropagation();
				if (event.key === 'Enter') finish(event.currentTarget.value);
				else if (event.key === 'Escape') finish(undefined);
			}}
			onBlur={event => finish(event.currentTarget.value)}
		/>
	);
};

export const DesignerEventBar = ({
	name,
	label,
	dataKey,
	renaming = false,
	window,
	row,
	startFixed = false,
	openEnd = false,
	jitter,
	baseSeconds,
	snapSeconds,
	pixelsPerSecond,
	valueText,
	tooltipText,
	tooltipId,
	selected,
	invalid,
	tabIndex,
	onChange,
	onSelect,
	onActivate,
	onRemove,
	onDuplicate,
	onRenameStart,
	onRename,
	onSibling,
	onFocus,
	ref,
}: DesignerEventBarProps) => {
	const [preview, setPreview] = useState<TimeWindow | null>(null);
	const origin = useRef(window);
	const latest = useRef<TimeWindow | null>(null);
	const minWidth = secondsToFraction(MIN_WIDTH_SECONDS, baseSeconds);
	// An Alt drag leaves the event where it is and drags a copy; Escape drops the drag.
	const duplicating = useRef(false);
	const cancelled = useRef(false);

	const drag = usePointerAxisDrag({
		pixelsPerSecond,
		onStart: () => {
			origin.current = window;
			cancelled.current = false;
			if (!duplicating.current) onSelect();
		},
		onDelta: (seconds, mode) => {
			if (cancelled.current) return;
			const next = editWindow(
				origin.current,
				duplicating.current ? 'move' : mode,
				secondsToFraction(snap(seconds, snapSeconds), baseSeconds),
				minWidth,
				openEnd,
			);
			latest.current = next;
			setPreview(next);
		},
		onEnd: commit => {
			const next = latest.current;
			latest.current = null;
			setPreview(null);
			const wasDuplicating = duplicating.current;
			duplicating.current = false;
			if (!commit || cancelled.current || !next) return;
			if (wasDuplicating) onDuplicate(next.start - window.start);
			else if (next.start !== window.start || next.end !== window.end) onChange(next);
		},
	});

	const onKeyDown = (event: ReactKeyboardEvent<HTMLDivElement>) => {
		if (event.key === 'Escape' && preview) {
			event.stopPropagation();
			cancelled.current = true;
			latest.current = null;
			setPreview(null);
			return;
		}
		if (event.key === 'F2') {
			event.preventDefault();
			event.stopPropagation();
			onRenameStart();
			return;
		}
		const action = barKeyAction(event);
		if (!action) return;
		event.preventDefault();
		event.stopPropagation();
		switch (action.type) {
			case 'edit': {
				const step = snapSeconds > 0 ? snapSeconds : 1;
				const next = editWindow(window, action.mode, secondsToFraction(step * action.steps, baseSeconds), minWidth, openEnd);
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
			case 'duplicate':
				onDuplicate(secondsToFraction(snapSeconds > 0 ? snapSeconds : 1, baseSeconds));
				return;
			case 'sibling':
				onSibling(action.direction);
		}
	};

	// A copy being dragged shows beside the event, which keeps its own place.
	const copy = preview && duplicating.current ? preview : null;
	const shown = preview && !copy ? preview : window;
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
			data-key={dataKey}
			data-tooltip-id={tooltipId}
			data-tip={tooltipText?.(shown)}
			data-dragging={preview ? '' : undefined}
			{...state}
			className={`absolute left-designer-pos z-1 col-start-1 row-(--row) cursor-grab text-xs font-bold text-white outline-none select-none focus-visible:focus-ring data-dragging:z-2 data-dragging:cursor-grabbing data-selected:z-2 ${BAR_CLASSES}`}
			style={cssVars({
				// An absolutely placed grid item with an auto end line reaches the grid's edge, so the span is spelled out.
				'--row': `${row + 1} / span 1`,
				'--t': String(start),
				'--dur': String(fractionToSeconds(shown.end - shown.start, baseSeconds)),
			})}
			onKeyDown={onKeyDown}
			onClick={onSelect}
			onFocus={onFocus}
			{...drag}
			onPointerDown={event => {
				duplicating.current = event.altKey;
				drag.onPointerDown(event);
			}}>
			<Whisker seconds={startJitter} at="start" />
			<Whisker seconds={endJitter} at="end" />
			{invalid && <Icon name="triangle-exclamation" className="ml-1 shrink-0" />}
			{renaming ? (
				<RenameInput name={name} onRename={onRename} />
			) : (
				<span className="min-w-0 truncate px-1" onDoubleClick={onRenameStart}>
					{label ?? name}
				</span>
			)}
			{/* At most a quarter of the bar each, so a short bar keeps a middle that moves it. */}
			{!startFixed && <span aria-hidden="true" data-handle="start" className="absolute inset-y-0 left-0 w-1 max-w-1/4 cursor-ew-resize" />}
			<span aria-hidden="true" data-handle="end" className="absolute inset-y-0 right-0 w-1 max-w-1/4 cursor-ew-resize" />
			{copy && (
				<span
					aria-hidden="true"
					data-copy=""
					className="pointer-events-none absolute top-0 left-timeline-pos h-full w-timeline-segment min-w-2 rounded-sm border border-white bg-(--bar) opacity-70"
					style={cssVars({
						'--t': String(fractionToSeconds(copy.start - window.start, baseSeconds)),
						'--dur': String(fractionToSeconds(copy.end - copy.start, baseSeconds)),
					})}
				/>
			)}
		</div>
	);
};
