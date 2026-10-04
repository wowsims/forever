import { Button as BaseButton } from '@base-ui/react/button';
import { externalRel } from '@sim/utils/links';
import { Spinner } from '@ui-kit/Spinner';
import clsx from 'clsx';

import type { ButtonProps, ButtonWithHref } from './types';

const SIZE_CLASS = {
	default: 'ui-button-md',
	sm: 'ui-button-sm',
	xs: 'ui-button-xs',
	none: undefined,
} as const;

const hasHref = (props: ButtonProps): props is ButtonWithHref => props.href !== undefined;

export const Button = (props: ButtonProps) => {
	const variant = props.variant === undefined ? (props.iconOnly ? null : 'primary') : props.variant;
	const isLink = variant === 'link' || variant === 'link-danger';
	const base = props.iconOnly ? 'ui-button-icon' : isLink ? 'ui-button-link' : 'ui-button';
	const classes =
		variant === 'unstyled'
			? props.className
			: clsx(
					base,
					// `ui-button-link` is the link colour; an iconOnly link keeps it on top of the frameless base.
					variant && (variant !== 'link' || props.iconOnly) && `ui-button-${variant}`,
					!props.iconOnly && SIZE_CLASS[props.size ?? 'default'],
					props.className,
				);
	const busy = props.loading || undefined;
	const content = !props.loading ? (
		props.children
	) : (
		<>
			<span className="invisible contents" data-hidden="">
				{props.children}
			</span>
			<Spinner size="xs" className="ui-button-spinner" />
		</>
	);

	if (props.as === 'label') {
		const {
			as: _as,
			variant: _variant,
			size: _size,
			iconOnly: _iconOnly,
			loading: _loading,
			className: _className,
			children: _children,
			...labelProps
		} = props;
		return (
			<label className={classes} aria-busy={busy} {...labelProps}>
				{content}
			</label>
		);
	}

	// A plain anchor, deliberately not `Base.Button`.
	if (hasHref(props)) {
		const {
			as: _as,
			variant: _variant,
			size: _size,
			iconOnly: _iconOnly,
			loading: _loading,
			className: _className,
			children: _children,
			...anchorProps
		} = props;
		return (
			<a className={classes} aria-busy={busy} {...anchorProps} rel={externalRel(anchorProps.href, anchorProps.rel)}>
				{content}
			</a>
		);
	}

	// `type` is defaulted here rather than by Base UI, which does not do it: a <button> inside a form submits it otherwise, and several in this tree are inside forms.
	const {
		as: _as,
		variant: _variant,
		size: _size,
		iconOnly: _iconOnly,
		loading: _loading,
		className: _className,
		children: _children,
		type = 'button',
		...buttonProps
	} = props;
	return (
		<BaseButton className={classes} type={type} aria-busy={busy} {...buttonProps}>
			{content}
		</BaseButton>
	);
};
