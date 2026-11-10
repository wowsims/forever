import type { AplNameDialogProps } from '@features/apl/components/AplNameDialog';
import type { AplPasteDialogProps } from '@features/apl/components/AplPasteDialog';
import { freeNameDialog } from '@features/apl/hooks/useRenamedCopy';
import { type AplClip, aplClipFromText, type AplClipItems, type AplClipKind, aplClipToText } from '@features/apl/model/clipboard';
import i18n from '@i18n/config';
import { toastManager } from '@ui-kit/Toast';
import { writeToClipboard } from '@ui-kit/utils/clipboard';
import { useState } from 'react';

const LABELS: Record<AplClipKind, { item: string; places: string; name?: string }> = {
	action: { item: 'rotation_tab.apl.priorityList.name', places: 'rotation_tab.apl.clipboard.places.action' },
	group: {
		item: 'rotation_tab.apl.actionGroups.name',
		places: 'rotation_tab.apl.tabs.actionGroups',
		name: 'rotation_tab.apl.actionGroups.attributes.name',
	},
	variable: { item: 'rotation_tab.apl.variables.name', places: 'rotation_tab.apl.tabs.variables', name: 'rotation_tab.apl.variables.attributes.name' },
};

/** Writes the item to the clipboard and toasts; a write that cannot land leaves the text in a toast for a manual copy. */
export const copyAplClip = (clip: AplClip) => {
	if (!clip.item) return;
	void writeToClipboard(aplClipToText(clip)).then(written => {
		if (written) toastManager.add({ variant: 'success', body: i18n.t('rotation_tab.apl.clipboard.copied', { itemName: i18n.t(LABELS[clip.kind].item) }) });
	});
};

/** `undefined` when the page cannot read the clipboard, whether `navigator.clipboard` or `readText` is missing or the read is refused. */
const readClipboardText = async (): Promise<string | undefined> => {
	try {
		return await navigator.clipboard.readText();
	} catch {
		return undefined;
	}
};

/** The `useListItems` binding of the list a paste appends to. */
interface AplPasteTarget<T> {
	items: ReadonlyArray<T>;
	append: (item: T) => void;
}

export interface AplPaste {
	/** For the paste button: reads the clipboard, or opens `textDialog` when the page cannot. */
	paste: () => void;
	textDialog: AplPasteDialogProps;
	/** Opens for a pasted group or variable, filled with the next free name; an action never opens it. */
	nameDialog: AplNameDialogProps;
}

/**
 * Pasting a copied item at the end of a list. A group or a variable is referenced by name, so it goes through
 * the name dialog first, as Copy to New does, and the dialog checks the name against the list as it stands then.
 */
export const useAplPaste = <K extends AplClipKind>(kind: K, list: AplPasteTarget<AplClipItems[K]>): AplPaste => {
	const [typing, setTyping] = useState(false);
	const [naming, setNaming] = useState<AplClipItems['group' | 'variable'] | null>(null);
	const pasteLabel = i18n.t('rotation_tab.apl.floatingActionBar.paste', { itemName: i18n.t(LABELS[kind].item) });

	const land = (text: string) => {
		const clip = aplClipFromText(text);
		if (!clip) {
			toastManager.add({ variant: 'error', body: i18n.t('rotation_tab.apl.clipboard.empty', { itemName: i18n.t(LABELS[kind].item) }) });
		} else if (clip.kind !== kind) {
			toastManager.add({
				variant: 'warning',
				body: i18n.t('rotation_tab.apl.clipboard.wrong_kind', { held: i18n.t(LABELS[clip.kind].item), places: i18n.t(LABELS[clip.kind].places) }),
			});
		} else if (clip.kind === 'action') {
			list.append(clip.item as AplClipItems[K]);
		} else {
			setNaming(clip.item);
		}
	};

	return {
		paste: () => void readClipboardText().then(text => (text === undefined ? setTyping(true) : land(text))),
		textDialog: { open: typing, title: pasteLabel, onSubmit: land, onClose: () => setTyping(false) },
		nameDialog: freeNameDialog({
			title: pasteLabel,
			inputLabel: LABELS[kind].name ? i18n.t(LABELS[kind].name) : '',
			confirmLabel: pasteLabel,
			name: naming?.name,
			taken: naming ? list.items.map(item => (item as AplClipItems['group' | 'variable']).name) : [],
			onSubmit: name => {
				if (naming) list.append({ ...naming, name } as AplClipItems[K]);
			},
			onClose: () => setNaming(null),
		}),
	};
};
