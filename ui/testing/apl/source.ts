/**
 * One listener set behind every `subscribe*` export, for the APL component tests.
 *
 * `vi.hoisted` cannot reach an import, so a test reaches this module twice: statically, to notify,
 * and inside the `vi.mock` factory, which the hoisting runs before the imports. Both resolve the
 * same module instance, so a `notify()` in the test reaches the listeners the components added:
 *
 * ```ts
 * vi.mock('@sim/state/subscriptions', async () => {
 * 	const { mockSubscriptions } = await import('@testing/sim/subscriptions');
 * 	return mockSubscriptions((await import('@testing/apl/source')).source.subscribe);
 * });
 * ```
 *
 * `vi.resetModules()` evaluates this module again, and the factory then hands the components a new
 * set, so a test file that resets modules keeps a `vi.hoisted` source of its own.
 */
const listeners = new Set<() => void>();

export const source = {
	listeners,
	subscribe: (onChange: () => void) => {
		listeners.add(onChange);
		return () => listeners.delete(onChange);
	},
	notify: () => Array.from(listeners).forEach(listener => listener()),
};
