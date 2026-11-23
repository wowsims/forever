import { usePlayer } from '@sim/context/SimHostContext';
import { bulkState } from '@sim/settings/bulk_settings';
import { Box } from '@ui-kit/Box';
import { type ReactNode, useState } from 'react';

import { useDropsSetup } from '../../hooks/useDropsSetup';
import { setDropsItemsIncluded, setDropsSources } from '../../model/drops/settings';
import { DROPS_L60_DUNGEON_ZONE_IDS, dropsQuickPickTicks, dropsTicksAfter } from '../../model/drops/sources';
import { DropsItemsView } from '../DropsItems';
import { DropsSourcesView } from '../DropsSources';

export interface DropsSetupViewProps {
	sources: ReactNode;
	items: ReactNode;
}

export const DropsSetupView = ({ sources, items }: DropsSetupViewProps) => (
	<div className="grid grid-cols-1 gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]" data-testid="drops-setup">
		<Box className="grid content-start gap-4" testId="drops-setup-sources">
			{sources}
		</Box>
		<Box className="grid content-start gap-4" testId="drops-setup-items">
			{items}
		</Box>
	</div>
);

// Grouping is the pane's own state: the settings keep the sources, the excluded ids and the preset pick, not how the list reads.
export const DropsSetup = () => {
	const player = usePlayer();
	const { tree, ticked, list, names } = useDropsSetup();
	const [grouping, setGrouping] = useState<'slot' | 'source'>('slot');

	return (
		<DropsSetupView
			sources={
				<DropsSourcesView
					tree={tree}
					ticked={ticked}
					onTick={(node, checked) => setDropsSources(player, dropsTicksAfter(bulkState(player).dropsSources, node, checked))}
					onQuickPick={pick =>
						setDropsSources(player, pick === 'clear' ? [] : dropsQuickPickTicks(bulkState(player).dropsSources, tree, DROPS_L60_DUNGEON_ZONE_IDS))
					}
				/>
			}
			items={
				<DropsItemsView
					list={list}
					grouping={grouping}
					onGroupingChange={setGrouping}
					names={names}
					onIncludedChange={(ids, included) => setDropsItemsIncluded(player, ids, included)}
				/>
			}
		/>
	);
};
