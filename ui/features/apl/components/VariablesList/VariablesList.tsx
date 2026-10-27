import { AplList, AplSectionTitle } from '@features/apl/components/AplList';
import { AplListToolbar } from '@features/apl/components/AplListToolbar';
import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplPasteDialog } from '@features/apl/components/AplPasteDialog';
import { useAplPaste } from '@features/apl/hooks/useAplClipboard';
import { useRenamedCopy } from '@features/apl/hooks/useRenamedCopy';
import { APL_VARIABLES, aplItemId } from '@features/apl/model/apply_move';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import { rotationSource } from '@features/apl/utils';
import { useListItems } from '@features/inputs/hooks/useListItems';
import { APLValueVariable } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';

import { VariableItem } from './VariableItem';

const variableName = () => i18n.t('rotation_tab.apl.variables.name');

/** The rotation's named values, which any condition can reference: one tree item each. */
export const VariablesList = () => {
	const player = usePlayer();
	const variables = useListItems(player, {
		storeSubscribe: rotationSource,
		getValue: (subject: Player<any>) => subject.aplRotation.valueVariables || [],
		setValue: (subject: Player<any>, next: Array<APLValueVariable>) =>
			subject.modifyAplRotation(rotation => {
				rotation.valueVariables = next;
			}),
		newItem: () => APLValueVariable.create({ name: i18n.t('rotation_tab.apl.variables.newVariableName'), value: undefined }),
	});
	const pasting = useAplPaste('variable', variables);
	const copying = useRenamedCopy(player, {
		itemName: variableName(),
		inputLabel: i18n.t('rotation_tab.apl.variables.attributes.name'),
		read: rotation => rotation.valueVariables || [],
		write: (rotation, next) => {
			rotation.valueVariables = next;
		},
		nameOf: variable => variable.name,
		// `create` re-creates nested messages recursively, so the copy's value shares nothing with the original's.
		copy: (variable, name) => freshUuids(APLValueVariable.create({ name, value: variable.value })),
	});

	return (
		// The tab panel is a flex column with a gap; this stays a single flex item so that gap doesn't land between the header and the tree.
		<div className="flex flex-col">
			<h6 className="mb-2 flex items-center gap-2">
				<AplSectionTitle text={i18n.t('rotation_tab.apl.variables.header')} tooltip={i18n.t('rotation_tab.apl.variables.tooltips.overview')} />
			</h6>
			<AplList.Root group={APL_VARIABLES} testId="apl-variables-tree">
				{variables.items.map((variable, index) => {
					const id = aplItemId(variable, index);
					return (
						<VariableItem
							key={id}
							id={id}
							player={player}
							config={variables.itemConfig(variable, index)}
							onCopy={() => copying.onCopyItem(index)}
							onDelete={() => variables.remove(index)}
						/>
					);
				})}
			</AplList.Root>
			<AplListToolbar
				itemName={variableName()}
				nameDialog={{
					inputLabel: i18n.t('rotation_tab.apl.variables.attributes.name'),
					existingNames: () => variables.items.map(variable => variable.name),
				}}
				onCreate={name => variables.append(APLValueVariable.create({ name: name!, value: undefined }))}
				onPaste={pasting.paste}
			/>
			<AplNameDialog {...copying.dialog} />
			<AplNameDialog {...pasting.nameDialog} />
			<AplPasteDialog {...pasting.textDialog} />
		</div>
	);
};
