import i18n from '@i18n/config';
import { Field } from '@ui-kit/Field';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconEnumPickerDirection } from '@ui-kit/IconEnumPicker';
import { improvedOverlay, type ImprovedState } from '@ui-kit/IconPicker/improvedOverlay';
import { IconMenu, type IconSwatchProps } from '@ui-kit/IconSwatch';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { isRightClick } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import { Children, type MouseEvent, type ReactNode, useId, useRef } from 'react';

export interface MultiIconPickerProps extends Pick<IconSwatchProps, 'keptTooltip'> {
	label?: string;
	icon?: ResolvedIcon;
	active: boolean;
	/**
	 * The shown input's own state, drawn on the trigger as an `IconPicker` draws it: the improved badges and the
	 * counter text. Display only, so the trigger keeps its hover and right click. Left out, the trigger carries none.
	 */
	state?: ImprovedState;
	onClear: () => void;
	/** Horizontal, the default, opens the menu to the right as a row; vertical opens it below, in `numColumns` columns. */
	direction?: IconEnumPickerDirection;
	numColumns?: number;
	/** Shown on hovering the trigger, and its name where there is no label. */
	tooltip?: string;
	/** One menu entry per child, each an icon toggle; a child that renders nothing takes no entry. The popup unmounts while closed, and the children with it. */
	children: ReactNode;
	testId?: string;
}

export const MultiIconPicker = ({
	label,
	icon,
	active,
	state,
	onClear,
	direction = IconEnumPickerDirection.Horizontal,
	numColumns,
	tooltip,
	children,
	keptTooltip,
	testId,
}: MultiIconPickerProps) => {
	const labelId = useId();
	const tooltipId = useId();
	const horizontal = direction === IconEnumPickerDirection.Horizontal;
	const controlRef = useRef<HTMLAnchorElement>(null);
	const overlay = state && improvedOverlay({ ...state, testIdPrefix: 'multi-icon-picker' });
	const menu = (
		<IconMenu
			trigger={{
				icon,
				active,
				keptTooltip,
				text: overlay?.text,
				badges: overlay?.badges,
				className: clsx(!active && 'grayscale'),
				testId: testId ?? 'multi-icon-picker-root',
				// The trigger is a bare anchor carrying a background image, so it announces nothing without this. Base UI
				// points the popup's `aria-labelledby` at the trigger, and accessible-name resolution does not chase a
				// second `aria-labelledby` hop through that id, so this literal string is what names the popup,
				// independently of the trigger's own `aria-labelledby` below.
				'aria-label': label ?? tooltip,
				...(label ? { ref: controlRef, 'aria-labelledby': labelId } : {}),
				...tooltipAnchorProps(tooltip ? tooltipId : undefined, tooltip),
				onContextMenu: (event: MouseEvent) => event.preventDefault(),
				onMouseDown: (event: MouseEvent) => {
					if (isRightClick(event.nativeEvent)) onClear();
				},
			}}
			side={horizontal ? 'right' : 'bottom'}
			positionerProps={{ 'data-testid': 'multi-icon-picker-positioner' }}
			className={clsx(horizontal && 'grid-flow-col')}
			// `role="group"`, not the `menu` Base UI would give it. A menu's children must be menuitems, and these are icon toggles — `Menu.Item` would close the popup on every click, and toggling several buffs in one visit is the whole point of this control.
			popupProps={{
				role: 'group',
				'data-testid': 'multi-icon-picker-menu',
				style: { gridTemplateColumns: !horizontal && numColumns ? `repeat(${numColumns}, 1fr)` : undefined },
			}}>
			<li>
				<a
					className="ui-icon-picker-swatch p-0 filter-[opacity(0.7)] hover:filter-none"
					data-testid="icon-picker-clear"
					aria-label={i18n.t('common.clear')}
					onClick={onClear}
				/>
			</li>
			{Children.map(children, child => (
				<li className="opacity-70 empty:hidden hover:opacity-100">{child}</li>
			))}
		</IconMenu>
	);

	const triggerTooltip = tooltip && <Tooltip id={tooltipId} />;

	if (!label)
		return (
			<>
				{menu}
				{triggerTooltip}
			</>
		);

	return (
		<Field role="group" label={label} layout="icon" leading={menu} labelId={labelId} controlRef={controlRef}>
			{triggerTooltip}
		</Field>
	);
};
