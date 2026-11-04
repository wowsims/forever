import i18n from '@i18n/config';
import { BUILD_INFO } from '@sim/constants/build_info';
import { REPO_NAME } from '@sim/constants/other';
import { isDevMode } from '@sim/utils/env';
import { Button } from '@ui-kit/Button';
import { toastManager } from '@ui-kit/Toast';

// Back, Forward and a tab the browser restores at start show the page from the browser's cache without asking the
// server: the host's `max-age=0, must-revalidate` does not hold on those loads, and the back/forward cache keeps the
// whole page in memory. The files the sim fetches by a fixed name after the load (lib.wasm.gz and the workers) still
// come from the server, so such a page runs its old code against the new sim until it is reloaded. A `no-store`
// fetch asks the server on those loads too.
const VERSION_URL = `/${REPO_NAME}/version.json`;
const CHECK_EVERY_MS = 10 * 60 * 1000;
// The build this tab last reloaded for, kept across the reload.
const RELOADED_FOR_KEY = 'forever_reloaded_for_build';

const liveCommit = async (): Promise<string | undefined> => {
	try {
		const response = await fetch(VERSION_URL, { cache: 'no-store' });
		// The host answers a missing file with a page, which does not parse.
		const { commit } = await response.json();
		return response.ok && typeof commit === 'string' && commit ? commit : undefined;
	} catch {
		return undefined;
	}
};

const offerReload = () => {
	toastManager.add({
		variant: 'info',
		title: i18n.t('sim.new_build.title'),
		autohide: false,
		body: (
			<>
				<p>{i18n.t('sim.new_build.message')}</p>
				<Button variant="outline-light" onClick={() => location.reload()}>
					{i18n.t('sim.new_build.reload_button')}
				</Button>
			</>
		),
	});
};

/** Reloads a page that opens on an older build than the deploy, and offers a reload once a newer one is deployed while it is open. */
export const watchForNewBuild = () => {
	if (isDevMode() || !BUILD_INFO.commit) return;

	// A link's settings are read from the hash and then cleared, so the reload goes back to the URL the page opened at.
	const openedAt = location.href;
	let lastCheck = 0;
	let offered = false;
	const newerBuild = async () => {
		lastCheck = Date.now();
		const commit = await liveCommit();
		return commit && commit !== BUILD_INFO.commit ? commit : undefined;
	};

	// Nothing has been done yet on a page that has just opened, so it reloads. It does so once per build, so a server
	// that still answers with the old page cannot loop it.
	void newerBuild().then(commit => {
		if (!commit) return;
		try {
			if (sessionStorage.getItem(RELOADED_FOR_KEY) === commit) return;
			sessionStorage.setItem(RELOADED_FOR_KEY, commit);
		} catch {
			return;
		}
		history.replaceState(history.state, '', openedAt);
		location.reload();
	});

	// A page that has been open for a while may hold results, so it asks first.
	const check = () => {
		if (offered) return;
		void newerBuild().then(commit => {
			if (!commit || offered) return;
			offered = true;
			offerReload();
		});
	};
	const checkIfDue = () => {
		if (document.visibilityState === 'visible' && Date.now() - lastCheck >= CHECK_EVERY_MS) check();
	};
	document.addEventListener('visibilitychange', checkIfDue);
	setInterval(checkIfDue, CHECK_EVERY_MS);
	// Back and Forward can bring the page back from memory as it was left.
	window.addEventListener('pageshow', event => {
		if (event.persisted) check();
	});
	// The host deletes the old build's chunks on a deploy, so a chunk the page imports late fails to load.
	window.addEventListener('vite:preloadError', check);
};
