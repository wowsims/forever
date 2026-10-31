import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { formatDurationSeconds } from '@sim/utils/format';
import { useEffect, useState } from 'react';

import { batchSecondsRemaining, roundSecondsRemaining, smoothSecondsRemaining } from '../../model/run/progress';
import { bulkProgress } from '../../model/run/run';

const TICK_MS = 1000;

/** Reads the batch's progress once a second, so the estimate neither jitters with every tick nor renders the dialog. */
export const BulkTimeLeft = () => {
	const player = usePlayer();
	const [secondsLeft, setSecondsLeft] = useState<number>();

	useEffect(() => {
		let shown: number | undefined;
		const timer = window.setInterval(() => {
			const steps = bulkProgress(player)?.steps;
			const raw = steps && batchSecondsRemaining(steps);
			shown = raw === undefined ? undefined : smoothSecondsRemaining(shown, raw, TICK_MS / 1000);
			setSecondsLeft(shown === undefined ? undefined : roundSecondsRemaining(shown));
		}, TICK_MS);
		return () => window.clearInterval(timer);
	}, [player]);

	if (secondsLeft === undefined) return null;
	return ` · ${i18n.t('bulk_tab.progress.time_left', { time: formatDurationSeconds(secondsLeft) })}`;
};
