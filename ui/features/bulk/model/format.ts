import { formatToNumber } from '@sim/utils/format';

// Always two decimals, so "776.30" lines up with "797.41" in a column of results.
export const formatBulkDps = (dps: number): string => formatToNumber(dps, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
