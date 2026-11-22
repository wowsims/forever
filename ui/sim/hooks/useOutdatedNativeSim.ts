import { useEffect, useState } from 'react';

import type { Sim } from '../sim';
import { useSimStatus } from './useSimStatus';

export interface OutdatedNativeSim {
	/** The running build's version, what an update has to differ from. */
	version: string;
	/** The newest release's tag. */
	latest: string;
	/** Whether the host can download that release and restart itself. */
	canUpdate: boolean;
}

/** The local Go host's /version answer, or null when it does not answer, as while it restarts. */
export const fetchNativeVersion = async (): Promise<{ version?: string; outdated?: number; latest?: string; canUpdate?: boolean } | null> => {
	try {
		const response = await fetch('/version');
		return response.ok ? await response.json() : null;
	} catch {
		return null;
	}
};

/** The local Go host's version info when it runs an out-of-date build, otherwise null. Always null in the browser. */
export const useOutdatedNativeSim = (sim: Sim): OutdatedNativeSim | null => {
	const ready = useSimStatus(sim).status === 'ready';
	const [outdated, setOutdated] = useState<OutdatedNativeSim | null>(null);

	useEffect(() => {
		// `sim.isNative` is decided during init, so this cannot be asked any earlier.
		if (!ready || !sim.isNative) return;
		let cancelled = false;
		fetchNativeVersion().then(versionInfo => {
			if (!cancelled && versionInfo?.outdated == 2) {
				setOutdated({ version: versionInfo.version ?? '', latest: versionInfo.latest ?? '', canUpdate: versionInfo.canUpdate === true });
			}
		});
		return () => {
			cancelled = true;
		};
	}, [ready, sim]);

	return outdated;
};
