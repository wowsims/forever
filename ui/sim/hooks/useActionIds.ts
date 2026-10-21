import type { ActionId } from '@sim/proto/action_id';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { useEffect, useRef, useState } from 'react';

import { actionIdHref, actionIdKey, isFilled } from './useActionId';

type Resolved<R> = ReadonlyArray<R | undefined>;

const iconOf = (actionId: ActionId | undefined): ResolvedIcon | undefined =>
	actionId?.iconUrl ? { iconUrl: actionId.iconUrl, name: actionId.name, href: actionIdHref(actionId) || undefined } : undefined;

const fillIcon = (actionId: ActionId) => (isFilled(actionId) ? undefined : actionId.fill().then(iconOf));

/**
 * A list of `ActionId`s, each resolved into what a component takes, index for index. `seed` gives an id's value on the
 * render that brings it, and `resolve` fetches the rest, answering `undefined` for an id with nothing to fetch, which
 * keeps its seed. A fetch that fails resolves its slot to `undefined` without holding back the others, and its
 * rejection surfaces once they settle. The ids are compared by content, so a fresh array of the same ids per render
 * neither refetches nor changes the returned array.
 */
export const useResolvedActionIds = <R>(
	actionIds: ReadonlyArray<ActionId | undefined>,
	seed: (actionId: ActionId | undefined) => R | undefined,
	resolve: (actionId: ActionId) => Promise<R | undefined> | undefined,
): Resolved<R> => {
	const key = actionIds.map(actionIdKey).join(',');
	const latest = useRef({ actionIds, seed, resolve });
	latest.current = { actionIds, seed, resolve };

	const [state, setState] = useState<{ key: string; values: Resolved<R> }>(() => ({ key, values: actionIds.map(id => seed(id)) }));
	// Re-seeding during the render that changed the ids keeps the previous ids' values from being painted under the new ones.
	if (state.key !== key) setState({ key, values: actionIds.map(id => seed(id)) });

	useEffect(() => {
		const { actionIds: ids, seed: seedOf, resolve: resolveOf } = latest.current;
		const pending = ids.map(id => (id ? resolveOf(id) : undefined));
		if (pending.every(promise => !promise)) return;

		const controller = new AbortController();
		void Promise.allSettled(pending.map((promise, index) => promise ?? Promise.resolve(seedOf(ids[index])))).then(results => {
			if (controller.signal.aborted) return;
			setState({ key, values: results.map(result => (result.status === 'fulfilled' ? result.value : undefined)) });
			const failed = results.find((result): result is PromiseRejectedResult => result.status === 'rejected');
			if (failed) throw failed.reason;
		});
		return () => controller.abort();
	}, [key]);

	return state.values;
};

/** A list of `ActionId`s as the icons a kit component takes, index for index; an id with no icon, or whose fill fails, is `undefined`. */
export const useActionIds = (actionIds: ReadonlyArray<ActionId | undefined>): Resolved<ResolvedIcon> => useResolvedActionIds(actionIds, iconOf, fillIcon);
