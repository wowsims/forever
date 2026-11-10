import i18n from '@i18n/config';
import { Box } from '@ui-kit/Box';
import type { ReactNode } from 'react';

import { BuildBundles } from '../BuildBundles';
import { EncounterList } from '../EncounterList';

export interface BulkBuildsColumnsViewProps {
	builds: ReactNode;
	encounters: ReactNode;
}

type Column = keyof BulkBuildsColumnsViewProps;

const COLUMNS: readonly Column[] = ['builds', 'encounters'];

export const BulkBuildsColumnsView = (props: BulkBuildsColumnsViewProps) => (
	<div className="grid grid-cols-1 gap-4 md:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
		{COLUMNS.map(column => (
			<Box key={column} className="grid content-start gap-4" testId={`bulk-builds-column-${column}`}>
				<div className="grid gap-2">
					<h5 className="mb-0">{i18n.t(`bulk_tab.builds.columns.${column}.title`)}</h5>
					<p className="mb-0 text-xs">{i18n.t(`bulk_tab.builds.columns.${column}.hint`)}</p>
				</div>
				{props[column]}
			</Box>
		))}
	</div>
);

export const BulkBuildsColumns = () => <BulkBuildsColumnsView builds={<BuildBundles />} encounters={<EncounterList />} />;
