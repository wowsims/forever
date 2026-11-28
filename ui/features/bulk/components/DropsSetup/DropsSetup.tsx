import { filterOptionsOf } from '@features/gear/model/filter_options';
import { DatabaseFilters } from '@generated/proto/ui';
import { usePlayer } from '@sim/context/SimHostContext';
import { bulkState } from '@sim/settings/bulk_settings';
import { Box } from '@ui-kit/Box';
import { type ReactNode, useMemo, useState } from 'react';

import { useBulkState } from '../../hooks/useBulkState';
import { useDropsSetup } from '../../hooks/useDropsSetup';
import { dropsActiveFilterCount, dropsDefaultFilters, dropsHiddenTicked } from '../../model/drops/filters';
import { dropsFiltersOf, setDropsFilters, setDropsItemsIncluded, setDropsSources } from '../../model/drops/settings';
import { dropsQuickPickTicks, dropsTicksAfter } from '../../model/drops/sources';
import { DropsItemsCardView } from '../DropsItemsCard';
import { type DropsGrouping, DropsItemsDialogView } from '../DropsItemsDialog';
import { DropsSourcesView } from '../DropsSources';

export interface DropsSetupViewProps {
	sources: ReactNode;
	items: ReactNode;
}

export const DropsSetupView = ({ sources, items }: DropsSetupViewProps) => (
	<div className="grid grid-cols-1 gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,3fr)] md:items-start" data-testid="drops-setup">
		<Box className="grid content-start gap-4" testId="drops-setup-sources">
			{sources}
		</Box>
		<Box className="grid content-start gap-4" testId="drops-setup-items">
			{items}
		</Box>
	</div>
);

// Grouping and the dialog's search and pick are the pane's own state: the settings keep the sources, the excluded ids, the
// filters and the preset pick, not how the list reads.
export const DropsSetup = () => {
	const player = usePlayer();
	const playerClass = player.getPlayerClass();
	const { tree, ticked, list, names, dungeonLevels } = useDropsSetup();
	const stored = useBulkState(slice => slice.dropsFilters);
	const filters = useMemo(() => stored ?? dropsDefaultFilters(playerClass), [stored, playerClass]);
	const options = useMemo(
		() =>
			filterOptionsOf(
				list.items.map(entry => entry.item),
				() => undefined,
			),
		[list],
	);
	const [grouping, setGrouping] = useState<DropsGrouping>('slot');
	const [open, setOpen] = useState(false);

	return (
		<DropsSetupView
			sources={
				<DropsSourcesView
					tree={tree}
					ticked={ticked}
					onTick={(node, checked) => setDropsSources(player, dropsTicksAfter(bulkState(player).dropsSources, node, checked))}
					onQuickPick={pick =>
						setDropsSources(player, pick === 'clear' ? [] : dropsQuickPickTicks(bulkState(player).dropsSources, tree, dungeonLevels))
					}
				/>
			}
			items={
				<>
					<DropsItemsCardView list={list} hidden={dropsHiddenTicked(list, filters)} onChoose={() => setOpen(true)} />
					<DropsItemsDialogView
						open={open}
						onOpenChange={setOpen}
						list={list}
						grouping={grouping}
						onGroupingChange={setGrouping}
						tree={tree}
						names={names}
						filters={filters}
						activeFilters={dropsActiveFilterCount(filters, playerClass)}
						filtersMenu={{
							onFiltersChange: edit => {
								const next = DatabaseFilters.clone(dropsFiltersOf(player));
								edit(next);
								setDropsFilters(player, next);
							},
							armorTypes: playerClass.armorTypes,
							weaponTypes: playerClass.weaponTypes.map(entry => entry.weaponType),
							rangedWeaponTypes: playerClass.rangedWeaponTypes,
							...options,
							canDualWield: player.getPlayerSpec().canDualWield,
							raceFaction: player.getFaction(),
						}}
						onIncludedChange={(ids, included) => setDropsItemsIncluded(player, ids, included)}
					/>
				</>
			}
		/>
	);
};
