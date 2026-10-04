import { Toolbar as BaseToolbar } from '@base-ui/react/toolbar';
import type { ButtonProps } from '@ui-kit/Button';
import { Button } from '@ui-kit/Button';

export type ToolbarButtonProps = ButtonProps & {
	testId?: string;
};

export const ToolbarButton = ({ testId, ...buttonProps }: ToolbarButtonProps) => (
	<BaseToolbar.Button disabled={'disabled' in buttonProps ? buttonProps.disabled : undefined} render={<Button {...buttonProps} />} data-testid={testId} />
);
