import { useEffect, useState } from 'react';

import type { Sim } from '../sim';
import { noop } from '../utils/misc';
import { useSimStatus } from './useSimStatus';

export interface OutdatedNativeSim {
	/** The running build's version, what an update has to differ from. */
	version: string;
	/** The newest release's tag. */
	latest: string;
	/** Whether the host can download that release and restart itself. Older hosts do not say, and cannot. */
	canUpdate: boolean;
}

/** The local Go host's version info when it runs an out-of-date build, otherwise null. Always null in the browser. */
export const useOutdatedNativeSim = (sim: Sim): OutdatedNativeSim | null => {
	const ready = useSimStatus(sim).status === 'ready';
	const [outdated, setOutdated] = useState<OutdatedNativeSim | null>(null);

	useEffect(() => {
		// `sim.isNative` is decided during init, so this cannot be asked any earlier.
		if (!ready || !sim.isNative) return;
		let cancelled = false;
		fetch('/version')
			.then(response =>
				response
					.json()
					.then(versionInfo => {
						if (!cancelled && versionInfo.outdated == 2) {
							setOutdated({ version: versionInfo.version ?? '', latest: versionInfo.latest ?? '', canUpdate: versionInfo.canUpdate === true });
						}
					})
					.catch(() => console.warn('No version info found!')),
			)
			.catch(noop);
		return () => {
			cancelled = true;
		};
	}, [ready, sim]);

	return outdated;
};
