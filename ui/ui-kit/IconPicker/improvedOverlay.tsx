import type { ResolvedIcon } from '@ui-kit/Field/types';

import { ImprovedAnchor, type ImprovedAnchorEvents } from './ImprovedAnchor';

export interface ImprovedState {
	states: number;
	value: number;
	improvedIcon?: ResolvedIcon;
	improvedIcon2?: ResolvedIcon;
}

export interface ImprovedOverlayProps extends ImprovedState {
	testIdPrefix: string;
	disabled?: boolean;
	/** Left out, the badges are display only. */
	events?: ImprovedAnchorEvents;
}

/**
 * The counter text and improved badges an icon of `states` carries, for an `IconPicker` and a multi icon trigger alike.
 * `badges` is undefined when there are none, so a swatch without them stays bare.
 */
export const improvedOverlay = ({ states, value, improvedIcon, improvedIcon2, testIdPrefix, disabled, events }: ImprovedOverlayProps) => {
	const improved1 = states >= 3 ? improvedIcon : undefined;
	const improved2 = states >= 4 ? improvedIcon2 : undefined;
	const badge = (icon: ResolvedIcon | undefined, n: 1 | 2, corner: 'left' | 'right') =>
		icon && <ImprovedAnchor icon={icon} testId={`${testIdPrefix}-improved${n}`} active={value > n} disabled={disabled} corner={corner} events={events} />;
	return {
		text: !improvedIcon && (states > 3 || states === 0) ? String(value) : undefined,
		badges:
			improved1 || improved2 ? (
				<>
					{/* Both badges stay visible in opposite corners: at a state below its own, a quadstate
					    icon's grey second badge is the only affordance that a fourth state exists. */}
					{badge(improved1, 1, 'right')}
					{badge(improved2, 2, 'left')}
				</>
			) : undefined,
	};
};
