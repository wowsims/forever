import type { Player } from '@sim/player/player';
import { render } from '@testing-library/react';
import { toastManager } from '@ui-kit/Toast';
import { afterEach, describe, expect, it, vi } from 'vitest';

import { makeSpecChangeWarningToast } from './spec_change_warning_toast';

const player = {} as Player<any>;

afterEach(() => {
	vi.restoreAllMocks();
});

describe('makeSpecChangeWarningToast', () => {
	it('adds no toast when no check matches', () => {
		const add = vi.spyOn(toastManager, 'add').mockReturnValue('toast');

		makeSpecChangeWarningToast([{ condition: () => false, message: 'never' }], player);

		expect(add).not.toHaveBeenCalled();
	});

	it('adds one warning listing every matching message, five seconds per message', () => {
		const add = vi.spyOn(toastManager, 'add').mockReturnValue('toast');

		makeSpecChangeWarningToast(
			[
				{ condition: () => true, message: 'Check your gear.' },
				{ condition: () => false, message: 'Not shown.' },
				{ condition: () => true, message: 'Check your talents.' },
			],
			player,
		);

		expect(add).toHaveBeenCalledTimes(1);
		const options = add.mock.calls[0][0];
		expect(options.variant).toBe('warning');
		expect(options.delay).toBe(10000);
		const { container } = render(<>{options.body}</>);
		expect([...container.querySelectorAll('p')].map(p => p.textContent)).toEqual(['Check your gear.', 'Check your talents.']);
	});
});
