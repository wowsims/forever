import { useId } from 'react';

import { After, Clear, Control, Description, Label, Leading, Root } from './FieldParts';
import type { FieldProps } from './types';

const FieldSugar = ({
	id,
	label,
	labelTooltip,
	labelTooltipPlace,
	description,
	layout,
	size,
	className,
	testId,
	disabled,
	grow,
	clear,
	leading,
	children,
	after,
	ref,
	role,
	'aria-labelledby': labelledBy,
	hidden,
	reverse,
	labelId: labelIdProp,
	controlRef,
}: FieldProps) => {
	const generatedLabelId = useId();
	const labelId = labelIdProp ?? generatedLabelId;
	// A group's own caption names it; without one there is nothing yet to point aria-labelledby at.
	const groupLabelledBy = role === 'group' && label ? labelId : undefined;

	return (
		<Root
			id={id}
			layout={layout}
			size={size}
			disabled={disabled}
			grow={grow}
			reverse={reverse}
			ref={ref}
			role={role}
			aria-labelledby={labelledBy ?? groupLabelledBy}
			hidden={hidden}
			testId={testId}
			className={className}
			labelId={labelId}
			controlRef={controlRef}>
			{leading}
			{label && (
				<Label tooltip={labelTooltip} tooltipPlace={labelTooltipPlace}>
					{label}
				</Label>
			)}
			{description && <Description>{description}</Description>}
			{clear ? <Clear {...clear}>{children}</Clear> : children}
			{after}
		</Root>
	);
};

FieldSugar.displayName = 'Field';

/** A labelled field in one component, and the namespace of the parts it is built from. */
export const Field = Object.assign(FieldSugar, { Root, Label, Control, Description, Leading, After, Clear });
