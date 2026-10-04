import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import { keyFor } from '@features/inputs/list_ids';
import { Target as TargetProto } from '@generated/proto/common';
import i18n from '@i18n/config';
import { Encounter } from '@sim/raid/encounter';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Tree, type TreeDropRef, type TreeRef, type TreeSortable } from '@ui-kit/Tree';
import { useMemo, useRef } from 'react';

import { trackEvent } from '../../../../tracking/utils';
import { TargetPicker } from './TargetPicker';

const GROUP = 'targets';

export interface TargetsPickerProps {
	encounter: Encounter;
}

export const TargetsPicker = ({ encounter }: TargetsPickerProps) => {
	const config = useMemo(
		(): ListItemsConfig<Encounter, TargetProto> => ({
			storeSubscribe: (subject: Encounter) => subscribeEncounterField(subject, 'targets'),
			getValue: (subject: Encounter) => subject.getTargets().slice(),
			setValue: (subject: Encounter, newValue: Array<TargetProto>) => {
				trackEvent({
					action: 'settings',
					category: 'encounter',
					label: newValue.length > subject.getTargets().length ? 'add-target' : 'remove-target',
				});
				subject.setTargets(newValue);
			},
			newItem: () => Encounter.defaultTargetProto(),
			copyItem: (oldItem: TargetProto) => TargetProto.clone(oldItem),
			minimumItems: 1,
		}),
		[],
	);
	const { items, create, remove, copy, move, canRemove, hidden, disabled } = useListItems(encounter, config);
	const moveRef = useRef(move);
	moveRef.current = move;
	const idsRef = useRef(new Set<string>());
	idsRef.current = new Set(items.map((target, targetIndex) => String(keyFor(target, targetIndex))));
	// The lists nested in a target row share this list's drag, so a move counts only when it is a
	// target moving among the targets.
	const sortable = useMemo((): TreeSortable => {
		const isOwn = (from: TreeRef, to: TreeDropRef) =>
			from.group === GROUP && to.group === GROUP && from.parentId === to.parentId && idsRef.current.has(from.id);
		return {
			canDrop: isOwn,
			onMove: ({ from, to }) => {
				if (isOwn(from, to)) moveRef.current(from.index, to.index);
			},
		};
	}, []);
	if (hidden) return null;

	const itemLabel = i18n.t('settings_tab.encounter.target');
	return (
		<Tree.Root group={GROUP} sortable={sortable} disabled={disabled} testId="targets-picker">
			<Tree.Items>
				{items.map((target, targetIndex) => {
					const key = String(keyFor(target, targetIndex));
					return (
						<Tree.Item key={key} id={key} testId="targets-picker-item">
							<Tree.Item.Header>
								<Tree.Item.Handle label={i18n.t('common.list_picker.move_drag_drop')} testId="targets-picker-handle" />
								<Tree.Item.Title as="h6">{`${itemLabel} ${targetIndex + 1}`}</Tree.Item.Title>
								<Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel })}>
									<Tree.Item.Action
										label={i18n.t('common.list_picker.copy_to_new', { itemLabel })}
										icon="copy"
										testId="targets-picker-copy"
										onClick={() => copy(targetIndex)}
									/>
									{canRemove(targetIndex) && (
										<Tree.Item.Action
											label={i18n.t('common.list_picker.delete_item', { itemLabel })}
											icon="xmark"
											tone="danger"
											testId="targets-picker-delete"
											onClick={() => remove(targetIndex)}
										/>
									)}
								</Tree.Item.Actions>
							</Tree.Item.Header>
							<Tree.Item.Body>
								<TargetPicker encounter={encounter} targetIndex={targetIndex} />
							</Tree.Item.Body>
						</Tree.Item>
					);
				})}
			</Tree.Items>
			<Tree.AddButton label={i18n.t('common.list_picker.new_item', { itemLabel })} onClick={create} />
		</Tree.Root>
	);
};
