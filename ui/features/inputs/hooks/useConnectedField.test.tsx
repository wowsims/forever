import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useConnectedField } from './useConnectedField';

const local = { on: false };

const config = {
	id: 'my-field',
	label: 'My Field',
	labelTooltip: 'a tooltip',
	description: 'a description',
	layout: 'inline' as const,
	extraClassNames: ['a', 'b'],
	getValue: (m: typeof local) => m.on,
	setValue: (m: typeof local, next: boolean) => {
		m.on = next;
	},
};

const Probe = () => {
	const { frameProps } = useConnectedField(local, config);
	return <span>{JSON.stringify(frameProps)}</span>;
};

describe('useConnectedField', () => {
	it('maps every frame field off the config', () => {
		render(<Probe />);
		expect(JSON.parse(screen.getByText(/./).textContent!)).toEqual({
			id: 'my-field',
			label: 'My Field',
			labelTooltip: 'a tooltip',
			description: 'a description',
			layout: 'inline',
			className: ['a', 'b'],
		});
	});
});
