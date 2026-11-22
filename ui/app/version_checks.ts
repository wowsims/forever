export const VERSION_CHECK_EVERY_MS = 10 * 60 * 1000;

export interface VersionChecks {
	checkNow: () => void;
	stop: () => void;
}

/**
 * For a page that has just checked: calls `check` every `VERSION_CHECK_EVERY_MS` while the page is visible, when it
 * becomes visible again after that long, and when Back or Forward brings it back from memory as it was left.
 */
export const checkWhileOpen = (check: () => void): VersionChecks => {
	let lastCheck = Date.now();
	const checkNow = () => {
		lastCheck = Date.now();
		check();
	};
	const checkIfDue = () => {
		if (document.visibilityState === 'visible' && Date.now() - lastCheck >= VERSION_CHECK_EVERY_MS) checkNow();
	};
	const onPageShow = (event: PageTransitionEvent) => {
		if (event.persisted) checkNow();
	};
	document.addEventListener('visibilitychange', checkIfDue);
	window.addEventListener('pageshow', onPageShow);
	const timer = setInterval(checkIfDue, VERSION_CHECK_EVERY_MS);
	return {
		checkNow,
		stop: () => {
			document.removeEventListener('visibilitychange', checkIfDue);
			window.removeEventListener('pageshow', onPageShow);
			clearInterval(timer);
		},
	};
};
