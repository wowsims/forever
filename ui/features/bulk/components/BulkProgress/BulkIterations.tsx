import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useEffect, useState } from 'react';
import { Trans } from 'react-i18next';

import type { BulkProgress } from '../../model/progress';
import { bulkProgress, subscribeBulkProgress } from '../../model/run';

/** Subscribes to the run's own ticks, so a tick commits this leaf and never the dialog. */
export const BulkIterations = () => {
	const player = usePlayer();
	const [iterations, setIterations] = useState(() => bulkProgress(player)?.iterations);

	useEffect(() => subscribeBulkProgress(player, (progress: BulkProgress) => setIterations(progress.iterations)), [player]);

	if (!iterations) return null;
	return (
		<p className="mb-0 text-center" data-testid="results-sim">
			<Trans i18n={i18n} i18nKey="bulk_tab.progress.iterations_complete" values={{ completed: iterations.completed, total: iterations.total }} />
		</p>
	);
};
