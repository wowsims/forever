import { LogLevel } from '@generated/proto/common';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { source } from '@testing/apl/source';
import { act, fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { ListItemHeader } from './ListItemHeader';
import { uuidValidations } from './utils';

vi.mock('@sim/state/subscriptions', async () => {
	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
});
vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));
vi.mock('@sim/proto/action_id', () => ({ ActionId: { replaceAllInString: vi.fn(async (str: string) => str) } }));

interface Row {
	hide: boolean;
}

const makePlayer = (touchRotation: () => void = () => {}) => ({
	touchRotation,
	getCurrentStats: () => ({ rotationStats: undefined }),
});

const mount = (item: Row, getValidations: (player: any) => Array<any>, player = makePlayer()) => {
	const rendered = render(
		<SimHostProvider host={{ player } as never}>
			<ListItemHeader player={player as any} getItem={() => item} getValidations={getValidations} />
		</SimHostProvider>,
	);
	return { ...rendered, player };
};

describe('ListItemHeader', () => {
	describe('hide picker', () => {
		it('marks the eye active while the item is hidden', () => {
			const notHidden = mount({ hide: false }, () => []);
			const hidden = mount({ hide: true }, () => []);

			expect(notHidden.container.querySelector('[data-testid="hide-picker-button"]')!.getAttribute('aria-pressed')).toBe('false');
			expect(hidden.container.querySelector('[data-testid="hide-picker-button"]')!.getAttribute('aria-pressed')).toBe('true');
		});

		it('writes hide and calls player.touchRotation() on click', () => {
			const touchRotation = vi.fn();
			const player = makePlayer(touchRotation);
			const item: Row = { hide: false };
			const { container } = mount(item, () => [], player);

			act(() => {
				fireEvent.click(container.querySelector('[data-testid="hide-picker-button"]') as HTMLButtonElement);
			});

			expect(item.hide).toBe(true);
			expect(touchRotation).toHaveBeenCalledTimes(1);
		});
	});

	describe('validations', () => {
		const validation = (logLevel: LogLevel, validation: string) => ({ logLevel, validation });

		// The sim formats rotation-supplied names into its messages ("Group reference '%s' not found"),
		// and rotations are shared and imported, so a message is attacker-controlled text.
		it('renders a message as text, so markup in a rotation-supplied name cannot execute', async () => {
			source.listeners.clear();
			const hostile = `Group reference '<img src=x onerror=alert(1)>' not found`;
			const { container } = mount({ hide: false }, () => [validation(LogLevel.Error, hostile)]);

			await act(async () => {
				source.notify();
			});
			await act(async () => {
				fireEvent.mouseEnter(container.querySelector('[data-testid="apl-validations"]') as HTMLElement);
			});

			const tooltip = document.querySelector('.sim-tooltip') as HTMLElement;
			expect(tooltip.querySelector('img')).toBeNull();
			expect(tooltip.querySelector('li')!.textContent).toBe(hostile);
		});
	});
});

describe('uuidValidations', () => {
	const makeStatsPlayer = (uuidValidationsList: Array<{ uuid: { value: string }; validations: Array<any> }>) => ({
		getCurrentStats: () => ({ rotationStats: { uuidValidations: uuidValidationsList } }),
	});

	it('returns [] for an undefined uuid', () => {
		const player = makeStatsPlayer([{ uuid: { value: 'a' }, validations: [{ logLevel: LogLevel.Warning, validation: 'x' }] }]);
		expect(uuidValidations(player as any, undefined)).toEqual([]);
	});

	it('returns [] for a uuid with no match', () => {
		const player = makeStatsPlayer([{ uuid: { value: 'a' }, validations: [{ logLevel: LogLevel.Warning, validation: 'x' }] }]);
		expect(uuidValidations(player as any, 'missing')).toEqual([]);
	});

	it('returns the matching entry’s validations', () => {
		const target = [{ logLevel: LogLevel.Error, validation: 'boom' }];
		const player = makeStatsPlayer([
			{ uuid: { value: 'a' }, validations: [{ logLevel: LogLevel.Warning, validation: 'x' }] },
			{ uuid: { value: 'b' }, validations: target },
		]);
		expect(uuidValidations(player as any, 'b')).toEqual(target);
	});
});
