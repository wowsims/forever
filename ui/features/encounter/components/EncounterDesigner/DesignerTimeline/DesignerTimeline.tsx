import type { EncounterDesign } from '@generated/proto/common';
import { Tooltip } from '@ui-kit/Tooltip';
import { cssVars } from '@ui-kit/utils/css';
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react';

import { Ruler } from '../../../../results/components/Timeline/rotation/ruler';
import { DEFAULT_PPS, ZoomController } from '../../../../results/components/Timeline/rotation/zoom';
import type { DesignEventKind, DesignEventRef } from '../../../model/design/lanes';
import { DESIGN_LANES } from '../../../model/design/lanes';
import type { DesignIssue } from '../../../model/design/validate';
import { DesignerLane } from '../DesignerLane';
import { refKey } from '../DesignerLane/utils';
import { DesignerTip } from '../DesignerTip';
import { DesignerToolbar } from '../DesignerToolbar';
import { designerText } from '../utils/configs';

export interface DesignerTimelineProps {
	design: EncounterDesign;
	/** The fight length the fractions in the design are drawn against. */
	baseSeconds: number;
	selected?: DesignEventRef;
	issues: ReadonlyArray<DesignIssue>;
	/** The encounter's target names, which a downtime's tooltip reads. */
	targetNames: ReadonlyArray<string>;
	/** `undefined` clears the selection, as removing an event does. */
	onSelect: (ref: DesignEventRef | undefined) => void;
	/** Enter on a bar, after `onSelect`; a click or a drag only selects, so this is where focus may move to the panel. */
	onActivate: (ref: DesignEventRef) => void;
	/** Called once per finished drag or key press, with a new design. */
	onChange: (design: EncounterDesign) => void;
	onAdd: (kind: DesignEventKind) => void;
}

const LABEL_WIDTH = '12rem';

export const DesignerTimeline = ({ design, baseSeconds, selected, issues, targetNames, onSelect, onActivate, onChange, onAdd }: DesignerTimelineProps) => {
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
	const tooltipId = `designer-tip-${useId().replace(/:/g, '')}`;
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
			fittedWidth = scroller.clientWidth;
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
			if (followsWidth.current && scroller.clientWidth > 0 && scroller.clientWidth !== fittedWidth) fit();
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

	const pixelsPerSecond = useCallback(() => zoomRef.current?.pps ?? DEFAULT_PPS, []);
	const invalid = useMemo(() => new Set(issues.filter(issue => issue.severity !== 'info').map(issue => refKey(issue.ref))), [issues]);

	return (
		<div
			ref={rootRef}
			className="flex min-w-0 flex-col gap-2 text-white"
			style={cssVars({ '--pps': `${DEFAULT_PPS}px`, '--label-w': LABEL_WIDTH, '--duration': String(baseSeconds) })}>
			<DesignerToolbar
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
					<div ref={gutterRef} aria-hidden="true" className="absolute left-(--designer-axis-gutter)" />
					<div ref={rulerTrackRef} className="ui-timeline-ruler-track" />
				</div>
			</div>
			<div ref={scrollerRef} role="region" aria-label={designerText('timeline')} className="relative min-w-0 overflow-x-auto overflow-y-hidden">
				<div className="w-designer-content-w min-w-full">
					{DESIGN_LANES.map(lane => (
						<DesignerLane
							key={lane}
							lane={lane}
							design={design}
							baseSeconds={baseSeconds}
							snapSeconds={snapSeconds}
							targetNames={targetNames}
							tooltipId={tooltipId}
							pixelsPerSecond={pixelsPerSecond}
							selected={selected}
							invalid={invalid}
							onSelect={onSelect}
							onActivate={onActivate}
							onChange={onChange}
							onAdd={onAdd}
						/>
					))}
				</div>
			</div>
			<Tooltip id={tooltipId} render={({ activeAnchor }) => <DesignerTip text={activeAnchor?.getAttribute('data-tip')} />} />
		</div>
	);
};
