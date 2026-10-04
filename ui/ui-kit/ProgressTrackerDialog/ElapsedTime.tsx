import { formatDurationSeconds } from '@sim/utils/format';
import { useEffect, useRef } from 'react';

export interface ElapsedTimeProps {
	running: boolean;
	/** `Date.now()` when the run began. */
	startedAt: number;
	testId?: string;
}

const TICK_MS = 100;
const ZERO = formatDurationSeconds(0);

/** `textContent` rather than state: the dialog is `keepMounted`, so a ticking `setState` would render it ten times a second for as long as the page lives. */
export const ElapsedTime = ({ running, startedAt, testId = 'time-elapsed' }: ElapsedTimeProps) => {
	const elapsed = useRef<HTMLSpanElement>(null);

	useEffect(() => {
		if (!running) return;
		const write = () => {
			if (elapsed.current) elapsed.current.textContent = formatDurationSeconds((Date.now() - startedAt) / 1000);
		};
		write();
		const timer = window.setInterval(write, TICK_MS);
		return () => window.clearInterval(timer);
	}, [running, startedAt]);

	return (
		<span ref={elapsed} data-testid={testId}>
			{ZERO}
		</span>
	);
};
