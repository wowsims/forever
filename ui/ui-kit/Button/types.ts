import type { AnchorHTMLAttributes, ButtonHTMLAttributes, LabelHTMLAttributes, ReactNode, Ref } from 'react';

export type ButtonVariant =
	| 'primary'
	| 'secondary'
	| 'danger'
	| 'cancel'
	| 'link'
	| 'link-danger'
	| 'outline-primary'
	| 'outline-secondary'
	| 'outline-light'
	| 'outline-cancel'
	| 'unstyled';

export type ButtonSize = 'sm' | 'xs' | 'none';

interface ButtonCommonProps {
	/** `null` emits the bare base bundle — no colour of its own. */
	variant?: ButtonVariant | null;
	size?: ButtonSize;
	/** Renders a small spinner before the children, or over them for `iconOnly`, and sets `aria-busy`; it does not disable the button. */
	loading?: boolean;
	className?: string;
	children?: ReactNode;
}

type ButtonIconProps =
	| {
			/** No padding, border or background of its own — an icon-sized hit target. */
			iconOnly: true;
			'aria-label': string;
	  }
	| { iconOnly?: false; 'aria-label'?: string };

export type ButtonBaseProps = ButtonCommonProps & ButtonIconProps;

export type ButtonAsButton = ButtonBaseProps &
	Omit<ButtonHTMLAttributes<HTMLButtonElement>, 'className' | 'children' | 'aria-label'> & {
		as?: 'button';
		href?: never;
		ref?: Ref<HTMLButtonElement>;
		/** While disabled, sets `aria-disabled` instead of `disabled`, so the button still takes focus and hover and a tooltip on it can say why. */
		focusableWhenDisabled?: boolean;
	};

/** `href` renders an anchor; `rel` gains `noopener noreferrer` when the href leaves this origin. */
export type ButtonWithHref = ButtonBaseProps &
	Omit<AnchorHTMLAttributes<HTMLAnchorElement>, 'className' | 'children' | 'aria-label'> & { href: string; as?: never };

export type ButtonAsLabel = ButtonBaseProps &
	Omit<LabelHTMLAttributes<HTMLLabelElement>, 'className' | 'children' | 'aria-label'> & { as: 'label'; htmlFor: string; href?: never };

export type ButtonProps = ButtonAsButton | ButtonWithHref | ButtonAsLabel;
