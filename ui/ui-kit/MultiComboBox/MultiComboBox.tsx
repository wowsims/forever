import { Combobox } from '@base-ui/react/combobox';
import { Chip } from '@ui-kit/Chip';
import { Field, useControlSize } from '@ui-kit/Field';
import { FieldContext } from '@ui-kit/Field/context/FieldContext';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import { Icon } from '@ui-kit/Icon';
import { useMemo, useRef } from 'react';

import type { EnumValueConfig, MultiComboBoxProps } from './types';

// base-ui's multi-select combobox, with the picks as chips below the field.
export const MultiComboBox = ({
	value,
	onChange,
	disabled,
	values,
	placeholder,
	emptyText,
	removeLabel,
	openLabel,
	testId,
	fitOptions,
	...frameProps
}: MultiComboBoxProps) => {
	const portalContainer = usePortalContainer();
	const anchor = useRef<HTMLDivElement>(null);
	const size = useControlSize(frameProps.size);

	const selected = useMemo(() => values.filter(entry => value.includes(entry.value)), [values, value]);

	return (
		<Field
			{...frameProps}
			className={[fitOptions && 'ui-multi-combo-box-fit', frameProps.className]}
			disabled={disabled}
			testId={testId ?? 'multi-combo-box-root'}>
			<Combobox.Root
				multiple
				items={values}
				value={selected}
				onValueChange={(next: Array<EnumValueConfig>) => onChange(next.map(entry => entry.value))}
				itemToStringLabel={(entry: EnumValueConfig) => entry.name}
				isItemEqualToValue={(a: EnumValueConfig, b: EnumValueConfig) => a.value === b.value}
				disabled={disabled}>
				<Combobox.InputGroup ref={anchor} className="ui-multi-combo-box-field" data-size={size} data-testid="multi-combo-box-field">
					<Combobox.Input id={frameProps.id} className="ui-multi-combo-box-input" placeholder={placeholder} data-testid="multi-combo-box-input" />
					{fitOptions && (
						<span className="ui-multi-combo-box-sizer" aria-hidden>
							{[placeholder, ...values.map(entry => entry.name)].map((text, index) => (
								<span key={index} data-label={text} />
							))}
						</span>
					)}
					<Combobox.Trigger className="ui-multi-combo-box-trigger" aria-label={openLabel} data-testid="multi-combo-box-trigger">
						<Icon name="chevron-down" />
					</Combobox.Trigger>
				</Combobox.InputGroup>
				<Combobox.Chips className="ui-multi-combo-box-chips" data-disabled={disabled ? '' : undefined} data-testid="multi-combo-box-chips">
					<Combobox.Value>
						{(picks: Array<EnumValueConfig>) =>
							picks.map(entry => (
								<Combobox.Chip
									key={entry.value}
									aria-label={entry.name}
									render={props => (
										<Chip
											label={entry.name}
											nameAs="span"
											// Base UI marks a disabled chip data-disabled, which hides a kit Chip; the picks stay in view.
											rootProps={{ ...props, 'data-disabled': undefined }}
											confirmDelete={false}
											deleteLabel={removeLabel}
											onDelete={disabled ? undefined : () => onChange(value.filter(picked => picked !== entry.value))}
											testId="multi-combo-box-chip"
										/>
									)}
								/>
							))
						}
					</Combobox.Value>
				</Combobox.Chips>
				{/* Out of the box's own Field, so the popup's contents take no size from it. */}
				<FieldContext.Provider value={null}>
					<Combobox.Portal container={portalContainer ?? undefined}>
						<Combobox.Positioner className="ui-multi-combo-box-positioner" anchor={anchor} align="start" sideOffset={2}>
							<Combobox.Popup className="ui-multi-combo-box-popup">
								<Combobox.Empty className="ui-multi-combo-box-empty">{emptyText}</Combobox.Empty>
								<Combobox.List className="ui-multi-combo-box-list" data-testid="multi-combo-box-list">
									{(entry: EnumValueConfig) => (
										<Combobox.Item key={entry.value} value={entry} className="ui-multi-combo-box-item" title={entry.tooltip}>
											<span>{entry.name}</span>
											<Combobox.ItemIndicator className="ui-multi-combo-box-indicator">
												<Icon name="check" />
											</Combobox.ItemIndicator>
										</Combobox.Item>
									)}
								</Combobox.List>
							</Combobox.Popup>
						</Combobox.Positioner>
					</Combobox.Portal>
				</FieldContext.Provider>
			</Combobox.Root>
		</Field>
	);
};
