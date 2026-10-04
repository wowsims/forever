import type { CSSProperties } from 'react';

import { resolvedIconStyle } from '../utils/icons';
import type { IconEnumOption } from './types';

export const iconStyleOf = <T>(option: IconEnumOption<T>): CSSProperties | undefined => {
	if (option.icon) return resolvedIconStyle(option.icon);
	if (option.iconUrl) return { backgroundImage: `url(${option.iconUrl})`, filter: 'grayscale(1)' };
	return { backgroundColor: option.color };
};
