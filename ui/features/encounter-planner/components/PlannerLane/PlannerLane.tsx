import { EncounterPlan } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { ContextMenu } from '@ui-kit/ContextMenu';
import { Icon } from '@ui-kit/Icon';
import { MenuItem } from '@ui-kit/Menu';
import { cssVars } from '@ui-kit/utils/css';
import type { MouseEvent as ReactMouseEvent, PointerEvent as ReactPointerEvent, TouchEvent as ReactTouchEvent } from 'react';
import { useEffect, useMemo, useRef, useState } from 'react';

import type { PlanEventKind, PlanEventRef, PlanLaneKind } from '../../model/lanes';
import {
	duplicateEvent,
	eventLabelKey,
	eventsOf,
	healthEndSegment,
	healthSegments,
	holdBlocker,
	holdDuringDowntime,
	holdSegment,
	LANE_EVENT_KINDS,
	laneLabelKey,
	removeEvent,
} from '../../model/lanes';
import type { TimeWindow } from '../../model/time';
import { fractionToSeconds, keyStep, secondsToFraction } from '../../model/time';
import { editEvent, plannerText } from '../EncounterPlanner/utils/configs';
import {
	eventName,
	eventTooltip,
	eventValueText,
	formatPercent,
	formatSeconds,
	occurrenceName,
	occurrenceTip,
	serializeTip,
} from '../EncounterPlanner/utils/tips';
import { PlannerEventBar } from '../PlannerEventBar';
import { barVars, tickVars } from '../PlannerEventBar/utils';
import {
	adjustHealth,
	followed,
	healthLabel,
	isHold,
	LANE_CLASSES,
	LANE_GRID_CLASSES,
	LANE_LABEL_CLASSES,
	laneEvents,
	refKey,
	rowKinds,
	withWindow,
} from './utils';

export interface PlannerLaneProps {
	lane: PlanLaneKind;
	plan: EncounterPlan;
	baseSeconds: number;
	snapSeconds: number;
	targetNames: ReadonlyArray<string>;
	/** The Tooltip every bar, ghost and tick anchors to. */
	tooltipId: string;
	pixelsPerSecond: () => number;
	selected?: PlanEventRef;
	/** `refKey`s of the events that carry an error or a warning. */
	invalid: ReadonlySet<string>;
	onSelect: (ref: PlanEventRef | undefined) => void;
	/** Enter on a bar, after `onSelect`. */
	onActivate: (ref: PlanEventRef) => void;
	onChange: (plan: EncounterPlan) => void;
	onAdd: (kind: PlanEventKind, targetIndex?: number) => void;
	/** Only for the 'boss' lane, where it makes the lane the row of one encounter target, under a label of its name that selects it. */
	target?: PlannerLaneTarget;
}

interface PlannerLaneTarget {
	index: number;
	selected: boolean;
	onSelect: () => void;
	/** Enter on the label, after `onSelect`; a click only selects. */
	onActivate: () => void;
}

const GHOST_CLASSES = 'top-1 bottom-1 flex w-timeline-segment min-w-2 items-center overflow-hidden rounded-xs text-xs font-bold text-white';

// The ref a later event in the same list has once `removed` is gone from it.
const afterRemoval = (ref: PlanEventRef, removed: PlanEventRef): PlanEventRef =>
	ref.kind === removed.kind && ref.index > removed.index ? { kind: ref.kind, index: ref.index - 1 } : ref;

export const PlannerLane = ({
	lane,
	plan,
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
	target,
}: PlannerLaneProps) => {
	const targetIndex = target?.index;
	const targetCount = targetNames.length;
	const events = useMemo(
		() => laneEvents(plan, lane, baseSeconds, targetIndex === undefined ? undefined : { index: targetIndex, count: targetCount }),
		[plan, lane, baseSeconds, targetIndex, targetCount],
	);
	const kinds = rowKinds(lane, targetIndex);
	// The DOM keeps kind-then-index order whatever the times are: reordering keyed nodes moves the focused one, and the browser blurs it.
	const domOrder = useMemo(() => {
		const order = LANE_EVENT_KINDS[lane];
		return [...events].sort((a, b) => order.indexOf(a.ref.kind) - order.indexOf(b.ref.kind) || a.ref.index - b.ref.index);
	}, [events, lane]);
	const rowCount = Math.max(1, ...events.map(event => event.row + 1));
	const bars = useRef(new Map<string, HTMLDivElement>());
	const addButton = useRef<HTMLButtonElement>(null);
	// Where focus goes once a removal has rendered (a bar's key, or the add button for null), and the list length that shows it has.
	const focusAfterRemoval = useRef<{ key: string | null; from: EncounterPlan; kind: PlanEventKind; length: number } | null>(null);
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
	const label = target ? targetNames[target.index] : i18n.t(laneLabelKey(lane));

	// The first render after a removal settles it: focus moves only when that render shows the event gone.
	useEffect(() => {
		if (focusAfterAdd.current && bars.current.has(focusAfterAdd.current)) {
			bars.current.get(focusAfterAdd.current)?.focus();
			focusAfterAdd.current = null;
		}
		const target = focusAfterRemoval.current;
		if (!target) return;
		focusAfterRemoval.current = null;
		if (plan === target.from || eventsOf(plan, target.kind).length !== target.length - 1) return;
		if (target.key) bars.current.get(target.key)?.focus();
		else addButton.current?.focus();
	});

	const focusSibling = (key: string, direction: -1 | 1) => {
		const next = keys[keys.indexOf(key) + direction];
		if (!next) return;
		setActive(next);
		bars.current.get(next)?.focus();
	};

	const remove = (ref: PlanEventRef) => {
		const at = keys.indexOf(refKey(ref));
		const neighbour = events[at + 1]?.ref ?? events[at - 1]?.ref;
		focusAfterRemoval.current = {
			key: neighbour ? refKey(afterRemoval(neighbour, ref)) : null,
			from: plan,
			kind: ref.kind,
			length: eventsOf(plan, ref.kind).length,
		};
		const next = EncounterPlan.clone(plan);
		removeEvent(next, ref);
		if (selected && refKey(selected) === refKey(ref)) onSelect(undefined);
		else if (selected && afterRemoval(selected, ref) !== selected) onSelect(afterRemoval(selected, ref));
		onChange(next);
	};

	const added = (next: EncounterPlan, ref: PlanEventRef) => {
		focusAfterAdd.current = refKey(ref);
		onChange(next);
		onSelect(ref);
	};

	const duplicate = (ref: PlanEventRef, deltaFraction: number) => {
		const next = EncounterPlan.clone(plan);
		added(next, duplicateEvent(next, ref, deltaFraction));
	};

	// A key ends the rename with focus on the field about to go, so the bar takes it; a click elsewhere has moved focus already.
	const rename = (ref: PlanEventRef, name: string | undefined, refocus: boolean) => {
		setRenaming(undefined);
		if (name !== undefined && name !== eventName(plan, ref)) onChange(editEvent(plan, ref, event => (event.name = name)));
		if (refocus) bars.current.get(refKey(ref))?.focus();
	};

	// The menu opens only on a bar, also for a touch long press, which Base UI opens without a contextmenu event.
	const barKeyAt = (target: EventTarget) => (target as Element).closest<HTMLElement>('[data-key]')?.dataset.key;
	const noteMenuKey = (event: ReactPointerEvent<HTMLElement> | ReactTouchEvent<HTMLElement>) => {
		menuKeyAt.current = barKeyAt(event.target);
	};
	const acted = (action: () => void) => () => {
		menuActed.current = true;
		action();
	};

	// Boss health held by hand: a segment's start cut off as a flat part, or the curve's value kept across a downtime of the boss.
	const hold = (ref: PlanEventRef) => {
		const next = EncounterPlan.clone(plan);
		const copy = holdSegment(next, ref, baseSeconds);
		if (copy) added(next, copy);
	};
	const holdBoss = (ref: PlanEventRef) => {
		const next = EncounterPlan.clone(plan);
		const [first] = holdDuringDowntime(next, ref.index);
		if (!first) return;
		onChange(next);
		onSelect(first);
	};

	const menuEvent = events.find(event => refKey(event.ref) === menuKey);
	const blocker = menuEvent?.ref.kind === 'downtime' ? holdBlocker(plan, menuEvent.ref.index) : undefined;
	const holdBlocked = blocker
		? plannerText('menu.hold_blocked', {
				name: eventName(plan, blocker),
				at: formatSeconds(fractionToSeconds(plan.bossHealth[blocker.index].at, baseSeconds)),
			})
		: undefined;
	const snapFraction = secondsToFraction(keyStep(snapSeconds), baseSeconds);
	const segments = healthSegments(plan);
	const endSegment = kinds.includes('bossHealth') ? healthEndSegment(plan) : undefined;
	const endTip = endSegment
		? {
				from: formatPercent(endSegment.fromPercent),
				to: 0,
				start: formatSeconds(fractionToSeconds(endSegment.start, baseSeconds)),
				end: formatSeconds(fractionToSeconds(endSegment.end, baseSeconds)),
			}
		: undefined;
	const tip = (ref: PlanEventRef, window?: TimeWindow) => serializeTip(eventTooltip(plan, ref, baseSeconds, targetNames, window));

	return (
		<div role="group" aria-label={label} className={LANE_CLASSES} style={cssVars({ '--bar': `var(--planner-lane-${lane}-bar)` })}>
			<div className={LANE_LABEL_CLASSES}>
				{target ? (
					<Button
						variant="outline-secondary"
						size="xs"
						aria-pressed={target.selected}
						data-selected={target.selected ? '' : undefined}
						className="max-w-full self-start data-selected:border-white data-selected:bg-(--bar) data-selected:text-white"
						onClick={target.onSelect}
						onKeyDown={event => {
							if (event.key !== 'Enter') return;
							event.preventDefault();
							target.onSelect();
							target.onActivate();
						}}>
						<span className="min-w-0 truncate">{label}</span>
						<Icon name="gear" />
					</Button>
				) : (
					<span className="text-sm font-bold text-white">{label}</span>
				)}
				<div data-lane-add="" className="flex flex-wrap gap-1">
					{kinds.map((kind, index) => {
						const eventLabel = i18n.t(eventLabelKey(kind));
						return (
							<Button
								key={kind}
								ref={index === 0 ? addButton : undefined}
								variant="outline-secondary"
								size="xs"
								aria-label={plannerText('add', { event: eventLabel })}
								onClick={() => onAdd(kind, target?.index)}>
								{eventLabel}
								<Icon name="plus" />
							</Button>
						);
					})}
				</div>
			</div>
			{/* One flat list placed on grid rows, so an edit that moves a bar to another row keeps the same element and its focus. */}
			<ContextMenu
				className={LANE_GRID_CLASSES}
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
					onPointerDownCapture: noteMenuKey,
					onTouchStartCapture: noteMenuKey,
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
								{plannerText('menu.duplicate')}
							</MenuItem>
							<MenuItem layout="block" onClick={acted(() => (renameAfterMenu.current = refKey(menuEvent.ref)))}>
								{plannerText('menu.rename')}
							</MenuItem>
							<MenuItem
								layout="block"
								onClick={acted(() => {
									onSelect(menuEvent.ref);
									onActivate(menuEvent.ref);
								})}>
								{plannerText('menu.edit')}
							</MenuItem>
							{menuEvent.ref.kind === 'bossHealth' && (
								<MenuItem layout="block" onClick={acted(() => hold(menuEvent.ref))}>
									{plannerText('menu.hold_health')}
								</MenuItem>
							)}
							{menuEvent.ref.kind === 'downtime' && plan.downtimes[menuEvent.ref.index]?.targetIndex === 0 && (
								<MenuItem
									layout="block"
									disabled={!!holdBlocked}
									data-tooltip-id={holdBlocked ? tooltipId : undefined}
									data-tip={holdBlocked}
									onClick={acted(() => holdBoss(menuEvent.ref))}>
									{plannerText('menu.hold_boss_health')}
								</MenuItem>
							)}
							<MenuItem layout="block" onClick={acted(() => remove(menuEvent.ref))}>
								{plannerText('menu.delete')}
							</MenuItem>
						</>
					)
				}>
				{Array.from({ length: rowCount }, (_, row) => (
					<div key={row} aria-hidden="true" />
				))}
				{domOrder.flatMap(event => {
					const name = eventName(plan, event.ref);
					return event.repeats.map((repeat, index) => {
						const label = occurrenceName(name, index + 2);
						return (
							<div
								key={`${refKey(event.ref)}:${index}`}
								aria-hidden="true"
								data-tooltip-id={tooltipId}
								data-tip={serializeTip(occurrenceTip(eventTooltip(plan, event.ref, baseSeconds, targetNames, repeat), label))}
								className={`absolute left-planner-pos col-start-1 row-(--row) bg-(--bar)/40 bg-(image:--planner-ticks) ${GHOST_CLASSES}`}
								style={{
									...barVars(
										event.row,
										fractionToSeconds(repeat.start, baseSeconds),
										fractionToSeconds(repeat.end - repeat.start, baseSeconds),
									),
									...tickVars(event.tickSeconds),
								}}>
								<span className="min-w-0 truncate pl-1">{name}</span>
								<span className="shrink-0 px-1">{plannerText('occurrence_number', { number: index + 2 })}</span>
							</div>
						);
					});
				})}
				{endSegment && endTip && (
					<div
						role="img"
						aria-label={plannerText('tip.health_sentence', endTip)}
						data-tooltip-id={tooltipId}
						data-tip={serializeTip({
							title: plannerText('tip.boss_health'),
							lines: [plannerText('tip.health', endTip), plannerText('tip.range', { start: endTip.start, end: endTip.end })],
						})}
						className={`absolute left-planner-pos col-start-1 row-(--row) border border-dashed border-white/50 bg-(--bar)/40 ${GHOST_CLASSES}`}
						style={barVars(0, fractionToSeconds(endSegment.start, baseSeconds), fractionToSeconds(endSegment.end - endSegment.start, baseSeconds))}>
						<span className="min-w-0 truncate px-1">{plannerText('health_end')}</span>
					</div>
				)}
				{domOrder.map(event => {
					const key = refKey(event.ref);
					return (
						<PlannerEventBar
							key={key}
							dataKey={key}
							ref={element => {
								if (element) bars.current.set(key, element);
								else bars.current.delete(key);
							}}
							name={eventName(plan, event.ref)}
							label={event.ref.kind === 'bossHealth' ? healthLabel(plan, segments, event.ref) : undefined}
							flat={isHold(segments, event.ref)}
							window={followed(segments, dragging, event)}
							row={event.row}
							startFixed={event.startFixed}
							openEnd={event.openEnd}
							fixedWidth={event.fixedWidth}
							tickSeconds={event.tickSeconds}
							jitter={event.jitter}
							baseSeconds={baseSeconds}
							snapSeconds={snapSeconds}
							pixelsPerSecond={pixelsPerSecond}
							valueText={window => eventValueText(plan, event.ref, baseSeconds, window)}
							tooltipId={tooltipId}
							tooltipText={window => tip(event.ref, window)}
							selected={key === selectedKey}
							invalid={invalid.has(key)}
							tabIndex={key === tabbable ? 0 : -1}
							renaming={key === renaming}
							onChange={window => onChange(withWindow(plan, event.ref, window, baseSeconds))}
							onSelect={() => onSelect(event.ref)}
							onActivate={() => onActivate(event.ref)}
							onRemove={() => remove(event.ref)}
							onDuplicate={delta => duplicate(event.ref, delta)}
							onRenameStart={() => setRenaming(key)}
							onRename={(name, refocus) => rename(event.ref, name, refocus)}
							adjust={event.ref.kind === 'bossHealth' ? adjustHealth(segments, event.ref) : undefined}
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
