import { EncounterDesign } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { cssVars } from '@ui-kit/utils/css';
import { useEffect, useMemo, useRef, useState } from 'react';

import type { DesignEventKind, DesignEventRef, DesignLaneKind } from '../../../model/design/lanes';
import { eventLabelKey, eventsOf, laneLabelKey, removeEvent } from '../../../model/design/lanes';
import { fractionToSeconds } from '../../../model/design/time';
import { DesignerEventBar, type DesignerEventBarShape } from '../DesignerEventBar';
import { designerText, eventName, eventValueText } from '../utils/configs';
import { laneEventKinds, laneEvents, refKey, withWindow } from './utils';

export interface DesignerLaneProps {
	lane: DesignLaneKind;
	design: EncounterDesign;
	baseSeconds: number;
	snapSeconds: number;
	pixelsPerSecond: () => number;
	selected?: DesignEventRef;
	/** `refKey`s of the events that carry an error or a warning. */
	invalid: ReadonlySet<string>;
	onSelect: (ref: DesignEventRef | undefined) => void;
	/** Enter on a bar, after `onSelect`. */
	onActivate: (ref: DesignEventRef) => void;
	onChange: (design: EncounterDesign) => void;
	onAdd: (kind: DesignEventKind) => void;
}

const GHOST_CLASSES: Record<DesignerEventBarShape, string> = {
	span: 'top-1 bottom-1 w-timeline-segment min-w-2 rounded-sm',
	dot: 'top-1/2 size-(--designer-keyframe-size) -translate-1/2 rounded-full',
	diamond: 'top-1/2 size-(--designer-keyframe-size) -translate-1/2 rotate-45',
};

// The ref a later event in the same list has once `removed` is gone from it.
const afterRemoval = (ref: DesignEventRef, removed: DesignEventRef): DesignEventRef =>
	ref.kind === removed.kind && ref.index > removed.index ? { kind: ref.kind, index: ref.index - 1 } : ref;

export const DesignerLane = ({
	lane,
	design,
	baseSeconds,
	snapSeconds,
	pixelsPerSecond,
	selected,
	invalid,
	onSelect,
	onActivate,
	onChange,
	onAdd,
}: DesignerLaneProps) => {
	const events = useMemo(() => laneEvents(design, lane), [design, lane]);
	// The DOM keeps kind-then-index order whatever the times are: reordering keyed nodes moves the focused one, and the browser blurs it.
	const domOrder = useMemo(() => {
		const kinds = laneEventKinds(lane);
		return [...events].sort((a, b) => kinds.indexOf(a.ref.kind) - kinds.indexOf(b.ref.kind) || a.ref.index - b.ref.index);
	}, [events, lane]);
	const rowCount = Math.max(1, ...events.map(event => event.row + 1));
	const bars = useRef(new Map<string, HTMLDivElement>());
	const addButton = useRef<HTMLButtonElement>(null);
	// Where focus goes once a removal has rendered (a bar's key, or the add button for null), and the list length that shows it has.
	const focusAfterRemoval = useRef<{ key: string | null; from: EncounterDesign; kind: DesignEventKind; length: number } | null>(null);
	const [active, setActive] = useState<string>();

	const keys = events.map(event => refKey(event.ref));
	const selectedKey = selected && keys.includes(refKey(selected)) ? refKey(selected) : undefined;
	const tabbable = selectedKey ?? (active && keys.includes(active) ? active : keys[0]);
	const label = i18n.t(laneLabelKey(lane));

	// The first render after a removal settles it: focus moves only when that render shows the event gone.
	useEffect(() => {
		const target = focusAfterRemoval.current;
		if (!target) return;
		focusAfterRemoval.current = null;
		if (design === target.from || eventsOf(design, target.kind).length !== target.length - 1) return;
		if (target.key) bars.current.get(target.key)?.focus();
		else addButton.current?.focus();
	});

	const focusSibling = (key: string, direction: -1 | 1) => {
		const next = keys[keys.indexOf(key) + direction];
		if (!next) return;
		setActive(next);
		bars.current.get(next)?.focus();
	};

	const remove = (ref: DesignEventRef) => {
		const at = keys.indexOf(refKey(ref));
		const neighbour = events[at + 1]?.ref ?? events[at - 1]?.ref;
		focusAfterRemoval.current = {
			key: neighbour ? refKey(afterRemoval(neighbour, ref)) : null,
			from: design,
			kind: ref.kind,
			length: eventsOf(design, ref.kind).length,
		};
		const next = EncounterDesign.clone(design);
		removeEvent(next, ref);
		if (selected && refKey(selected) === refKey(ref)) onSelect(undefined);
		else if (selected && afterRemoval(selected, ref) !== selected) onSelect(afterRemoval(selected, ref));
		onChange(next);
	};

	return (
		<div
			role="group"
			aria-label={label}
			data-lane={lane}
			className="flex min-w-full border-b border-gray-700"
			style={cssVars({ '--bar': `var(--designer-lane-${lane}-bar)` })}>
			<div className="sticky left-0 z-3 box-border flex w-(--label-w) shrink-0 grow-0 basis-(--label-w) flex-col justify-center gap-1 border-r border-gray-700 bg-background px-2 py-1">
				<span className="text-sm font-bold text-white">{label}</span>
				<div className="flex flex-wrap gap-1">
					{laneEventKinds(lane).map((kind, index) => {
						const eventLabel = i18n.t(eventLabelKey(kind));
						return (
							<Button
								key={kind}
								ref={index === 0 ? addButton : undefined}
								variant="outline-secondary"
								size="xs"
								aria-label={designerText('add', { event: eventLabel })}
								onClick={() => onAdd(kind)}>
								<Icon name="plus" /> {eventLabel}
							</Button>
						);
					})}
				</div>
			</div>
			{/* One flat list placed on grid rows, so an edit that moves a bar to another row keeps the same element and its focus. */}
			<div className="relative grid shrink-0 grow basis-timeline-total auto-rows-(--designer-lane-height) grid-cols-1">
				{Array.from({ length: rowCount }, (_, row) => (
					<div key={row} aria-hidden="true" />
				))}
				{domOrder.flatMap(event =>
					event.repeats.map((repeat, index) => (
						<div
							key={`${refKey(event.ref)}:${index}`}
							aria-hidden="true"
							className={`pointer-events-none absolute left-timeline-pos col-start-1 row-(--row) bg-(--bar) opacity-40 ${GHOST_CLASSES[event.shape]}`}
							style={cssVars({
								'--row': `${event.row + 1} / span 1`,
								'--t': String(fractionToSeconds(repeat.start, baseSeconds)),
								'--dur': String(fractionToSeconds(repeat.end - repeat.start, baseSeconds)),
							})}
						/>
					)),
				)}
				{domOrder.map(event => {
					const key = refKey(event.ref);
					return (
						<DesignerEventBar
							key={key}
							ref={element => {
								if (element) bars.current.set(key, element);
								else bars.current.delete(key);
							}}
							name={eventName(design, event.ref)}
							window={event.window}
							row={event.row}
							shape={event.shape}
							openEnd={event.openEnd}
							jitter={event.jitter}
							baseSeconds={baseSeconds}
							snapSeconds={snapSeconds}
							pixelsPerSecond={pixelsPerSecond}
							valueText={window => eventValueText(design, event.ref, baseSeconds, window)}
							selected={key === selectedKey}
							invalid={invalid.has(key)}
							tabIndex={key === tabbable ? 0 : -1}
							onChange={window => onChange(withWindow(design, event.ref, window))}
							onSelect={() => onSelect(event.ref)}
							onActivate={() => onActivate(event.ref)}
							onRemove={() => remove(event.ref)}
							onSibling={direction => focusSibling(key, direction)}
							onFocus={() => setActive(key)}
						/>
					);
				})}
			</div>
		</div>
	);
};
