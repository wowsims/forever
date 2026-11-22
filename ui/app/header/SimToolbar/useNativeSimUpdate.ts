import i18n from '@i18n/config';
import { useState } from 'react';

import { fetchNativeVersion, type OutdatedNativeSim } from './useOutdatedNativeSim';

export type NativeSimUpdateStatus = { state: 'idle' | 'updating' } | { state: 'error'; error: string };

const POLL_INTERVAL_MS = 1000;
const GIVE_UP_MS = 60_000;

const sleep = (ms: number) => new Promise<void>(resolve => setTimeout(resolve, ms));

/** Asks the host to update itself, waits for the new build to answer, then reloads the page. */
export const useNativeSimUpdate = (info: OutdatedNativeSim | null) => {
	const [status, setStatus] = useState<NativeSimUpdateStatus>({ state: 'idle' });

	const update = async () => {
		if (!info) return;
		setStatus({ state: 'updating' });
		try {
			const response = await fetch('/update', { method: 'POST' });
			if (!response.ok) {
				const error = await response
					.json()
					.then(body => String(body.error ?? response.statusText))
					.catch(() => response.statusText);
				setStatus({ state: 'error', error: i18n.t('info.native_sim_update_failed', { error }) });
				return;
			}
		} catch (error) {
			setStatus({ state: 'error', error: i18n.t('info.native_sim_update_failed', { error: String(error) }) });
			return;
		}

		const deadline = Date.now() + GIVE_UP_MS;
		while (Date.now() < deadline) {
			await sleep(POLL_INTERVAL_MS);
			const version = (await fetchNativeVersion())?.version;
			if (version != null && version !== info.version) {
				window.location.reload();
				return;
			}
		}
		setStatus({ state: 'error', error: i18n.t('info.native_sim_update_timeout') });
	};

	return { status, update };
};
