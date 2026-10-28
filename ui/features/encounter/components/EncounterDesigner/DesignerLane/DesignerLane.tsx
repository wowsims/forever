import { EncounterDesign } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ContextMenu } from '@ui-kit/ContextMenu';
import type { AxisDragMode } from '@ui-kit/hooks/usePointerAxisDrag';
import { Icon } from '@ui-kit/Icon';
import { MenuItem } from '@ui-kit/Menu';
import { cssVars } from '@ui-kit/utils/css';
import type { MouseEvent as ReactMouseEvent, PointerEvent as ReactPointerEvent, TouchEvent as ReactTouchEvent } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';

import type { DesignEventKind, DesignEventRef, DesignLaneKind } from '../../../model/design/lanes';
import {
	duplicateEvent,
	eventLabelKey,
	eventsOf,
	getEvent,
	healthEndSegment,
	healthSegments,
	holdBlocker,
	holdDuringDowntime,
	holdSegment,
	laneLabelKey,
	removeEvent,
} from '../../../model/design/lanes';
import type { TimeWindow } from '../../../model/design/time';
import { fractionToSeconds, secondsToFraction } from '../../../model/design/time';
import { DesignerEventBar } from '../DesignerEventBar';
import {
	designerText,
	editEvent,
	eventName,
	eventTooltip,
	eventValueText,
	formatPercent,
	formatSeconds,
	occurrenceName,
	occurrenceTip,
	serializeTip,
} from '../utils/configs';
import { type LaneEvent, laneEventKinds, laneEvents, refKey, withWindow } from './utils';

export interface DesignerLaneProps {
	lane: DesignLaneKind;
	design: EncounterDesign;
	baseSeconds: number;
	snapSeconds: number;
	targetNames: ReadonlyArray<string>;
	/** The Tooltip every bar, ghost and tick anchors to. */
	tooltipId: string;
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

const GHOST_CLASSES = 'top-1 bottom-1 flex w-timeline-segment min-w-2 items-center overflow-hidden rounded-sm text-xs font-bold text-white';

// The ref a later event in the same list has once `removed` is gone from it.
const afterRemoval = (ref: DesignEventRef, removed: DesignEventRef): DesignEventRef =>
	ref.kind === removed.kind && ref.index > removed.index ? { kind: ref.kind, index: ref.index - 1 } : ref;

export const DesignerLane = ({
	lane,
	design,
	baseSeconds,
	snapSeconds,
	targetNames,
	tooltipId,
	pixelsPerSecond,
	selected,
	invalid,
	onSelect,
	onActivate,
	onChange,
	onAdd,
}: DesignerLaneProps) => {
	const events = useMemo(() => laneEvents(design, lane, baseSeconds), [design, lane, baseSeconds]);
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
	const [menuKey, setMenuKey] = useState<string>();
	const [menuOpen, setMenuOpen] = useState(false);
	const menuKeyAt = useRef<string | undefined>(undefined);
	// An item's action puts focus where it wants it; a dismissal returns it to the bar the menu opened on.
	const menuActed = useRef(false);
	const renameAfterMenu = useRef<string | undefined>(undefined);
	// The boss health segment being dragged, whose neighbours' shared edge follows it.
	const [dragging, setDragging] = useState<{ index: number; window: TimeWindow } | null>(null);
	const [renaming, setRenaming] = useState<string>();
	// The copy a duplicate added, focused once it renders.
	const focusAfterAdd = useRef<string | null>(null);

	const keys = events.map(event => refKey(event.ref));
	const selectedKey = selected && keys.includes(refKey(selected)) ? refKey(selected) : undefined;
	const tabbable = selectedKey ?? (active && keys.includes(active) ? active : keys[0]);
	const label = i18n.t(laneLabelKey(lane));

	// The first render after a removal settles it: focus moves only when that render shows the event gone.
	useEffect(() => {
		if (focusAfterAdd.current && bars.current.has(focusAfterAdd.current)) {
			bars.current.get(focusAfterAdd.current)?.focus();
			focusAfterAdd.current = null;
		}
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

	const duplicate = (ref: DesignEventRef, deltaFraction: number) => {
		const next = EncounterDesign.clone(design);
		const copy = duplicateEvent(next, ref, deltaFraction);
		focusAfterAdd.current = refKey(copy);
		onChange(next);
		onSelect(copy);
	};

	// A key ends the rename with focus on the field about to go, so the bar takes it; a click elsewhere has moved focus already.
	const rename = (ref: DesignEventRef, name: string | undefined, refocus: boolean) => {
		setRenaming(undefined);
		if (name !== undefined && name !== eventName(design, ref)) onChange(editEvent(design, ref, event => (event.name = name)));
		if (refocus) bars.current.get(refKey(ref))?.focus();
	};

	// The menu opens only on a bar, also for a touch long press, which Base UI opens without a contextmenu event.
	const barKeyAt = (target: EventTarget) => (target as Element).closest<HTMLElement>('[data-key]')?.dataset.key;
	const acted = (action: () => void) => () => {
		menuActed.current = true;
		action();
	};

	// Boss health held by hand: a segment's start cut off as a flat part, or the curve's value kept across a downtime of the boss.
	const hold = (ref: DesignEventRef) => {
		const next = EncounterDesign.clone(design);
		const copy = holdSegment(next, ref, baseSeconds);
		if (!copy) return;
		focusAfterAdd.current = refKey(copy);
		onChange(next);
		onSelect(copy);
	};
	const holdBoss = (ref: DesignEventRef) => {
		const next = EncounterDesign.clone(design);
		const [first] = holdDuringDowntime(next, ref.index);
		if (!first) return;
		onChange(next);
		onSelect(first);
	};

	const menuEvent = events.find(event => refKey(event.ref) === menuKey);
	const blocker = menuEvent?.ref.kind === 'downtime' ? holdBlocker(design, menuEvent.ref.index) : undefined;
	const holdBlocked = blocker
		? designerText('menu.hold_blocked', {
				name: eventName(design, blocker),
				at: formatSeconds(fractionToSeconds(design.bossHealth[blocker.index].at, baseSeconds)),
			})
		: undefined;
	const snapFraction = secondsToFraction(snapSeconds > 0 ? snapSeconds : 1, baseSeconds);
	const segments = healthSegments(design);
	const healthLabel = (ref: DesignEventRef) => {
		const segment = segments.find(found => found.index === ref.index);
		const [from, to] = [segment?.fromPercent ?? 100, segment?.toPercent ?? 100];
		const curve =
			from === to
				? designerText('health_hold', { percent: formatPercent(from) })
				: designerText('health_label', { from: formatPercent(from), to: formatPercent(to) });
		const named = getEvent(design, ref)?.name;
		return named ? `${named}: ${curve}` : curve;
	};
	// A segment dragged by its end, or its start handle, moves the edge it shares with its neighbour, which follows it.
	const followed = (event: LaneEvent): TimeWindow => {
		if (event.ref.kind !== 'bossHealth' || !dragging) return event.window;
		const dragged = segments.find(found => found.index === dragging.index);
		const mine = segments.find(found => found.index === event.ref.index);
		if (!dragged || !mine || dragged === mine) return event.window;
		if (mine.previous === dragged.index) return { start: dragging.window.end, end: event.window.end };
		if (dragged.previous === mine.index) return { start: event.window.start, end: dragging.window.start };
		return event.window;
	};
	// The window a boss health edit stores: the keyframe at the segment's end between its neighbours, or the one before the segment.
	const adjustHealth =
		(ref: DesignEventRef) =>
		(next: TimeWindow, mode: AxisDragMode, from: TimeWindow): TimeWindow => {
			const position = segments.findIndex(found => found.index === ref.index);
			const segment = segments[position];
			if (!segment) return from;
			if (mode === 'start') {
				if (segment.previous === undefined) return from;
				return { start: Math.min(Math.max(next.start, segments[position - 1]?.start ?? 0), segment.end), end: from.end };
			}
			return { start: from.start, end: Math.min(Math.max(next.end, segment.start), segments[position + 1]?.end ?? 1) };
		};
	const isHold = (ref: DesignEventRef) => {
		const segment = ref.kind === 'bossHealth' ? segments.find(found => found.index === ref.index) : undefined;
		return !!segment && segment.fromPercent === segment.toPercent;
	};
	const endSegment = lane === 'boss' ? healthEndSegment(design) : undefined;
	const endTip = endSegment
		? {
				from: formatPercent(endSegment.fromPercent),
				to: 0,
				start: formatSeconds(fractionToSeconds(endSegment.start, baseSeconds)),
				end: formatSeconds(fractionToSeconds(endSegment.end, baseSeconds)),
			}
		: undefined;
	const tip = (ref: DesignEventRef, window?: TimeWindow) => serializeTip(eventTooltip(design, ref, baseSeconds, targetNames, window));

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
			<ContextMenu
				className="relative grid shrink-0 grow basis-designer-total auto-rows-(--designer-lane-height) grid-cols-1"
				open={menuOpen}
				onOpenChange={next => {
					if (next && !menuKeyAt.current) return;
					if (next) {
						menuActed.current = false;
						setMenuKey(menuKeyAt.current);
					}
					setMenuOpen(next);
				}}
				// The field takes focus once the menu has let go of it, or the menu's own focus handling blurs the field and ends the rename.
				onOpenChangeComplete={next => {
					if (next || !renameAfterMenu.current) return;
					setRenaming(renameAfterMenu.current);
					renameAfterMenu.current = undefined;
				}}
				finalFocus={() => (menuActed.current ? false : (bars.current.get(menuKey ?? '') ?? true))}
				triggerProps={{
					onPointerDownCapture: (event: ReactPointerEvent<HTMLElement>) => {
						menuKeyAt.current = barKeyAt(event.target);
					},
					onTouchStartCapture: (event: ReactTouchEvent<HTMLElement>) => {
						menuKeyAt.current = barKeyAt(event.target);
					},
					// A right click on empty lane keeps the browser's own menu.
					onContextMenuCapture: (event: ReactMouseEvent<HTMLElement>) => {
						menuKeyAt.current = barKeyAt(event.target);
						if (!menuKeyAt.current) event.stopPropagation();
					},
				}}
				menu={
					menuEvent && (
						<>
							<MenuItem layout="block" onClick={acted(() => duplicate(menuEvent.ref, snapFraction))}>
								{designerText('menu.duplicate')}
							</MenuItem>
							<MenuItem layout="block" onClick={acted(() => (renameAfterMenu.current = refKey(menuEvent.ref)))}>
								{designerText('menu.rename')}
							</MenuItem>
							<MenuItem
								layout="block"
								onClick={acted(() => {
									onSelect(menuEvent.ref);
									onActivate(menuEvent.ref);
								})}>
								{designerText('menu.edit')}
							</MenuItem>
							{menuEvent.ref.kind === 'bossHealth' && (
								<MenuItem layout="block" onClick={acted(() => hold(menuEvent.ref))}>
									{designerText('menu.hold_health')}
								</MenuItem>
							)}
							{menuEvent.ref.kind === 'downtime' && design.downtimes[menuEvent.ref.index]?.targetIndex === 0 && (
								<MenuItem
									layout="block"
									disabled={!!holdBlocked}
									data-tooltip-id={holdBlocked ? tooltipId : undefined}
									data-tip={holdBlocked}
									onClick={acted(() => holdBoss(menuEvent.ref))}>
									{designerText('menu.hold_boss_health')}
								</MenuItem>
							)}
							<MenuItem layout="block" onClick={acted(() => remove(menuEvent.ref))}>
								{designerText('menu.delete')}
							</MenuItem>
						</>
					)
				}>
				{Array.from({ length: rowCount }, (_, row) => (
					<div key={row} aria-hidden="true" />
				))}
				{domOrder.flatMap(event => {
					const name = eventName(design, event.ref);
					const rowStyle = (window: TimeWindow) => ({
						'--row': `${event.row + 1} / span 1`,
						'--t': String(fractionToSeconds(window.start, baseSeconds)),
						'--dur': String(fractionToSeconds(window.end - window.start, baseSeconds)),
					});
					const ghosts = event.repeats.map((repeat, index) => {
						const label = occurrenceName(name, index + 2);
						return (
							<div
								key={`${refKey(event.ref)}:${index}`}
								aria-hidden="true"
								data-tooltip-id={tooltipId}
								data-tip={serializeTip(occurrenceTip(eventTooltip(design, event.ref, baseSeconds, targetNames, repeat), label))}
								className={`absolute left-designer-pos col-start-1 row-(--row) bg-(--bar)/40 ${GHOST_CLASSES}`}
								style={cssVars(rowStyle(repeat))}>
								<span className="min-w-0 truncate pl-1">{name}</span>
								<span className="shrink-0 px-1">{designerText('occurrence_number', { number: index + 2 })}</span>
							</div>
						);
					});
					// The ticks sit over their bar and take no pointer, so the bar's own grab and tooltip reach through them.
					const ticks = event.ticks.map((at, index) => (
						<div
							key={`${refKey(event.ref)}:tick:${index}`}
							aria-hidden="true"
							className="pointer-events-none absolute top-1 bottom-1 left-designer-pos z-2 col-start-1 row-(--row) w-px bg-white"
							style={cssVars({ '--row': `${event.row + 1} / span 1`, '--t': String(fractionToSeconds(at, baseSeconds)) })}
						/>
					));
					return [...ghosts, ...ticks];
				})}
				{endSegment && endTip && (
					<div
						role="img"
						aria-label={designerText('tip.health_sentence', endTip)}
						data-tooltip-id={tooltipId}
						data-tip={serializeTip({
							title: designerText('tip.boss_health'),
							lines: [designerText('tip.health', endTip), designerText('tip.range', { start: endTip.start, end: endTip.end })],
						})}
						className={`absolute left-designer-pos col-start-1 row-(--row) border border-dashed border-white/50 bg-(--bar)/40 ${GHOST_CLASSES}`}
						style={cssVars({
							'--row': '1 / span 1',
							'--t': String(fractionToSeconds(endSegment.start, baseSeconds)),
							'--dur': String(fractionToSeconds(endSegment.end - endSegment.start, baseSeconds)),
						})}>
						<span className="min-w-0 truncate px-1">{designerText('health_end')}</span>
					</div>
				)}
				{domOrder.map(event => {
					const key = refKey(event.ref);
					return (
						<DesignerEventBar
							key={key}
							dataKey={key}
							ref={element => {
								if (element) bars.current.set(key, element);
								else bars.current.delete(key);
							}}
							name={eventName(design, event.ref)}
							label={event.ref.kind === 'bossHealth' ? healthLabel(event.ref) : undefined}
							flat={isHold(event.ref)}
							window={followed(event)}
							row={event.row}
							startFixed={event.startFixed}
							openEnd={event.openEnd}
							jitter={event.jitter}
							baseSeconds={baseSeconds}
							snapSeconds={snapSeconds}
							pixelsPerSecond={pixelsPerSecond}
							valueText={window => eventValueText(design, event.ref, baseSeconds, window)}
							tooltipId={tooltipId}
							tooltipText={window => tip(event.ref, window)}
							selected={key === selectedKey}
							invalid={invalid.has(key)}
							tabIndex={key === tabbable ? 0 : -1}
							renaming={key === renaming}
							onChange={window => onChange(withWindow(design, event.ref, window, baseSeconds))}
							onSelect={() => onSelect(event.ref)}
							onActivate={() => onActivate(event.ref)}
							onRemove={() => remove(event.ref)}
							onDuplicate={delta => duplicate(event.ref, delta)}
							onRenameStart={() => setRenaming(key)}
							onRename={(name, refocus) => rename(event.ref, name, refocus)}
							adjust={event.ref.kind === 'bossHealth' ? adjustHealth(event.ref) : undefined}
							onPreview={event.ref.kind === 'bossHealth' ? window => setDragging(window && { index: event.ref.index, window }) : undefined}
							onSibling={direction => focusSibling(key, direction)}
							onFocus={() => setActive(key)}
						/>
					);
				})}
			</ContextMenu>
		</div>
	);
};
