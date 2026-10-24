import type { IndividualSimHost } from '@sim/sim_host';

export interface ImporterDefinition {
	title: string;
	allowFileUpload?: boolean;
	/**
	 * Throw to report a problem; the shell shows it as the error toast, or an `ImportProblems` as a list under the
	 * text. `fileName` is the name of the file the text was uploaded from, and none for pasted text.
	 */
	onImport: (host: IndividualSimHost<any>, data: string, fileName?: string) => Promise<void>;
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
