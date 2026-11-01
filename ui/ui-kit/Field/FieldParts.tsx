import { Field as BaseField } from '@base-ui/react/field';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { LocaleHtml, Tooltip, tooltipAnchorProps, type TooltipRefProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { isValidElement, type ReactNode, useCallback, useContext, useId, useLayoutEffect, useMemo, useRef, useState } from 'react';

import { FieldContext } from './context/FieldContext';
import { useField } from './hooks/useField';
import type { FieldClearPartProps, FieldControlProps, FieldDescriptionProps, FieldLabelProps, FieldRootProps, FieldSlotProps } from './types';

const dedupe = (classes: string) => Array.from(new Set(classes.split(' '))).join(' ');

const NATIVE_CONTROLS = new Set(['input', 'select', 'textarea']);

/**
 * Below a `28rem` wide container the field tightens: `Field.css` queries the nearest ancestor
 * container (a tab pane, a dialog body, the sidebar or a `Tree.Root`), since a container query
 * cannot style the container itself.
 */
export const Root = ({
	id,
	layout,
	size: ownSize,
	disabled,
	grow,
	className,
	testId,
	ref,
	role,
	'aria-labelledby': labelledBy,
	hidden,
	reverse,
	labelId: labelIdProp,
	controlRef: controlRefProp,
	children,
}: FieldRootProps) => {
	const [descriptionIds, setDescriptionIds] = useState<ReadonlyArray<string>>([]);
	const registerDescription = useCallback((descriptionId: string) => {
		setDescriptionIds(ids => [...ids, descriptionId]);
		return () => setDescriptionIds(ids => ids.filter(other => other !== descriptionId));
	}, []);
	const describedBy = descriptionIds.length > 0 ? descriptionIds.join(' ') : undefined;
	const generatedLabelId = useId();
	const labelId = labelIdProp ?? generatedLabelId;
	const generatedControlRef = useRef<HTMLElement | null>(null);
	const controlRef = controlRefProp ?? generatedControlRef;
	const labelTooltip = useRef<{ open: () => void; close: () => void } | null>(null);
	const parentSize = useContext(FieldContext)?.size;
	const size = ownSize ?? parentSize;
	const context = useMemo(
		() => ({ id, role, layout, size, disabled: !!disabled, grow: grow !== false, describedBy, registerDescription, labelId, controlRef, labelTooltip }),
		[id, role, layout, size, disabled, grow, describedBy, registerDescription, labelId, controlRef],
	);

	return (
		<FieldContext.Provider value={context}>
			<BaseField.Root
				ref={ref}
				role={role}
				aria-labelledby={labelledBy}
				hidden={hidden}
				disabled={disabled}
				data-disabled={disabled ? '' : undefined}
				data-layout={layout}
				data-size={size}
				data-reverse={reverse ? '' : undefined}
				data-testid={testId ?? 'input-root'}
				className={dedupe(clsx(describedBy && 'flex-wrap', 'ui-field', grow === false && 'flex-none', className))}>
				{children}
			</BaseField.Root>
		</FieldContext.Provider>
	);
};

export const Label = ({ as, id, tooltip, tooltipPlace, className, testId, children }: FieldLabelProps) => {
	const field = useField();
	const renderable = typeof tooltip === 'string' || isValidElement(tooltip);
	if (tooltip !== undefined && !renderable) {
		console.warn(`Field "${field.id ?? children}": tooltip is neither a string nor an element, so it is not rendered.`, tooltip);
	}
	const generatedId = useId();
	const tooltipId = renderable ? `${field.id ?? generatedId}-tooltip` : undefined;
	const tooltipRef = useRef<TooltipRefProps>(null);
	// A string goes through LocaleHtml, as ContentBlock's header tooltip already does: several spec
	// tooltips carry <b> to mark the option they are talking about, and passed raw those render as
	// literal tags.
	const tooltipNode = useMemo(() => {
		if (!tooltipId) return null;
		const content = typeof tooltip === 'string' ? <LocaleHtml html={tooltip} /> : (tooltip as ReactNode);
		return <Tooltip ref={tooltipRef} id={tooltipId} content={content} place={tooltipPlace} />;
	}, [tooltipId, tooltip, tooltipPlace]);
	// A control that opens the tooltip on its own focus opens it here, anchored at the label, so a side placement sits
	// beside the label text rather than over it.
	const { labelTooltip } = field;
	useLayoutEffect(() => {
		if (!tooltipId) return;
		const anchorSelect = `.ui-field-label[data-tooltip-id="${tooltipId}"]`;
		labelTooltip.current = { open: () => tooltipRef.current?.open({ anchorSelect }), close: () => tooltipRef.current?.close() };
		return () => {
			labelTooltip.current = null;
		};
	}, [labelTooltip, tooltipId]);

	// A group's caption and an icon field's label both name a control (or several) they cannot carry a
	// `for` to, so both fall back to `field.labelId`, which the field's `aria-labelledby` and control
	// share. An icon field's label also clicks the control it names, since it is the only thing a
	// wowhead anchor swatch can be labelled by.
	const spanned = !as && (field.role === 'group' || field.layout === 'icon');
	const tag = as ?? (spanned ? 'span' : 'label');
	const activates = spanned && field.layout === 'icon';
	const props = {
		id: id ?? (field.role === 'group' || field.layout === 'icon' ? field.labelId : undefined),
		className: clsx('ui-field-label', className),
		title: children,
		'data-testid': testId ?? 'form-label',
		...(activates ? { onClick: () => field.controlRef.current?.click(), 'data-activates': '' } : {}),
		...tooltipAnchorProps(tooltipId),
	};

	return (
		<>
			{tag === 'span' ? (
				<span {...props}>{children}</span>
			) : (
				// `htmlFor` explicitly rather than letting Base UI derive it.
				<BaseField.Label {...props} htmlFor={field.id}>
					{children}
				</BaseField.Label>
			)}
			{tooltipNode}
		</>
	);
};

export const Description = ({ id, className, testId, children }: FieldDescriptionProps) => {
	const { registerDescription } = useField();
	const generatedId = useId();
	const descriptionId = id ?? generatedId;
	useLayoutEffect(() => registerDescription(descriptionId), [registerDescription, descriptionId]);
	const props = {
		id: descriptionId,
		render: <div />,
		className: clsx('ui-field-description', className),
		'data-testid': testId ?? 'input-description',
	};

	return <BaseField.Description {...props}>{children}</BaseField.Description>;
};

/**
 * A native `input`, `select` or `textarea` becomes Base UI's `Field.Control`, which gives it the
 * root's id, its disabled state and the descriptions as `aria-describedby`. Anything else sits in a
 * wrapper that carries `aria-describedby` and no id, since a label cannot point at a `div`.
 */
export const Control = ({ className, testId, children }: FieldControlProps) => {
	const { id, describedBy } = useField();
	const classes = clsx(className) || undefined;
	if (typeof children.type === 'string' && NATIVE_CONTROLS.has(children.type)) {
		return <BaseField.Control render={children} id={id} className={classes} data-testid={testId} />;
	}
	return (
		<div className={classes} data-testid={testId ?? 'field-control'} aria-describedby={describedBy}>
			{children}
		</div>
	);
};

export const Leading = ({ className, testId, children }: FieldSlotProps) => {
	useField();
	return (
		<div className={clsx(className) || undefined} data-testid={testId ?? 'field-leading'}>
			{children}
		</div>
	);
};

export const After = ({ className, testId, children }: FieldSlotProps) => {
	useField();
	return (
		<div className={clsx(className) || undefined} data-testid={testId ?? 'field-after'}>
			{children}
		</div>
	);
};

/** Wraps its control with an absolute clear button; `show: false` keeps the wrapper and hides the button. */
export const Clear = ({ show, label, className, testId, onClear, children }: FieldClearPartProps) => {
	const { grow } = useField();
	return (
		<div className={clsx('relative flex', grow ? 'w-full' : 'flex-none')}>
			{children}
			{show && (
				<Button
					variant="link"
					className={clsx('absolute inset-y-0 right-0 flex items-center', className)}
					data-testid={testId}
					aria-label={label}
					onClick={onClear}>
					<Icon name="xmark" />
				</Button>
			)}
		</div>
	);
};

Root.displayName = 'Field.Root';
Label.displayName = 'Field.Label';
Description.displayName = 'Field.Description';
Control.displayName = 'Field.Control';
Leading.displayName = 'Field.Leading';
After.displayName = 'Field.After';
Clear.displayName = 'Field.Clear';
