import { SimHostProvider } from '@sim/context/SimHostContext';
import { ActionId } from '@sim/proto/action_id';
import { createSimStore } from '@sim/state/sim_store';
import { render } from '@testing-library/react';
import type { ReactNode } from 'react';
import { vi } from 'vitest';

/** A filled ActionId as fill() returns one: same id fields, plus name and iconUrl. */
export const filled = (actionId: ActionId, name: string, iconUrl: string) =>
	Object.assign(Object.create(ActionId.prototype), actionId, { name, iconUrl }) as ActionId;

/** Holds every `ActionId.fill` open; the test settles each one through `resolvers`, in call order. */
export const deferFill = () => {
	const resolvers: Array<(value: ActionId) => void> = [];
	const spy = vi.spyOn(ActionId.prototype, 'fill').mockImplementation(function (this: ActionId) {
		return new Promise<ActionId>(resolve => resolvers.push(resolve));
	});
	return { spy, resolvers };
};

export const mountWithSimStore = (children: ReactNode) => {
	const store = createSimStore();
	const host = { sim: { store } } as never;
	return { store, ...render(<SimHostProvider host={host}>{children}</SimHostProvider>) };
};
