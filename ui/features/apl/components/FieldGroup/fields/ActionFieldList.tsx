import { ActionPicker } from '@features/apl/components/ActionPicker';
import { AplItemActions, AplRevealTarget } from '@features/apl/components/AplList';
import { AplProvider, useApl } from '@features/apl/context/AplContext';
import { useOwnListSortable } from '@features/apl/hooks/useOwnListSortable';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import type { InputConfig } from '@features/inputs/input';
import { keyFor } from '@features/inputs/list_ids';
import { APLAction } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Tree } from '@ui-kit/Tree';
import clsx from 'clsx';
import { type ReactNode, useMemo } from 'react';

const UNMARKED: ReadonlyArray<number> = [];

interface MarkedStepProps {
	markedSteps: ReadonlyArray<number>;
	index: number;
	children: ReactNode;
}

// The step on the way to the one that fired hands the rest of the indexes down; every other step hands down an
// empty list, since `AplProvider` passes its parent's on for `undefined`.
const MarkedStep = ({ markedSteps, index, children }: MarkedStepProps) => {
	const onPath = markedSteps[0] === index;
	const below = useMemo(() => (onPath ? markedSteps.slice(1) : UNMARKED), [onPath, markedSteps]);
	return (
		<AplProvider markedSteps={below}>
			<div
				data-marked={onPath && markedSteps.length === 1 ? '' : undefined}
				className="data-marked:bg-primary-dampened data-marked:outline-2 data-marked:-outline-offset-2 data-marked:outline-primary data-marked:outline-solid">
				{children}
			</div>
		</AplProvider>
	);
};

export interface ActionFieldListProps {
	player: Player<any>;
	config: InputConfig<Player<any>, Array<APLAction>> & { id: string };
}

/**
 * The action list of a sequence or strict sequence — the action tree's own recursion edge.
 *
 * Its rows sit in the group the enclosing `ActionPicker` names after its sequence (`actionListGroup`),
 * which is not the priority list's, so a sequence's rows never mix with a priority list's. A row's
 * id is its list key: an `APLAction` carries no uuid.
 */
export const ActionFieldList = ({ player, config }: ActionFieldListProps) => {
	const listConfig: ListItemsConfig<Player<any>, APLAction> = {
		...config,
		newItem: APLAction.create,
		// Called through, not passed: `clone` reads `this`, so the bare method copies nothing.
		copyItem: (oldItem: APLAction) => freshUuids(APLAction.clone(oldItem)),
	};
	const { listGroup, markedSteps } = useApl();
	const group = listGroup ?? 'apl-actions';
	const { items, itemConfig, create, remove, copy, move, hidden, disabled } = useListItems(player, listConfig);
	const sortable = useOwnListSortable(move, group);
	if (hidden) return null;

	const itemLabel = i18n.t('rotation_tab.apl.priorityList.name');
	return (
		<Tree.Root
			group={group}
			layout="inline"
			density="compact"
			sortable={sortable}
			disabled={disabled}
			className={clsx(config.extraClassNames)}
			testId="apl-action-list">
			<Tree.Items>
				{items.map((action, index) => {
					const key = String(keyFor(action, index));
					const stepConfig = itemConfig(action, index);
					const picker = <ActionPicker player={player} config={stepConfig} />;
					return (
						<Tree.Item key={key} id={key} testId="apl-action-list-item">
							<Tree.Item.Header>
								<Tree.Item.Handle label={i18n.t('common.list_picker.move_drag_drop')} testId="apl-action-list-handle" />
								<Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel })}>
									<AplItemActions
										label={itemLabel}
										onCopy={() => copy(index)}
										onDelete={() => remove(index)}
										testIdPrefix="apl-action-list"
									/>
								</Tree.Item.Actions>
							</Tree.Item.Header>
							<Tree.Item.Body>
								<AplRevealTarget getItem={() => stepConfig.getValue(player)}>
									{markedSteps.length ? (
										<MarkedStep markedSteps={markedSteps} index={index}>
											{picker}
										</MarkedStep>
									) : (
										picker
									)}
								</AplRevealTarget>
							</Tree.Item.Body>
						</Tree.Item>
					);
				})}
			</Tree.Items>
			<Tree.AddButton label={i18n.t('common.list_picker.new_item', { itemLabel })} icon="plus" tone="success" onClick={create} />
		</Tree.Root>
	);
};
