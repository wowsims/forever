import type { ActionId } from '@sim/proto/action_id';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { useEffect, useRef, useState } from 'react';

import { actionIdHref, actionIdKey, isFilled } from './useActionId';

type Icons = ReadonlyArray<ResolvedIcon | undefined>;

const iconOf = (actionId: ActionId | undefined): ResolvedIcon | undefined =>
	actionId?.iconUrl ? { iconUrl: actionId.iconUrl, name: actionId.name, href: actionIdHref(actionId) || undefined } : undefined;

/**
 * A list of `ActionId`s as the icons a kit component takes, index for index; an id with no icon,
 * or whose fill fails, resolves to `undefined` without holding back the others. The ids are compared by content, so a fresh array of the same ids per
 * render neither refetches nor changes the returned array.
 */
export const useActionIds = (actionIds: ReadonlyArray<ActionId | undefined>): Icons => {
	const key = actionIds.map(actionIdKey).join(',');
	const idsRef = useRef(actionIds);
	idsRef.current = actionIds;

	const [state, setState] = useState<{ key: string; icons: Icons }>(() => ({ key, icons: actionIds.map(iconOf) }));
	// Re-seeding during the render that changed the ids keeps the previous ids' icons from being painted under the new ones.
	if (state.key !== key) setState({ key, icons: actionIds.map(iconOf) });

	useEffect(() => {
		const ids = idsRef.current;
		if (ids.every(id => !id || isFilled(id))) return;

		const controller = new AbortController();
		void Promise.allSettled(ids.map(id => (id && !isFilled(id) ? id.fill() : Promise.resolve(id)))).then(results => {
			if (controller.signal.aborted) return;
			setState({ key, icons: results.map(result => iconOf(result.status === 'fulfilled' ? result.value : undefined)) });
			const failed = results.find((result): result is PromiseRejectedResult => result.status === 'rejected');
			if (failed) throw failed.reason;
		});
		return () => controller.abort();
	}, [key]);

	return state.icons;
};
