import i18n from '@i18n/config';
import { Field } from '@ui-kit/Field';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconMenu } from '@ui-kit/IconSwatch';
import { isRightClick } from '@ui-kit/utils/dom';
import clsx from 'clsx';
import { Children, type MouseEvent, type ReactNode, useId, useRef } from 'react';

export interface MultiIconPickerProps {
	label?: string;
	icon?: ResolvedIcon;
	active: boolean;
	onClear: () => void;
	/** One menu entry per child, each an icon toggle. The popup unmounts while closed, and the children with it. */
	children: ReactNode;
	testId?: string;
}

export const MultiIconPicker = ({ label, icon, active, onClear, children, testId }: MultiIconPickerProps) => {
	const labelId = useId();
	const controlRef = useRef<HTMLAnchorElement>(null);
	const menu = (
		<IconMenu
			trigger={{
				icon,
				active,
				className: clsx(!active && 'grayscale'),
				testId: testId ?? 'multi-icon-picker-root',
				// The trigger is a bare anchor carrying a background image, so it announces nothing without this. Base UI
				// points the popup's `aria-labelledby` at the trigger, and accessible-name resolution does not chase a
				// second `aria-labelledby` hop through that id, so this literal string is what names the popup,
				// independently of the trigger's own `aria-labelledby` below.
				'aria-label': label,
				...(label ? { ref: controlRef, 'aria-labelledby': labelId } : {}),
				onContextMenu: (event: MouseEvent) => event.preventDefault(),
				onMouseDown: (event: MouseEvent) => {
					if (isRightClick(event.nativeEvent)) onClear();
				},
			}}
			side="right"
			positionerProps={{ 'data-testid': 'multi-icon-picker-positioner' }}
			className="grid-flow-col"
			// `role="group"`, not the `menu` Base UI would give it. A menu's children must be menuitems, and these are icon toggles — `Menu.Item` would close the popup on every click, and toggling several buffs in one visit is the whole point of this control.
			popupProps={{ role: 'group', 'data-testid': 'multi-icon-picker-menu' }}>
			<li>
				<a
					className="ui-icon-picker-swatch p-0 filter-[opacity(0.7)] hover:filter-none"
					data-testid="icon-picker-clear"
					aria-label={i18n.t('common.clear')}
					onClick={onClear}
				/>
			</li>
			{Children.map(children, child => (
				<li className="opacity-70 hover:opacity-100">{child}</li>
			))}
		</IconMenu>
	);

	if (!label) return menu;

	return <Field role="group" label={label} layout="icon" leading={menu} labelId={labelId} controlRef={controlRef} />;
};
