import { Field, hasFieldFrame } from '@ui-kit/Field';
import type { ControlledProps, FieldFrameProps, ResolvedIcon } from '@ui-kit/Field/types';
import { IconSwatch, type IconSwatchProps } from '@ui-kit/IconSwatch';
import { isRightClick } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import { type MouseEvent, useId, useRef } from 'react';

import { ImprovedAnchor } from './ImprovedAnchor';

/** No `size`: the control is an icon swatch, not a form-height control. */
export interface IconPickerProps extends Omit<FieldFrameProps, 'size'>, ControlledProps<number>, Pick<IconSwatchProps, 'keptTooltip'> {
	/** The number of states the icon cycles through: 2 is on/off, 0 is unlimited. */
	states: number;
	icon?: ResolvedIcon;
	/** Shown from states 3 up. */
	improvedIcon?: ResolvedIcon;
	/** Shown from states 4 up. */
	improvedIcon2?: ResolvedIcon;
}

/** Without a label, description or layout the swatch renders bare; with any of them it sits in a `Field` laid out as `icon`, whatever layout the caller passes. */
export const IconPicker = ({
	states,
	icon,
	improvedIcon,
	improvedIcon2,
	keptTooltip,
	value,
	onChange,
	disabled,
	className,
	testId,
	...frameProps
}: IconPickerProps) => {
	const labelId = useId();
	const controlRef = useRef<HTMLAnchorElement>(null);
	const fillImproved1 = states >= 3 && !!improvedIcon;
	const fillImproved2 = states >= 4 && !!improvedIcon2;
	const showCounterText = !improvedIcon && (states > 3 || states === 0);
	const rooted = hasFieldFrame(frameProps);

	const handleLeftClick = () => {
		// An <a> cannot be disabled, so the state change has to be stopped here.
		if (disabled) return;

		if (states === 0 || value + 1 < states) {
			onChange(value + 1);
		} else if (value > 0) {
			onChange(0);
		}
	};

	const handleRightClick = () => {
		if (disabled) return;

		if (value > 0) {
			onChange(value - 1);
		} else {
			onChange(states === 0 ? 1 : states - 1);
		}
	};

	// The improved icons are anchors beside the swatch, not inside it, so each carries these too.
	const stateEvents = {
		onActivate: handleLeftClick,
		onContextMenu: (event: MouseEvent) => event.preventDefault(),
		onMouseDown: (event: MouseEvent) => {
			if (isRightClick(event.nativeEvent)) {
				event.preventDefault();
				handleRightClick();
			}
		},
	};

	const badges = (fillImproved1 || fillImproved2) && (
		<>
			{/* Both badges stay visible in opposite corners: at a state below its own, a quadstate
			    icon's grey second badge is the only affordance that a fourth state exists. */}
			{fillImproved1 && improvedIcon && (
				<ImprovedAnchor icon={improvedIcon} testId="icon-input-improved1" active={value > 1} disabled={disabled} corner="right" {...stateEvents} />
			)}
			{fillImproved2 && improvedIcon2 && (
				<ImprovedAnchor icon={improvedIcon2} testId="icon-input-improved2" active={value > 2} disabled={disabled} corner="left" {...stateEvents} />
			)}
		</>
	);

	const swatch = (
		<IconSwatch
			icon={icon}
			href={icon?.href}
			active={value > 0}
			disabled={disabled}
			text={showCounterText ? String(value) : undefined}
			badges={badges || undefined}
			keptTooltip={keptTooltip}
			className={clsx(value === 0 && 'grayscale', !rooted && className)}
			testId={testId ?? 'icon-picker-root'}
			// The glyph is a background image, so without this the anchor announces nothing.
			aria-label={icon?.name || undefined}
			{...(rooted ? { ref: controlRef, 'aria-labelledby': labelId } : {})}
			{...stateEvents}
		/>
	);

	if (!rooted) return swatch;

	return <Field {...frameProps} layout="icon" className={className} disabled={disabled} leading={swatch} labelId={labelId} controlRef={controlRef} />;
};
