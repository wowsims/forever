import i18n from '@i18n/config';
import { toastManager } from '@ui-kit/Toast';

/** A toast that says the copy did not reach the clipboard and holds `text` for a manual copy. */
export const showForManualCopy = (text: string) =>
	toastManager.add({
		variant: 'error',
		autohide: false,
		title: i18n.t('common.copy_button.manual_copy'),
		body: <pre className="max-h-40 overflow-auto break-all whitespace-pre-wrap select-all">{text}</pre>,
	});

/**
 * Writes `text` to the clipboard. With no `navigator.clipboard` or `writeText` (an insecure origin, an older
 * browser) or a refused write, a toast holds the text for a manual copy. Resolves whether the write landed.
 */
export const writeToClipboard = async (text: string): Promise<boolean> => {
	try {
		await navigator.clipboard.writeText(text);
		return true;
	} catch {
		showForManualCopy(text);
		return false;
	}
};
