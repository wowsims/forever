import { ActionId } from '@sim/proto/action_id';
import type { StoreSubscribe } from '@sim/state/subscriptions';

// A filled ActionId as fill() returns one, so the icon resolves synchronously and no test touches the network.
export const filled = (actionId: ActionId, name: string, iconUrl: string) =>
	Object.assign(Object.create(ActionId.prototype), actionId, { name, iconUrl }) as ActionId;

/**
 * Stands in for a domain facade in the input tests: a value plus the `(onChange) => unsubscribe`
 * contract every storeSubscribe helper in state/subscriptions.ts returns.
 */
export class FakeSource<T> {
	private listeners = new Set<() => void>();
	constructor(public value: T) {}
	set(next: T) {
		this.value = next;
		this.notify();
	}
	notify() {
		this.listeners.forEach(listener => listener());
	}
	readonly subscribe: StoreSubscribe = listener => {
		this.listeners.add(listener);
		return () => this.listeners.delete(listener);
	};
	get listenerCount() {
		return this.listeners.size;
	}
}
