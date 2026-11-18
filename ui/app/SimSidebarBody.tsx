import { CharacterStats } from '@features/character-stats';
import { SimWarnings } from '@features/results/components/SimResultsPanel';
import type { Sim } from '@sim/sim';
import type { ReactNode } from 'react';

import type { SimHostObject } from './individual_sim_ui';
import { SimRunCard } from './SimRunCard';

export interface SimSidebarBodyProps {
	/** Null on the render that builds the containers it is constructed from. */
	host: SimHostObject<any> | null;
	sim: Sim;
	simReady: boolean;
	actions?: ReactNode;
}

/** The run card, the warnings and the character stats: the aside's from `lg` up, the dock's sheet below it. */
export const SimSidebarBody = ({ host, sim, simReady, actions }: SimSidebarBodyProps) => (
	<>
		<SimRunCard host={host} sim={sim} actions={actions} />
		{host && <SimWarnings warnings={host.warnings} ready={simReady} className="my-auto" />}
		{host && <CharacterStats className="mt-auto" />}
	</>
);
