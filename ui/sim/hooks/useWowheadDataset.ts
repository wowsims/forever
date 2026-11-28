import { useEffect, useMemo, useState } from 'react';

/** `data-wowhead` for wowhead's tooltip, or the props that open the site's own spell tooltip (ui/sim/proto/tooltips.ts). */
export type WowheadDatasetProps = {
	'data-wowhead'?: string;
} & Record<string, string | undefined>;

/** Resolves one tooltip query and returns it as props to spread onto the anchor it describes: a string is wowhead's dataset, an object the props themselves. `null` is a subject with nothing to show, and `resolve`'s identity is what says the subject changed. */
export const useWowheadDataset = (resolve: (() => Promise<string | Record<string, string>>) | null): WowheadDatasetProps => {
	const [resolved, setResolved] = useState<{ resolve: typeof resolve; props?: WowheadDatasetProps }>({ resolve });

	if (resolved.resolve !== resolve) setResolved({ resolve });

	useEffect(() => {
		if (!resolve) return;
		resolve()
			.then(result => {
				const props = typeof result === 'string' ? { 'data-wowhead': result } : result;
				setResolved(current => (current.resolve === resolve ? { resolve, props } : current));
			})
			.catch(console.error);
	}, [resolve]);

	return useMemo(() => (resolved.resolve === resolve && resolved.props ? resolved.props : { 'data-wowhead': undefined }), [resolve, resolved]);
};
