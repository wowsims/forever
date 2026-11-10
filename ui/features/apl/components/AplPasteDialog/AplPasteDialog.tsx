import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Dialog } from '@ui-kit/Dialog';
import { Field } from '@ui-kit/Field';
import { TextArea } from '@ui-kit/FormControl';
import { useEffect, useId, useRef, useState } from 'react';

export interface AplPasteDialogProps {
	open: boolean;
	/** The paste button's own label, e.g. "Paste Action". */
	title: string;
	onSubmit: (text: string) => void;
	onClose: () => void;
}

/**
 * Where a paste goes when the page cannot read the clipboard: an insecure origin, a browser with no
 * `readText`, or a refused read. The user pastes the copied text into the box with the keyboard.
 */
export const AplPasteDialog = ({ open, title, onSubmit, onClose }: AplPasteDialogProps) => {
	const inputId = useId();
	const inputRef = useRef<HTMLTextAreaElement>(null);
	const [text, setText] = useState('');

	// Each opening starts empty; the dialog focuses the box (`initialFocus`).
	useEffect(() => {
		if (open) setText('');
	}, [open]);

	const submit = () => {
		if (!text.trim()) return;
		onSubmit(text);
		onClose();
	};

	return (
		<Dialog
			open={open}
			onOpenChange={next => {
				if (!next) onClose();
			}}
			initialFocus={inputRef}
			size="sm"
			title={title}
			footer={
				<Button variant="primary" disabled={!text.trim()} onClick={submit}>
					{title}
				</Button>
			}>
			<Field
				id={inputId}
				label={i18n.t('rotation_tab.apl.clipboard.paste_dialog.label')}
				description={i18n.t('rotation_tab.apl.clipboard.paste_dialog.description')}>
				<TextArea id={inputId} ref={inputRef} rows={6} value={text} onChange={event => setText(event.target.value)} />
			</Field>
		</Dialog>
	);
};
