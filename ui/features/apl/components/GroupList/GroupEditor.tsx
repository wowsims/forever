import { AplItemActions, AplList, AplSectionTitle } from '@features/apl/components/AplList';
import { AplListItem } from '@features/apl/components/AplListItem';
import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplPasteDialog } from '@features/apl/components/AplPasteDialog';
import { ListItemHeader } from '@features/apl/components/ListItemHeader';
import { uuidValidations } from '@features/apl/components/ListItemHeader/utils';
import { NameDisplay } from '@features/apl/components/NameDisplay';
import { copyAplClip, useAplPaste } from '@features/apl/hooks/useAplClipboard';
import { useAplInput } from '@features/apl/hooks/useAplInput';
import { useVariableExtraction } from '@features/apl/hooks/useVariableExtraction';
import { aplItemId } from '@features/apl/model/apply_move';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import { rotationSource } from '@features/apl/utils';
import { useListItems } from '@features/inputs/hooks/useListItems';
import type { InputConfig } from '@features/inputs/input';
import { APLGroup, APLListItem } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { renameAPLReference } from '@sim/proto/apl_utils';
import { Tree } from '@ui-kit/Tree';
import { useState } from 'react';

export interface GroupEditorProps {
	/** The group's tree item id, `aplItemId(group, index)`; its actions' moves name it as their parent. */
	id: string;
	player: Player<any>;
	config: InputConfig<Player<any>, APLGroup>;
	/** Position in the rotation's group list, which is how the sim reports this group's validations. */
	groupIndex: number;
	onCopy: () => void;
	onDelete: () => void;
}

/** One action group: its name in the title, and the priority list it holds as the item's children. */
export const GroupEditor = ({ id, player, config, groupIndex, onCopy, onDelete }: GroupEditorProps) => {
	const { value: group, hidden, disabled } = useAplInput(player, config);
	const [renaming, setRenaming] = useState(false);

	const actions = useListItems(player, {
		storeSubscribe: rotationSource,
		getValue: () => config.getValue(player)?.actions || [],
		setValue: (subject: Player<any>, next: Array<APLListItem>) => {
			const current = config.getValue(subject);
			if (!current) return;
			current.actions = next;
			subject.touchRotation();
		},
		newItem: () => APLListItem.create({ action: {} }),
		copyItem: item => freshUuids(APLListItem.clone(item)),
	});
	const extraction = useVariableExtraction(
		player,
		index => config.getValue(player)?.actions?.[index]?.action?.condition,
		(index, reference) => {
			config.getValue(player)!.actions[index].action!.condition = reference;
		},
	);
	const pasting = useAplPaste('action', actions);
	const groupLabel = i18n.t('rotation_tab.apl.actionGroups.name');
	const label = i18n.t('rotation_tab.apl.priorityList.name');

	if (hidden) return null;
	return (
		<AplList.Item
			id={id}
			label={groupLabel}
			title={<NameDisplay name={group?.name || ''} onRename={() => setRenaming(true)} />}
			actions={
				<AplItemActions
					label={groupLabel}
					onCopy={onCopy}
					onCopyToClipboard={() => copyAplClip({ kind: 'group', item: config.getValue(player) })}
					onDelete={onDelete}
				/>
			}
			disabled={disabled}
			testId="apl-group-editor-root">
			<Tree.Item.Children testId="apl-group-actions">
				<span className="flex items-center gap-2">
					<AplSectionTitle
						text={i18n.t('rotation_tab.apl.actionGroups.attributes.actions')}
						tooltip={i18n.t('rotation_tab.apl.actionGroups.tooltips.actions')}
					/>
				</span>
				{actions.items.length > 0 && (
					<AplList.Rows>
						{actions.items.map((item, index) => {
							const itemConfig = actions.itemConfig(item, index);
							const itemId = aplItemId(item, index);
							return (
								<AplList.Item
									key={itemId}
									id={itemId}
									label={label}
									muted={item.hide}
									aside={
										<ListItemHeader
											player={player}
											getItem={itemConfig.getValue}
											getValidations={subject => [
												...(subject.getCurrentStats().rotationStats?.groups?.[groupIndex]?.actions?.[index]?.validations || []),
												...uuidValidations(subject, itemConfig.getValue(subject)?.action?.condition?.uuid?.value),
											]}
										/>
									}
									actions={
										<AplItemActions
											label={label}
											onCopy={() => actions.copy(index)}
											onCopyToClipboard={() => copyAplClip({ kind: 'action', item: itemConfig.getValue(player) })}
											onExtract={() => extraction.extract(index)}
											canExtract={() => extraction.canExtract(index)}
											onDelete={() => actions.remove(index)}
										/>
									}>
									<Tree.Item.Body>
										<AplListItem player={player} config={itemConfig} actionFullWidth />
									</Tree.Item.Body>
								</AplList.Item>
							);
						})}
					</AplList.Rows>
				)}
				<span className="flex items-center gap-1">
					<Tree.AddButton
						label={i18n.t('rotation_tab.apl.floatingActionBar.new', { itemName: label })}
						icon="plus"
						tone="success"
						onClick={() => actions.create()}
					/>
					<Tree.AddButton
						label={i18n.t('rotation_tab.apl.floatingActionBar.paste', { itemName: label })}
						icon="paste"
						testId="apl-paste"
						onClick={pasting.paste}
					/>
				</span>
			</Tree.Item.Children>
			<AplNameDialog
				open={renaming}
				title={i18n.t('rotation_tab.apl.nameModal.rename', { itemName: groupLabel })}
				inputLabel={i18n.t('rotation_tab.apl.actionGroups.attributes.name')}
				confirmLabel={i18n.t('rotation_tab.apl.nameModal.renameConfirm')}
				defaultValue={group?.name}
				existingNames={(player.aplRotation.groups || []).filter(other => other !== group).map(other => other.name)}
				onSubmit={name =>
					player.modifyAplRotation(rotation => {
						if (!group) return;
						renameAPLReference(rotation, { type: 'group', oldName: group.name, newName: name });
						group.name = name;
					})
				}
				onClose={() => setRenaming(false)}
			/>
			<AplNameDialog {...extraction.dialog} />
			<AplPasteDialog {...pasting.textDialog} />
		</AplList.Item>
	);
};
