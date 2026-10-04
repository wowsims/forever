import type { ResolvedIcon } from '@ui-kit/Field/types';
import type { CSSProperties } from 'react';

export const resolvedIconStyle = (icon: ResolvedIcon | undefined): CSSProperties | undefined =>
	icon ? { backgroundImage: `url('${icon.iconUrl}')` } : undefined;
