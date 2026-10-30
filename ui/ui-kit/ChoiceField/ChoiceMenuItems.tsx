import { Menu as BaseMenu } from '@base-ui/react/menu';
import { Menu } from '@ui-kit/Menu';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type PointerEvent, useRef } from 'react';

import { OptionContent } from './OptionContent';
import type { ChoiceEntry, ChoiceOption } from './types';
import { isChoiceOption, isChoiceSection } from './utils';

export interface ChoiceMenuItemsProps<V> {
	entries: ReadonlyArray<ChoiceEntry<V>>;
	/** The option's position in the flat list, which is what the radio group compares. */
	indexOf: (option: ChoiceOption<V>) => number;
	tooltipId: string;
	onSelect: (index: number) => void;
	/** Closes the whole menu. */
	onClose: () => void;
}

const tooltipProps = <V,>(tooltipId: string, option: ChoiceOption<V> | undefined) =>
	tooltipAnchorProps(option?.tooltip === undefined ? undefined : tooltipId, option?.tooltip);

/**
 * The menu's rows, one level at a time.
 *
 * A submenu is a real nested popup. A submenu opened by an option keeps that option selectable —
 * the owner of a pet, or a spell's highest rank, both opens the submenu and can be chosen. That
 * trigger is not a radio item, so it reports no `aria-checked`.
 */
export const ChoiceMenuItems = <V,>({ entries, indexOf, tooltipId, onSelect, onClose }: ChoiceMenuItemsProps<V>) => {
	// Read at `pointerdown` because a `click` names its pointer type only in newer browsers.
	const pointerType = useRef('');

	/**
	 * Picks the option a submenu's row stands for. A tap also opens the submenu (Base UI opens one
	 * on a touch click, where a mouse opens it on hover), so it leaves the menu open with the other
	 * choices showing; a mouse or keyboard pick is final and closes it.
	 */
	const selectTrigger = (index: number) => {
		onSelect(index);
		if (pointerType.current !== 'touch') onClose();
		pointerType.current = '';
	};

	return (
		<>
			{entries.map((entry, position) => {
				if (isChoiceSection(entry))
					return (
						<li key={`section-${position}`} role="none">
							<BaseMenu.Group render={<ul />} className="m-0 list-none p-0">
								<BaseMenu.GroupLabel className="ui-choice-field-section">{entry.label}</BaseMenu.GroupLabel>
								<ChoiceMenuItems entries={entry.options} indexOf={indexOf} tooltipId={tooltipId} onSelect={onSelect} onClose={onClose} />
							</BaseMenu.Group>
						</li>
					);
				const option = isChoiceOption(entry) ? entry : undefined;
				if (option && !option.submenu)
					return (
						<li key={`option-${indexOf(option)}`} role="none">
							<BaseMenu.RadioItem
								render={<button type="button" />}
								nativeButton
								value={indexOf(option)}
								closeOnClick
								className={clsx('ui-menu-item', 'ui-menu-item-row', option.className)}
								{...tooltipProps(tooltipId, option)}>
								<OptionContent option={option} />
							</BaseMenu.RadioItem>
						</li>
					);
				return (
					<Menu
						key={`submenu-${position}`}
						submenu
						surface="menu"
						align="start"
						side="right"
						popupProps={{ 'data-testid': 'choice-field-submenu' }}
						triggerRender={<button type="button" />}
						triggerProps={{
							onPointerDown: (event: PointerEvent) => (pointerType.current = event.pointerType),
							onClick: option ? () => selectTrigger(indexOf(option)) : undefined,
							className: clsx('ui-menu-item', 'ui-menu-item-row', option?.className),
							...tooltipProps(tooltipId, option),
						}}
						trigger={
							option ? (
								<>
									<OptionContent option={option} />
									{' »'}
								</>
							) : (
								entry.label
							)
						}>
						<ChoiceMenuItems entries={entry.submenu!} indexOf={indexOf} tooltipId={tooltipId} onSelect={onSelect} onClose={onClose} />
					</Menu>
				);
			})}
		</>
	);
};
