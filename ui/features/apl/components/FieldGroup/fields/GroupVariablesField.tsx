import { placeholderNames } from '@features/apl/model/placeholders';
import { rotationSource } from '@features/apl/utils';
import { type ListItemsConfig, useListItems } from '@features/inputs/hooks/useListItems';
import type { InputConfig } from '@features/inputs/input';
import i18n from '@i18n/config';
import { translateItemLabel } from '@i18n/localization';
import type { Player } from '@sim/player/player';
import { randomUUID } from '@sim/utils/misc';
import { Field } from '@ui-kit/Field';
import { Tree } from '@ui-kit/Tree';
import { useCallback, useRef } from 'react';

import { GroupVariableRow } from './GroupVariableRow';

export interface GroupVariablesFieldProps {
	player: Player<any>;
	config: InputConfig<Player<any>, Array<any>> & { id: string };
	/** The sibling field naming the group being referenced. */
	groupNameField: string;
	getParentValue: () => any;
}

const newVariable = (name: string) => ({
	name,
	value: { uuid: { value: randomUUID() }, value: { oneofKind: 'variableRef', variableRef: { name: '' } } },
});

/** One array for every guarded read, so two of them are the same snapshot and re-render nothing. */
const NOTHING_TO_PASS: Array<any> = [];

/**
 * The variables a referenced group expects, one row per placeholder the group defines.
 *
 * The row set is derived: it is exactly the placeholder names of the selected group, and
 * `reconcile` rewrites the stored `variables` to that set, keeping whatever each name was already
 * assigned. It is a write on a read path, and **every** read goes through it: the list's own read
 * and each row's re-read of its entry are separate `useSyncExternalStore` subscribers, either may
 * read first on a notification, and a reader that skipped the reconcile would cache a row set the
 * other is about to replace. Calling it twice is safe because it is idempotent — the second call
 * finds every entry and writes nothing.
 *
 * Delete therefore acts on a list rebuilt underneath it: the row returns with its assignment
 * cleared, which is the useful reading of the button and the only one it can have. Copy is not
 * offered — a duplicate entry carries an existing name, so the next reconcile drops it — and
 * neither are add or drag, because the group fixes the set and its order.
 *
 * With no group selected, one whose name resolves to nothing, or one with no placeholders, there is
 * nothing to derive: the stored `variables` are left exactly as the user wrote them and the
 * container is `hidden` — not the shell's `hide`, which is a different class.
 */
export const GroupVariablesField = ({ player, config, groupNameField, getParentValue }: GroupVariablesFieldProps) => {
	const parentRef = useRef(getParentValue);
	parentRef.current = getParentValue;

	const reconcile = useCallback((): Array<any> => {
		const parentValue = parentRef.current();
		const groupName = parentValue?.[groupNameField];
		const group = groupName ? (player.aplRotation?.groups || []).find(candidate => candidate.name === groupName) : undefined;
		const names = placeholderNames(group);
		// One check for three cases: no group named, a name that resolves to nothing — deleting
		// a group leaves every reference to it dangling — and a group that declares no placeholder all
		// land here, and none of them may write. Deriving an empty set from them would empty the live
		// rotation's `variables` on a render alone, and retribution's default APL ships a reference
		// passing a variable to a group that declares no placeholder at all.
		if (!names.length) return NOTHING_TO_PASS;
		const existing: Array<any> = parentValue.variables || [];
		const reconciled = names.map(name => existing.find(entry => entry.name === name) || newVariable(name));
		const changed = reconciled.length !== existing.length || reconciled.some((entry, index) => entry !== existing[index]);
		if (changed) parentValue.variables = reconciled;
		return changed ? reconciled : existing;
	}, [player, groupNameField]);

	const listConfig: ListItemsConfig<Player<any>, any> = {
		storeSubscribe: rotationSource,
		getValue: () => reconcile().slice(),
		setValue: (subject: Player<any>, next: Array<any>) => {
			const parentValue = parentRef.current();
			if (parentValue) parentValue.variables = next;
			config.setValue(subject, next);
		},
		newItem: () => {
			throw new Error('newItem should not be called for auto-populated group variables');
		},
	};
	const { items, itemConfig, remove } = useListItems(player, listConfig);

	const itemLabel = translateItemLabel('Variable');
	return (
		<Field.Root
			layout={config.layout}
			className={['w-full', config.extraClassNames, !items.length && 'hidden']}
			hidden={!items.length}
			testId="group-reference-variables-container">
			<Field.Label as="span" tooltip={i18n.t('rotation_tab.apl.actions.group_reference.variables_tooltip')}>
				{i18n.t('rotation_tab.apl.actions.group_reference.variables')}
			</Field.Label>
			<Tree.Root group="apl-group-variables" layout="inline" density="compact" testId="apl-group-variables">
				<Tree.Items>
					{items.map((entry, index) => (
						<Tree.Item key={entry.name} id={entry.name} testId="apl-group-variables-item">
							<Tree.Item.Header>
								<Tree.Item.Actions label={i18n.t('common.list_picker.item_actions', { itemLabel })}>
									<Tree.Item.Action
										label={i18n.t('common.list_picker.delete_item', { itemLabel })}
										icon="xmark"
										tone="danger"
										testId="apl-group-variables-delete"
										onClick={() => remove(index)}
									/>
								</Tree.Item.Actions>
							</Tree.Item.Header>
							<Tree.Item.Body>
								<GroupVariableRow player={player} config={itemConfig(entry, index)} name={entry.name} />
							</Tree.Item.Body>
						</Tree.Item>
					))}
				</Tree.Items>
			</Tree.Root>
		</Field.Root>
	);
};
