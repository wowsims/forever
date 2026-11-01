import type { IndividualSimHost } from '@sim/sim_host';
import type { ReactNode } from 'react';

interface ImporterOptions {
	title: string;
	allowFileUpload?: boolean;
	/** The text is a link, which a plain box takes; a data file gets the code box. */
	holdsLink?: boolean;
	/** A bigger upload is refused before it is read, and the text goes with it. */
	maxFileBytes?: number;
}

export interface ImporterDefinition extends ImporterOptions {
	/**
	 * Throw to report a problem; the shell shows it as the error toast, or an `ImportProblems` as a list under the
	 * text. `fileName` is the name of the file the text was uploaded from, and none for pasted text. `signal` is
	 * aborted when the dialog closes before the import ends, after which an import that waits on a read loads nothing.
	 */
	onImport: (host: IndividualSimHost<any>, data: string, fileName?: string, signal?: AbortSignal) => Promise<void>;
	preview?: undefined;
}

/** A step between the text and the import: the dialog shows what the text holds, and Import acts on what it shows. */
export interface ImporterPreview<P> {
	/**
	 * Reads the text whenever it changes, so the identity has to hold between renders. Throw to refuse the text, which the
	 * shell reports as it does for `onImport`. `signal` is aborted once the text moves on or the dialog closes.
	 */
	parse: (host: IndividualSimHost<any>, data: string, signal: AbortSignal) => Promise<P>;
	/** Shown under the text. `change` replaces the preview, for one the user can edit. */
	render: (preview: P, change: (next: P) => void) => ReactNode;
	importLabel: (preview: P) => string;
	canImport: (preview: P) => boolean;
	/** The import, of the preview as it stands. It reports problems as `onImport` does. */
	commit: (preview: P) => Promise<void>;
}

export interface PreviewImporterDefinition<P> extends ImporterOptions {
	preview: ImporterPreview<P>;
	onImport?: undefined;
}

/** A refusal for problems at places in the text, one line each, which the dialog lists under the text they point into. */
export class ImportProblems extends Error {
	readonly problems: string[];

	constructor(message: string, problems: string[]) {
		super(message);
		this.name = 'ImportProblems';
		this.problems = problems;
	}
}
