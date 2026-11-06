import { Ruler } from '@features/results/components/Timeline/rotation/ruler';
import { DEFAULT_PPS, fitWidth, ZoomController } from '@features/results/components/Timeline/rotation/zoom';
import type { EncounterPlan } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { usePointerPan } from '@ui-kit/hooks/usePointerPan';
import { Icon } from '@ui-kit/Icon';
import { Tooltip } from '@ui-kit/Tooltip';
import { cssVars } from '@ui-kit/utils/css';
import { findScrollParent } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';

import type { PlanEventKind, PlannerSelection } from '../../model/lanes';
import { isEventRef, PLAN_LANES } from '../../model/lanes';
import type { PlanIssue } from '../../model/validate';
import { plannerText } from '../EncounterPlanner/utils/configs';
import { PlannerLane } from '../PlannerLane';
import { LANE_CLASSES, LANE_LABEL_CLASSES, refKey } from '../PlannerLane/utils';
import { PlannerTip } from '../PlannerTip';
import { PlannerToolbar } from '../PlannerToolbar';

// react-tooltip reads the anchor once when it opens, so a bar whose `data-tip` moves during a drag is watched here.
const AnchorTip = ({ anchor }: { anchor: Element | null }) => {
	const [text, setText] = useState(() => anchor?.getAttribute('data-tip'));
	useEffect(() => {
		if (!anchor) return;
		const read = () => setText(anchor.getAttribute('data-tip'));
		read();
		const observer = new MutationObserver(read);
		observer.observe(anchor, { attributes: true, attributeFilter: ['data-tip'] });
		return () => observer.disconnect();
	}, [anchor]);
	return <PlannerTip text={text} />;
};

export interface PlannerTimelineProps {
	plan: EncounterPlan;
	/** The fight length the fractions in the plan are drawn against. */
	baseSeconds: number;
	selected?: PlannerSelection;
	issues: ReadonlyArray<PlanIssue>;
	/** The encounter's target names: one row each, and what a downtime's tooltip reads. */
	targetNames: ReadonlyArray<string>;
	/** `undefined` clears the selection, as removing an event does. */
	onSelect: (selection: PlannerSelection | undefined) => void;
	/** Enter on a bar, after `onSelect`; a click or a drag only selects, so this is where focus may move to the panel. */
	onActivate: (selection: PlannerSelection) => void;
	/** Called once per finished drag or key press, with a new plan. */
	onChange: (plan: EncounterPlan) => void;
	/** A downtime added on a target's row takes that target; one added elsewhere takes the first. */
	onAdd: (kind: PlanEventKind, targetIndex?: number) => void;
	onAddTarget: () => void;
}

const LABEL_WIDTH = '12rem';
// The boss lane is the targets' rows.
const EVENT_LANES = PLAN_LANES.filter(lane => lane !== 'boss');

const SectionHeader = ({ label }: { label: string }) => (
	<div className="flex min-w-full border-b border-gray-700 bg-white/10">
		<span className="sticky left-0 px-2 py-1 text-xs font-bold tracking-wide text-white uppercase">{label}</span>
	</div>
);

export const PlannerTimeline = ({
	plan,
	baseSeconds,
	selected,
	issues,
	targetNames,
	onSelect,
	onActivate,
	onChange,
	onAdd,
	onAddTarget,
}: PlannerTimelineProps) => {
	const rootRef = useRef<HTMLDivElement>(null);
	const cornerRef = useRef<HTMLDivElement>(null);
	const rulerViewportRef = useRef<HTMLDivElement>(null);
	const rulerTrackRef = useRef<HTMLDivElement>(null);
	const gutterRef = useRef<HTMLDivElement>(null);
	const scrollerRef = useRef<HTMLDivElement>(null);
	const zoomRef = useRef<ZoomController | null>(null);
	const fitRef = useRef<(() => void) | null>(null);
	const followsWidth = useRef(true);
	const rulerRef = useRef<Ruler | null>(null);
	const frameHandle = useRef<number | null>(null);
	const duration = useRef(baseSeconds);
	const tooltipId = `planner-tip-${useId().replace(/:/g, '')}`;
	const [snapSeconds, setSnapSeconds] = useState(1);

	const draw = useCallback(() => {
		frameHandle.current = null;
		const scroller = scrollerRef.current;
		const zoom = zoomRef.current;
		if (!scroller || !zoom) return;
		rulerRef.current?.draw({
			scrollLeft: scroller.scrollLeft,
			pps: zoom.pps,
			duration: duration.current,
			width: rulerViewportRef.current?.clientWidth ?? 0,
			inset: gutterRef.current?.offsetLeft ?? 0,
		});
	}, []);

	const schedule = useCallback(() => {
		if (frameHandle.current == null) frameHandle.current = requestAnimationFrame(draw);
	}, [draw]);

	useEffect(() => {
		const root = rootRef.current;
		const scroller = scrollerRef.current;
		const rulerTrack = rulerTrackRef.current;
		if (!root || !scroller || !rulerTrack) return;

		// The whole fight shows at the scroller's width, and keeps showing as that width or the duration changes (the
		// dialog's own scrollbar takes some width), until the user zooms; Fit returns to it.
		let fitting = false;
		let fittedWidth = 0;
		const fit = () => {
			followsWidth.current = true;
			fittedWidth = fitWidth(scroller);
			fitting = true;
			zoomRef.current?.fitToWidth();
			fitting = false;
		};

		rulerRef.current = new Ruler(rulerTrack);
		const zoom = new ZoomController({
			scroller,
			styleHosts: [root],
			labelWidth: () => cornerRef.current?.offsetWidth ?? 0,
			inset: () => gutterRef.current?.offsetLeft ?? 0,
			// Bars take the arrow keys, so the zoom never handles them and never scrolls vertically.
			scrollVerticalBy: () => {},
			onChange: () => {
				if (!fitting) followsWidth.current = false;
				schedule();
			},
		});
		zoom.setDuration(duration.current);
		zoom.attach();
		zoomRef.current = zoom;
		fitRef.current = fit;

		const observer = new ResizeObserver(() => {
			const width = fitWidth(scroller);
			if (followsWidth.current && width > 0 && width !== fittedWidth) fit();
			schedule();
		});
		observer.observe(root);
		observer.observe(scroller);
		scroller.addEventListener('scroll', schedule, { passive: true });

		return () => {
			if (frameHandle.current != null) cancelAnimationFrame(frameHandle.current);
			frameHandle.current = null;
			scroller.removeEventListener('scroll', schedule);
			observer.disconnect();
			zoom.dispose();
			zoomRef.current = null;
			fitRef.current = null;
			rulerRef.current = null;
		};
	}, [schedule]);

	useEffect(() => {
		duration.current = baseSeconds;
		zoomRef.current?.setDuration(baseSeconds);
		if (followsWidth.current) fitRef.current?.();
		schedule();
	}, [baseSeconds, schedule]);

	// The scroller grows with its lanes, so the dialog's body is what moves them vertically.
	const { panning, handlers: pan } = usePointerPan({ verticalScroller: () => (rootRef.current ? findScrollParent(rootRef.current) : null) });
	const pixelsPerSecond = useCallback(() => zoomRef.current?.pps ?? DEFAULT_PPS, []);
	const invalid = useMemo(() => new Set(issues.filter(issue => issue.severity !== 'info').map(issue => refKey(issue.ref))), [issues]);
	const laneProps = {
		plan,
		baseSeconds,
		snapSeconds,
		targetNames,
		tooltipId,
		pixelsPerSecond,
		selected: isEventRef(selected) ? selected : undefined,
		invalid,
		onSelect,
		onActivate,
		onChange,
		onAdd,
	};
	const targetsLabel = plannerText('targets');
	const eventLanesLabel = plannerText('event_lanes');
	const targetLabel = i18n.t('settings_tab.encounter.target');

	return (
		<div
			ref={rootRef}
			className="flex min-w-0 flex-col gap-2 text-white"
			style={cssVars({ '--pps': `${DEFAULT_PPS}px`, '--label-w': LABEL_WIDTH, '--duration': String(baseSeconds) })}>
			<PlannerToolbar
				snapSeconds={snapSeconds}
				onSnapChange={setSnapSeconds}
				onZoomIn={() => zoomRef.current?.stepIn()}
				onZoomOut={() => zoomRef.current?.stepOut()}
				onFit={() => fitRef.current?.()}
				onAdd={onAdd}
			/>
			<div className="flex h-6 shrink-0">
				<div ref={cornerRef} className="w-(--label-w) shrink-0 grow-0 basis-(--label-w)" />
				<div ref={rulerViewportRef} className="relative min-w-0 shrink grow basis-0 overflow-hidden border-b border-white text-xs font-bold">
					<div ref={gutterRef} aria-hidden="true" className="absolute left-(--planner-axis-gutter)" />
					<div ref={rulerTrackRef} className="ui-timeline-ruler-track" />
				</div>
			</div>
			<div
				ref={scrollerRef}
				role="region"
				aria-label={plannerText('timeline')}
				className={clsx('relative min-w-0 cursor-grab overflow-x-auto overflow-y-hidden', panning && 'cursor-grabbing select-none')}
				{...pan}>
				<div className="w-planner-content-w min-w-full">
					<div role="group" aria-label={targetsLabel}>
						<SectionHeader label={targetsLabel} />
						{targetNames.map((_, index) => (
							<PlannerLane
								key={index}
								{...laneProps}
								lane="boss"
								target={{
									index,
									selected: selected?.kind === 'target' && selected.index === index,
									onSelect: () => onSelect({ kind: 'target', index }),
									onActivate: () => onActivate({ kind: 'target', index }),
								}}
							/>
						))}
						<div className={LANE_CLASSES}>
							<div className={LANE_LABEL_CLASSES}>
								<Button
									variant="outline-secondary"
									size="xs"
									className="self-start"
									aria-label={plannerText('add', { event: targetLabel })}
									onClick={onAddTarget}>
									<Icon name="plus" /> {targetLabel}
								</Button>
							</div>
						</div>
					</div>
					<div role="group" aria-label={eventLanesLabel}>
						<SectionHeader label={eventLanesLabel} />
						{EVENT_LANES.map(lane => (
							<PlannerLane key={lane} {...laneProps} lane={lane} />
						))}
					</div>
				</div>
			</div>
			<Tooltip id={tooltipId} render={({ activeAnchor }) => <AnchorTip anchor={activeAnchor} />} />
		</div>
	);
};
