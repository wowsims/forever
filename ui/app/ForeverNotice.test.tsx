import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ForeverNotice } from './ForeverNotice';
import type { SimHostObject } from './individual_sim_ui';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));
const trackEvent = vi.hoisted(() => vi.fn());
vi.mock('../tracking/utils', () => ({ trackEvent }));

const restoreButton = () => screen.getByRole('button', { name: 'info.options.restore_defaults.button' });

describe('ForeverNotice', () => {
	it('restores defaults the way the settings dialog does', () => {
		const host = { applyDefaults: vi.fn() } as unknown as SimHostObject<any>;
		render(<ForeverNotice host={host} />);

		expect(screen.getByText('info.forever_notice')).toBeTruthy();
		fireEvent.click(restoreButton());

		expect(host.applyDefaults).toHaveBeenCalledTimes(1);
		expect(trackEvent).toHaveBeenCalledWith({ action: 'settings', category: 'restore-defaults', label: 'restore' });
		expect(toasts).toEqual([{ variant: 'success', body: 'info.options.restore_defaults.success_message' }]);
	});

	it('disables the button until the host exists', () => {
		render(<ForeverNotice host={null} />);

		expect(restoreButton().hasAttribute('disabled') || restoreButton().getAttribute('aria-disabled') === 'true').toBe(true);
	});
});
