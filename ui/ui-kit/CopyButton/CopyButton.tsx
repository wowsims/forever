import i18n from '@i18n/config';
import { Button, type ButtonSize, type ButtonVariant } from '@ui-kit/Button';
import { useCopyToClipboard } from '@ui-kit/hooks/useCopyToClipboard';
import { Icon } from '@ui-kit/Icon';
import { toastManager } from '@ui-kit/Toast';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx, { type ClassValue } from 'clsx';
import { useId } from 'react';

export interface CopyButtonProps {
	/** Read at click time, not at mount: the reforge toast re-exports the whole sim on every click. */
	getContent: () => string;
	variant?: ButtonVariant | null;
	size?: ButtonSize;
	className?: ClassValue;
	text?: string;
	tooltip?: string;
	testId?: string;
	disabled?: boolean;
}

/**
 * Parameterises the label, the look and the tooltip; fixes the `copy-button` testid and the copy
 * itself: the check icon and the "copied" label for 1500ms, further clicks ignored inside it.
 */
export const CopyButton = ({ getContent, variant = null, size, className, text, tooltip, testId = 'copy-button', disabled }: CopyButtonProps) => {
	const { copy, copied } = useCopyToClipboard(getContent);
	const tooltipId = useId();

	const onClick = () => {
		// `navigator.clipboard` is absent on an insecure origin: this toasts the payload for a manual copy instead and never enters the copied window, so a second click toasts again.
		if (navigator.clipboard == undefined) {
			toastManager.add({
				variant: 'error',
				autohide: false,
				body: <pre className="max-h-40 overflow-auto break-all whitespace-pre-wrap select-all">{getContent()}</pre>,
			});
		} else {
			if (copied) return;
			copy();
		}
	};

	return (
		<>
			<Button
				variant={variant}
				size={size}
				className={clsx(className)}
				data-testid={testId}
				disabled={disabled}
				onClick={onClick}
				{...(tooltip ? tooltipAnchorProps(tooltipId) : {})}>
				<Icon name={copied ? 'check' : 'copy'} />
				{copied ? i18n.t('common.copy_button.copied') : (text ?? i18n.t('common.copy_button.default_text'))}
			</Button>
			{tooltip && <Tooltip id={tooltipId} content={tooltip} />}
		</>
	);
};
