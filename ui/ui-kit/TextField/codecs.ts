import { formatToNumber } from '@sim/utils/format';

import type { NumberCodecOptions, TextCodec } from './types';

export const numberCodec = ({ float = false, positive = false, showZeroes = true, maxDecimalDigits = 2 }: NumberCodecOptions = {}): TextCodec<number> => {
	const formatFloat = (value: number) => formatToNumber(value, { useGrouping: false, minimumFractionDigits: 2, maximumFractionDigits: maxDecimalDigits });
	return {
		format: value => {
			if (value === 0 && !showZeroes) return '';
			return float ? formatFloat(value) : String(value);
		},
		parse: text => {
			const entered = positive ? (float ? formatFloat(Math.abs(Number(text))) : Math.abs(parseInt(text)).toString()) : text;
			return float ? Number(entered || '') || 0 : parseInt(entered || '') || 0;
		},
	};
};

export const stringCodec: TextCodec<string> = {
	format: value => value,
	parse: text => text,
};

/** Comma-separated floats; an entry that does not parse is dropped. */
export const numberListCodec: TextCodec<Array<number>> = {
	format: value => value.map(entry => String(entry)).join(','),
	parse: text => {
		if (!text) return [];
		return text
			.split(',')
			.map(parseFloat)
			.filter(entry => !isNaN(entry));
	},
};
