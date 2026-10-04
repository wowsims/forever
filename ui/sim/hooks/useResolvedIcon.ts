import type { ActionId } from '@sim/proto/action_id';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { useMemo } from 'react';

import { useActionId } from './useActionId';

/** An `ActionId` as the icon a kit component takes: `undefined` until it has resolved to an icon url. */
export const useResolvedIcon = (actionId: ActionId | undefined): ResolvedIcon | undefined => {
	const { iconUrl, name, href, ready } = useActionId(actionId);
	return useMemo(() => (ready && iconUrl ? { iconUrl, name, href: href || undefined } : undefined), [ready, iconUrl, name, href]);
};
