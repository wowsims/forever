import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import type { TalentConfig } from '@sim/talents/config';
import { isRightClick } from '@ui-kit/utils/dom';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { useEffect, useRef } from 'react';

export interface TalentPickerProps<TalentsProto> {
	config: TalentConfig<TalentsProto>;
	points: number;
	canAdd: boolean;
	allPointsSpent: boolean;
	zIndex?: number;
	onSetPoints: (newPoints: number) => void;
}

const LONG_TOUCH_MS = 750;

export const TalentPicker = <TalentsProto,>({ config, points, canAdd, allPointsSpent, zIndex, onSetPoints }: TalentPickerProps<TalentsProto>) => {
	const rootRef = useRef<HTMLAnchorElement>(null);
	const definitionId = config.definitionId ?? 0;
	// The icon lookup leaves the rank out, so spending a point does not re-seed it and blank the
	// cell for a frame -- a talent is one spell with one icon. Only the wowhead link carries the
	// rank, and that needs no lookup at all.
	const { iconUrl } = useActionId(ActionId.fromTalent(config.spellId, 1, definitionId));
	const href = ActionId.makeSpellUrl(config.spellId, Math.max(points, 1), definitionId);
	const isFull = points >= config.maxPoints;

	const spend = () => onSetPoints(points + 1);
	const unspend = () => onSetPoints(points - 1);
	const clear = () => onSetPoints(0);
	// A tap cycles rather than selects: past the last rank it wraps back to nothing (talents_picker.tsx:410-414).
	const cycle = () => onSetPoints(points + 1 > config.maxPoints ? 0 : points + 1);

	// The native listeners are attached once, so they read the current handlers through a ref rather than re-attaching on every talents-string change.
	const handlers = useRef({ spend, unspend, clear, cycle });
	handlers.current = { spend, unspend, clear, cycle };

	useEffect(() => {
		const elem = rootRef.current;
		if (!elem) return;

		let timer: number | undefined;
		const cancel = () => {
			if (timer === undefined) return false;
			clearTimeout(timer);
			timer = undefined;
			return true;
		};
		const onTouchStart = (event: TouchEvent) => {
			event.preventDefault();
			timer = window.setTimeout(() => {
				timer = undefined;
				handlers.current.clear();
			}, LONG_TOUCH_MS);
		};
		const onTouchEnd = (event: TouchEvent) => {
			event.preventDefault();
			// The long press already fired and cleared the timer; releasing must not then re-spend it.
			if (!cancel()) return;
			handlers.current.cycle();
		};

		elem.addEventListener('touchmove', cancel);
		// React registers touchstart as a passive listener, where preventDefault is ignored.
		elem.addEventListener('touchstart', onTouchStart, { passive: false });
		elem.addEventListener('touchend', onTouchEnd, { passive: false });
		return () => {
			cancel();
			elem.removeEventListener('touchmove', cancel);
			elem.removeEventListener('touchstart', onTouchStart);
			elem.removeEventListener('touchend', onTouchEnd);
		};
	}, []);

	return (
		<WowheadAnchor
			ref={rootRef}
			className="ui-talent-picker-root"
			testId="talent-picker-root"
			href={href || undefined}
			data-points={String(points)}
			data-max-points={String(config.maxPoints)}
			data-can-add={String(canAdd)}
			data-full={String(isFull)}
			data-all-spent={String(allPointsSpent)}
			style={{
				gridRow: config.location.rowIdx + 1,
				gridColumn: config.location.colIdx + 1,
				zIndex,
				backgroundImage: iconUrl ? `url('${iconUrl}')` : undefined,
			}}
			// `mousedown` rather than the click commits, which is why a right click reaches it at all, so the click only cancels the wowhead link.
			navigate={false}
			onContextMenu={event => event.preventDefault()}
			onMouseDown={event => (isRightClick(event.nativeEvent) ? handlers.current.unspend() : handlers.current.spend())}>
			<span className="ui-talent-picker-points" data-testid="talent-picker-points">
				{points}/{config.maxPoints}
			</span>
		</WowheadAnchor>
	);
};
