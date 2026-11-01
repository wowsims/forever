import { Select } from '@base-ui/react/select';
import { Field, hasFieldFrame, useControlSize } from '@ui-kit/Field';
import { FieldContext } from '@ui-kit/Field/context/FieldContext';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import { LocaleHtml, Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { useId, useMemo, useState } from 'react';

import { ChoiceMenu } from './ChoiceMenu';
import { OptionContent } from './OptionContent';
import type { ChoiceFieldProps, ChoiceOption } from './types';
import { flattenChoices, hasSubmenus, isChoiceOption, isChoiceSection } from './utils';

/**
 * One choice among options, on Base UI's `Select`; options filed under submenus open as a menu.
 *
 * The select and the menu compare an option's position in the flat list, never its value: `equals`
 * alone decides which option `value` selects, so a number enum and a message compare alike, and an
 * `undefined` value is an option like any other rather than an uncontrolled select.
 */
export const ChoiceField = <V,>({
	options,
	optionTree,
	value,
	onChange,
	disabled,
	equals = Object.is,
	variant = 'select',
	invalid,
	placeholder,
	renderValue,
	ariaLabel,
	'aria-describedby': describedBy,
	side,
	positionMethod,
	className,
	testId,
	...frameProps
}: ChoiceFieldProps<V>) => {
	const [open, setOpen] = useState(false);
	const tooltipId = `${useId()}-option`;
	const portalContainer = usePortalContainer();
	const size = useControlSize(frameProps.size);

	const flat = useMemo(() => flattenChoices(options), [options]);
	const indexOf = useMemo(() => new Map<ChoiceOption<V>, number>(flat.map((option, index) => [option, index])), [flat]);
	const selectedIndex = flat.findIndex(option => equals(option.value, value));
	const selected: ChoiceOption<V> | undefined = flat[selectedIndex];
	const select = (index: number) => onChange(flat[index].value);
	// Without a placeholder, the select look shows its first option while nothing matches; nothing is written back.
	const shown = selected ?? (variant === 'select' && placeholder === undefined ? flat[0] : undefined);

	const rooted = hasFieldFrame(frameProps);
	const trigger = shown ? renderValue ? renderValue(shown) : <OptionContent option={shown} /> : placeholder;
	const triggerProps = {
		id: frameProps.id,
		className: clsx(!rooted && ['ui-choice-field', className], 'ui-choice-field-trigger', shown?.className),
		'data-variant': variant,
		'data-size': variant === 'select' ? size : undefined,
		'data-testid': rooted ? undefined : testId,
		'aria-label': ariaLabel,
		'aria-describedby': describedBy,
		'aria-invalid': invalid || undefined,
		disabled,
	};

	const selectItem = (option: ChoiceOption<V>) => (
		<Select.Item
			key={indexOf.get(option)}
			value={indexOf.get(option)}
			className={clsx('ui-menu-item', 'ui-menu-item-row', 'ui-choice-field-option', option.className)}
			{...tooltipAnchorProps(option.tooltip === undefined ? undefined : tooltipId, option.tooltip)}>
			<OptionContent option={option} />
		</Select.Item>
	);

	const control =
		optionTree || hasSubmenus(options) ? (
			<ChoiceMenu
				entries={optionTree ?? options}
				value={value}
				equals={equals}
				onSelect={onChange}
				open={open}
				onOpenChange={setOpen}
				trigger={trigger}
				triggerProps={triggerProps}
				tooltipId={tooltipId}
				side={side}
				positionMethod={positionMethod}
			/>
		) : (
			<Select.Root<number>
				value={selectedIndex === -1 ? null : selectedIndex}
				onValueChange={index => index !== null && select(index)}
				open={open}
				onOpenChange={setOpen}
				disabled={disabled}
				modal={false}>
				<Select.Trigger {...triggerProps}>
					<Select.Value className="flex min-w-0 items-center overflow-hidden">{trigger}</Select.Value>
				</Select.Trigger>
				{/* Out of any Field the trigger sits in, so the popup's contents take no size from that field. */}
				<FieldContext.Provider value={null}>
					<Select.Portal container={portalContainer ?? undefined}>
						<Select.Positioner
							alignItemWithTrigger={false}
							align="start"
							side={side}
							positionMethod={positionMethod}
							sideOffset={2}
							className="ui-menu-positioner"
							data-testid="choice-field-positioner">
							<Select.Popup className="ui-popup-surface ui-menu">
								<Select.List className="ui-choice-field-list">
									{options.map((entry, position) =>
										isChoiceSection(entry) ? (
											<Select.Group key={`section-${position}`}>
												<Select.GroupLabel className="ui-choice-field-section">{entry.label}</Select.GroupLabel>
												{entry.options.map(selectItem)}
											</Select.Group>
										) : (
											isChoiceOption(entry) && selectItem(entry)
										),
									)}
								</Select.List>
							</Select.Popup>
						</Select.Positioner>
					</Select.Portal>
				</FieldContext.Provider>
			</Select.Root>
		);

	// The options exist only while the list is open, and so does the one tooltip they share.
	const tooltip = open && flat.some(option => option.tooltip !== undefined) && (
		<Tooltip
			id={tooltipId}
			maxWidth="max-w-tooltip-dropdown"
			align="start"
			render={({ activeAnchor }) => {
				const content = activeAnchor?.getAttribute('data-tooltip-content');
				return content ? <LocaleHtml html={content} /> : null;
			}}
		/>
	);

	if (!rooted)
		return (
			<>
				{control}
				{tooltip}
			</>
		);

	return (
		<Field {...frameProps} className={className} testId={testId} disabled={disabled}>
			{control}
			{tooltip}
		</Field>
	);
};
