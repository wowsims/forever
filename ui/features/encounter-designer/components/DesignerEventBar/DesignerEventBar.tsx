import { Input } from '@ui-kit/FormControl';
import type { AxisDragMode } from '@ui-kit/hooks/usePointerAxisDrag';
import { usePointerAxisDrag } from '@ui-kit/hooks/usePointerAxisDrag';
import { Icon } from '@ui-kit/Icon';
import { cssVars } from '@ui-kit/utils/css';
import type { KeyboardEvent as ReactKeyboardEvent, Ref } from 'react';
import { useRef, useState } from 'react';

import type { TimeWindow } from '../../model/time';
import { fractionToSeconds, secondsToFraction, snap } from '../../model/time';
import { designerText } from '../EncounterDesigner/utils/configs';
import { barKeyAction, editWindow } from './utils';

export interface DesignerEventBarProps {
	name: string;
	/** The text on the bar when it is not the name. */
	label?: string;
	/** Marks the bar for the lane's context menu, which finds the event it opened on. */
	dataKey?: string;
	/** A stretch with nothing to do, drawn fainter: a boss health segment that holds. */
	flat?: boolean;
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
	/** Enter or blur with the text; Escape with undefined. `refocus` is for a key, which leaves focus on the field about to go. */
	onRename: (name: string | undefined, refocus: boolean) => void;
	/** Bounds an edit the way the commit will, so the bar previews what is stored: the window the edit gives, its mode and the window it started from. */
	adjust?: (next: TimeWindow, mode: AxisDragMode, from: TimeWindow) => TimeWindow;
	/** The window a drag is previewing, or null when it ends; neighbours that follow it read this. */
	onPreview?: (window: TimeWindow | null) => void;
	onSibling: (direction: -1 | 1) => void;
	onFocus?: () => void;
	ref?: Ref<HTMLDivElement>;
}

const MIN_WIDTH_SECONDS = 1;

const BAR_CLASSES =
	'top-1 bottom-1 flex w-timeline-segment min-w-2 items-center rounded-xs border border-transparent bg-(--bar) data-selected:border-white data-invalid:border-dashed data-invalid:not-data-selected:border-danger';

const Whisker = ({ seconds, at }: { seconds: number; at: 'start' | 'end' | 'center' }) =>
	seconds > 0 ? (
		<span
			aria-hidden="true"
			data-at={at}
			className="pointer-events-none absolute top-1/2 h-px w-timeline-segment -translate-x-1/2 bg-(--designer-jitter-whisker) data-[at=center]:left-1/2 data-[at=end]:left-full data-[at=start]:left-0"
			style={cssVars({ '--dur': String(2 * seconds) })}
		/>
	) : null;

// Enter or leaving the field keeps the text, Escape drops it. Its keys stay off the dialog, which Escape would close.
const RenameInput = ({
	name,
	row,
	start,
	duration,
	onRename,
}: {
	name: string;
	row: number;
	start: number;
	duration: number;
	onRename: DesignerEventBarProps['onRename'];
}) => {
	const done = useRef(false);
	const finish = (value: string | undefined, refocus: boolean) => {
		if (done.current) return;
		done.current = true;
		onRename(value, refocus);
	};
	return (
		<Input
			type="text"
			autoFocus
			defaultValue={name}
			aria-label={designerText('menu.rename_label')}
			className="absolute top-1 bottom-1 left-designer-pos z-3 col-start-1 row-(--row) w-timeline-segment min-w-32 px-1 text-xs"
			style={cssVars({ '--row': `${row + 1} / span 1`, '--t': String(start), '--dur': String(duration) })}
			onFocus={event => event.currentTarget.select()}
			onKeyDown={event => {
				event.stopPropagation();
				if (event.key === 'Enter') finish(event.currentTarget.value, true);
				else if (event.key === 'Escape') finish(undefined, true);
			}}
			onBlur={event => finish(event.currentTarget.value, false)}
		/>
	);
};

export const DesignerEventBar = ({
	name,
	label,
	dataKey,
	flat = false,
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
	adjust,
	onPreview,
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

	// The window an edit gives, bounded as the commit will bound it.
	const edit = (from: TimeWindow, mode: AxisDragMode, deltaFraction: number): TimeWindow => {
		if (startFixed && mode === 'start') return from;
		const next = editWindow(from, mode, deltaFraction, minWidth, openEnd);
		return adjust ? adjust(next, mode, from) : next;
	};
	const showPreview = (next: TimeWindow | null) => {
		setPreview(next);
		onPreview?.(next);
	};

	const drag = usePointerAxisDrag({
		pixelsPerSecond,
		onStart: () => {
			origin.current = window;
			cancelled.current = false;
			if (!duplicating.current) onSelect();
		},
		onDelta: (seconds, mode) => {
			if (cancelled.current) return;
			const delta = secondsToFraction(snap(seconds, snapSeconds), baseSeconds);
			// A copy moves whole, whatever the stored event does when its body moves.
			const next = duplicating.current ? editWindow(origin.current, 'move', delta, minWidth, openEnd) : edit(origin.current, mode, delta);
			latest.current = next;
			showPreview(next);
		},
		onEnd: commit => {
			const next = latest.current;
			latest.current = null;
			showPreview(null);
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
			showPreview(null);
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
				const next = edit(window, action.mode, secondsToFraction(step * action.steps, baseSeconds));
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

	const duration = fractionToSeconds(shown.end - shown.start, baseSeconds);

	return (
		<>
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
				data-flat={flat ? '' : undefined}
				data-tooltip-id={tooltipId}
				data-tip={tooltipText?.(shown)}
				data-dragging={preview ? '' : undefined}
				{...state}
				className={`absolute left-designer-pos z-1 col-start-1 row-(--row) cursor-grab text-xs font-bold text-white outline-none select-none focus-visible:focus-ring data-dragging:z-2 data-dragging:cursor-grabbing data-flat:opacity-(--designer-hold-opacity) data-selected:z-2 ${BAR_CLASSES}`}
				style={cssVars({
					// An absolutely placed grid item with an auto end line reaches the grid's edge, so the span is spelled out.
					'--row': `${row + 1} / span 1`,
					'--t': String(start),
					'--dur': String(duration),
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
				{!renaming && (
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
						className="pointer-events-none absolute top-0 left-timeline-pos h-full w-timeline-segment min-w-2 rounded-xs border border-white bg-(--bar) opacity-70"
						style={cssVars({
							'--t': String(fractionToSeconds(copy.start - window.start, baseSeconds)),
							'--dur': String(fractionToSeconds(copy.end - copy.start, baseSeconds)),
						})}
					/>
				)}
			</div>
			{/* Outside the slider, whose children are presentational: a text box assistive technology can reach. */}
			{renaming && <RenameInput name={name} row={row} start={start} duration={duration} onRename={onRename} />}
		</>
	);
};
