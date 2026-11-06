import i18n from '@i18n/config';
import { formatDurationSeconds } from '@sim/utils/format';
import { roundSecondsRemaining, smoothSecondsRemaining } from '@sim/utils/math';
import { useEffect, useRef, useState } from 'react';

export interface TimeLeftProps {
	/** The raw estimate, read once a second; undefined while there is none. */
	secondsRemaining: () => number | undefined;
}

const TICK_MS = 1000;

/** Ticks once a second, so the estimate neither jitters with every progress report nor renders the dialog. */
export const TimeLeft = ({ secondsRemaining }: TimeLeftProps) => {
	// Read through a ref: a caller passes a fresh closure each render, and restarting the timer would restart the smoothing.
	const read = useRef(secondsRemaining);
	read.current = secondsRemaining;
	const [secondsLeft, setSecondsLeft] = useState<number>();

	useEffect(() => {
		let shown: number | undefined;
		const timer = window.setInterval(() => {
			const raw = read.current();
			shown = raw === undefined ? undefined : smoothSecondsRemaining(shown, raw, TICK_MS / 1000);
			setSecondsLeft(shown === undefined ? undefined : roundSecondsRemaining(shown));
		}, TICK_MS);
		return () => window.clearInterval(timer);
	}, []);

	if (secondsLeft === undefined) return null;
	return ` · ${i18n.t('common.time_left', { time: formatDurationSeconds(secondsLeft) })}`;
};
