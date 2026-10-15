export type CsvValue = string | number;

// Quotes a field only when it holds a quote, comma or line break, doubling the quotes inside.
const csvField = (value: CsvValue) => {
	const text = String(value);
	return /[",\r\n]/.test(text) ? `"${text.replaceAll('"', '""')}"` : text;
};

/** The rows as CSV text, one line each and a closing line break. */
export const toCsv = (rows: ReadonlyArray<ReadonlyArray<CsvValue>>): string => rows.map(row => row.map(csvField).join(',')).join('\n') + '\n';
