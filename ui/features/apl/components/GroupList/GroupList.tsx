import { AplList, AplSectionTitle } from '@features/apl/components/AplList';
import { AplListToolbar } from '@features/apl/components/AplListToolbar';
import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplPasteDialog } from '@features/apl/components/AplPasteDialog';
import { useAplPaste } from '@features/apl/hooks/useAplClipboard';
import { useRenamedCopy } from '@features/apl/hooks/useRenamedCopy';
import { APL_GROUPS, aplItemId } from '@features/apl/model/apply_move';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import { rotationSource } from '@features/apl/utils';
import { useListItems } from '@features/inputs/hooks/useListItems';
import { APLGroup } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';

import { GroupEditor } from './GroupEditor';

const groupName = () => i18n.t('rotation_tab.apl.actionGroups.name');
const newGroup = (name: string) => APLGroup.create({ name, actions: [], variables: [] });

/**
 * The rotation's action groups, one tree item each with its actions under it.
 *
 * Creating, copying and pasting a group all ask for a name first, because a group is referenced by
 * name, so unlike every other list here none of them can just append a clone. A copy and a paste
 * fill the dialog with the next free name (`freeName`).
 */
export const GroupList = () => {
	const player = usePlayer();
	const groups = useListItems(player, {
		storeSubscribe: rotationSource,
		getValue: (subject: Player<any>) => subject.aplRotation.groups || [],
		setValue: (subject: Player<any>, next: Array<APLGroup>) =>
			subject.modifyAplRotation(rotation => {
				rotation.groups = next;
			}),
		newItem: () => newGroup(i18n.t('rotation_tab.apl.actionGroups.newGroupName')),
	});
	const pasting = useAplPaste('group', groups);
	const copying = useRenamedCopy(player, {
		itemName: groupName(),
		inputLabel: i18n.t('rotation_tab.apl.actionGroups.attributes.name'),
		read: rotation => rotation.groups || [],
		write: (rotation, next) => {
			rotation.groups = next;
		},
		nameOf: group => group.name,
		copy: (group, name) => {
			const clone = freshUuids(APLGroup.clone(group));
			clone.name = name;
			return clone;
		},
	});

	return (
		// The tab panel is a flex column with a gap; this stays a single flex item so that gap doesn't land between the header and the tree.
		<div className="flex flex-col">
			<h6 className="mb-2 flex items-center gap-2">
				<AplSectionTitle text={i18n.t('rotation_tab.apl.actionGroups.header')} tooltip={i18n.t('rotation_tab.apl.actionGroups.tooltips.overview')} />
			</h6>
			<AplList.Root group={APL_GROUPS} isGroup testId="apl-groups-tree">
				{groups.items.map((group, index) => {
					const id = aplItemId(group, index);
					return (
						<GroupEditor
							key={id}
							id={id}
							player={player}
							config={groups.itemConfig(group, index)}
							groupIndex={index}
							onCopy={() => copying.onCopyItem(index)}
							onDelete={() => groups.remove(index)}
						/>
					);
				})}
			</AplList.Root>
			<AplListToolbar
				itemName={groupName()}
				nameDialog={{
					inputLabel: i18n.t('rotation_tab.apl.actionGroups.attributes.name'),
					existingNames: () => groups.items.map(group => group.name),
				}}
				onCreate={name => groups.append(newGroup(name!))}
				onPaste={pasting.paste}
			/>
			<AplNameDialog {...copying.dialog} />
			<AplNameDialog {...pasting.nameDialog} />
			<AplPasteDialog {...pasting.textDialog} />
		</div>
	);
};
