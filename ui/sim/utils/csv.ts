export type CsvValue = string | number;

// Quotes a field only when it holds a quote, comma or line break, doubling the quotes inside.
const csvField = (value: CsvValue) => {
	const text = String(value);
	return /[",\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
};

/**
 * A name someone else wrote, as a cell a spreadsheet reads as text: one opening with any of these runs as a formula, and
 * the leading apostrophe stops that. Applied per column, because cells the sim writes (a talents string with an empty
 * first tree) open with a dash and have to stay as they are.
 */
export const csvText = (name: string): string => (/^[=+\-@\t\r]/.test(name) ? `'${name}` : name);

/** The rows as CSV text, one line each and a closing line break. */
export const toCsv = (rows: ReadonlyArray<ReadonlyArray<CsvValue>>): string => rows.map(row => row.map(csvField).join(',')).join('\n') + '\n';
