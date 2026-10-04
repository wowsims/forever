import { AplItemActions } from '@features/apl/components/AplList';
import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplValidations } from '@features/apl/components/ListItemHeader';
import { uuidValidations } from '@features/apl/components/ListItemHeader/utils';
import { ValuePicker } from '@features/apl/components/ValuePicker';
import { useApl } from '@features/apl/context/AplContext';
import { useOwnListSortable } from '@features/apl/hooks/useOwnListSortable';
import { useVariableExtraction } from '@features/apl/hooks/useVariableExtraction';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import type { InputConfig } from '@features/inputs/input';
import { keyFor } from '@features/inputs/list_ids';
import { APLValue } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { translateItemLabel } from '@i18n/localization';
import type { Player } from '@sim/player/player';
import { randomUUID } from '@sim/utils/misc';
import { Tree } from '@ui-kit/Tree';
import clsx from 'clsx';

export interface ValueFieldListProps {
	player: Player<any>;
	config: InputConfig<Player<any>, Array<APLValue | undefined>> & { id: string };
}

const newValue = () => APLValue.create({ uuid: { value: randomUUID() } });

/**
 * The operand list of `and`, `or`, `min` and `max`.
 *
 * Each row is a whole `ValuePicker`, so this is one of the two edges the value tree recurses
 * across. The row's own validations ride in the item title, and a row's id is its value's uuid.
 */
export const ValueFieldList = ({ player, config }: ValueFieldListProps) => {
	const values = () => config.getValue(player) || [];
	const extraction = useVariableExtraction(
		player,
		index => values()[index],
		(index, reference) => {
			const list = values();
			list[index] = reference;
			config.setValue(player, list);
		},
	);

	const listConfig: ListItemsConfig<Player<any>, APLValue | undefined> = {
		...config,
		// An empty row is stored as a value with a uuid, never as a hole.
		setValue: (subject: Player<any>, next: Array<APLValue | undefined>) =>
			config.setValue(
				subject,
				next.map(value => value || newValue()),
			),
		newItem: newValue,
		copyItem: (oldValue: APLValue | undefined) => (oldValue ? freshUuids(APLValue.clone(oldValue)) : oldValue),
	};
	const group = useApl().listGroup ?? 'apl-values';
	const { items, itemConfig, create, remove, copy, move, hidden, disabled } = useListItems(player, listConfig);
	const sortable = useOwnListSortable(move, group);
	if (hidden) return null;

	const itemLabel = translateItemLabel('Value');
	return (
		<>
			<Tree.Root
				group={group}
				layout="inline"
				density="compact"
				sortable={sortable}
				disabled={disabled}
				className={clsx(config.extraClassNames)}
				testId="apl-value-list">
				<Tree.Items>
					{items.map((value, index) => {
						const key = String(keyFor(value, index));
						const valueConfig = itemConfig(value, index);
						return (
							<Tree.Item key={key} id={value?.uuid?.value || key} testId="apl-value-list-item">
								<Tree.Item.Header>
									<Tree.Item.Handle label={i18n.t('common.list_picker.move_drag_drop')} testId="apl-value-list-handle" />
									<Tree.Item.Aside>
										<AplValidations getValidations={subject => uuidValidations(subject, valueConfig.getValue(subject)?.uuid?.value)} />
									</Tree.Item.Aside>
									<Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel })}>
										<AplItemActions
											label={itemLabel}
											onCopy={() => copy(index)}
											onExtract={() => extraction.extract(index)}
											canExtract={() => extraction.canExtract(index)}
											onDelete={() => remove(index)}
											testIdPrefix="apl-value-list"
										/>
									</Tree.Item.Actions>
								</Tree.Item.Header>
								<Tree.Item.Body>
									<ValuePicker player={player} config={valueConfig} />
								</Tree.Item.Body>
							</Tree.Item>
						);
					})}
				</Tree.Items>
				<Tree.AddButton label={i18n.t('common.list_picker.new_item', { itemLabel })} icon="plus" tone="success" onClick={create} />
			</Tree.Root>
			<AplNameDialog {...extraction.dialog} />
		</>
	);
};
