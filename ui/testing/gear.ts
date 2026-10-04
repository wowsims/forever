const listeners = new Set<() => void>();

// A `vi.mock` factory reaches this through `await import`, which `vi.hoisted` cannot, and the test
// imports the same instance to fire it.
export const storeSource = {
	subscribe: (callback: () => void) => {
		listeners.add(callback);
		return () => listeners.delete(callback);
	},
	notify: () => listeners.forEach(listener => listener()),
};
