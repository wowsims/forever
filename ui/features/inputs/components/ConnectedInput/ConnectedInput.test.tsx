import { renderWithHost } from '@testing/react';
import { noopSubscribe } from '@testing/sim/fake_host';
import { fireEvent, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ConnectedInput } from './ConnectedInput';
import type { ConnectedInputConfig } from './types';

const player = { delay: 3, stance: 1, active: false };

const shared = { storeSubscribe: () => noopSubscribe };

const configs = {
	number: {
		type: 'number',
		id: 'cast-delay',
		label: 'Cast Delay',
		...shared,
		getValue: (p: typeof player) => p.delay,
		setValue: (p: typeof player, value: number) => {
			p.delay = value;
		},
	},
	boolean: {
		type: 'boolean',
		id: 'use-cooldowns',
		label: 'Use Cooldowns',
		...shared,
		getValue: (p: typeof player) => p.active,
		setValue: (p: typeof player, value: boolean) => {
			p.active = value;
		},
	},
	enum: {
		type: 'enum',
		id: 'stance',
		label: 'Stance',
		values: [
			{ name: 'Battle', value: 0 },
			{ name: 'Defensive', value: 1 },
		],
		...shared,
		getValue: (p: typeof player) => p.stance,
		setValue: (p: typeof player, value: number) => {
			p.stance = value;
		},
	},
} as unknown as Record<'number' | 'boolean' | 'enum', ConnectedInputConfig<any>>;

const mount = (config: ConnectedInputConfig<any>, inline?: boolean) =>
	renderWithHost(<ConnectedInput config={config} inline={inline} />, { player }).container.firstElementChild!;

describe('ConnectedInput', () => {
	it('renders a number input as a number picker bound to the player', () => {
		const root = mount(configs.number);
		expect(root.getAttribute('data-testid')).toBe('number-picker-root');
		expect(screen.getByRole<HTMLInputElement>('textbox').value).toBe('3');
	});

	it('renders a boolean input as a checkbox after its label, bound to the player', () => {
		const root = mount(configs.boolean);
		expect(root.getAttribute('data-testid')).toBe('boolean-picker-root');
		expect(root.hasAttribute('data-reverse')).toBe(true);

		fireEvent.click(screen.getByRole('checkbox'));
		expect(player.active).toBe(true);
	});

	it('renders an enum input as a choice field bound to the player', () => {
		const root = mount(configs.enum);
		expect(root.contains(screen.getByRole('combobox', { name: 'Stance' }))).toBe(true);
		expect(screen.getByRole('combobox').textContent).toBe('Defensive');
	});

	it('lays the field out inline when asked, and keeps the config layout otherwise', () => {
		expect(mount(configs.number, true).getAttribute('data-layout')).toBe('inline');
		expect(mount({ ...configs.enum, layout: 'split' } as ConnectedInputConfig<any>).getAttribute('data-layout')).toBe('split');
	});
});
