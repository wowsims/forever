import { Unstyled } from '@storybook/addon-docs/blocks';
import { useSyncExternalStore } from 'react';

import { tokensNamed } from '../tokens/theme_tokens';

const subscribe = (onChange: () => void) => {
	window.addEventListener('resize', onChange);
	return () => window.removeEventListener('resize', onChange);
};

const breakpoints = tokensNamed('breakpoint').map(token => ({ name: token.name.slice('--breakpoint-'.length), minWidth: parseFloat(token.value) }));

export const ViewportReadout = () => {
	const width = useSyncExternalStore(subscribe, () => window.innerWidth);
	const active = breakpoints.filter(breakpoint => width >= breakpoint.minWidth).map(breakpoint => breakpoint.name);
	return (
		<Unstyled>
			<p className="mb-6 border border-border-muted p-2">
				Preview viewport: <strong>{width}px</strong>. Active: <code>{active.length ? active.join(' ') : 'none (below sm)'}</code>
			</p>
		</Unstyled>
	);
};
