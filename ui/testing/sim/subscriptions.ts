import type { StoreSubscribe } from '@sim/state/subscriptions';
import { vi } from 'vitest';

import { noopSubscribe } from './fake_host';

type Subscriptions = typeof import('@sim/state/subscriptions');

/**
 * Every `subscribe*` export answered from one source; non-subscription exports stay real.
 *
 * `vi.mock` is hoisted above the imports, so the factory has to reach this module lazily:
 *
 * ```ts
 * vi.mock('@sim/state/subscriptions', async () => (await import('@testing/sim/subscriptions')).mockSubscriptions(source.subscribe));
 * ```
 */
export const mockSubscriptions = async (
	source: StoreSubscribe = noopSubscribe,
	overrides: Partial<Record<keyof Subscriptions, unknown>> = {},
): Promise<Subscriptions> => {
	const actual = await vi.importActual<Subscriptions>('@sim/state/subscriptions');
	const answered = Object.keys(actual).map(name => [name, name.startsWith('subscribe') ? () => source : (actual as Record<string, unknown>)[name]]);
	return { ...Object.fromEntries(answered), ...overrides } as Subscriptions;
};
