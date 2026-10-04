import { useState } from 'react';

/** A text control's uncommitted edit. Keyed on the formatted string, so a caller whose value is a fresh object per render does not wipe the draft per keystroke. */
export const useDraft = (display: string) => {
	const [draft, setDraft] = useState<string | null>(null);
	const [seen, setSeen] = useState(display);
	if (seen !== display) {
		setSeen(display);
		setDraft(null);
	}
	return { text: draft ?? display, setDraft, clear: () => setDraft(null) };
};
