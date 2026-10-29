import { AplItemActions, AplList } from '@features/apl/components/AplList';
import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplValidations } from '@features/apl/components/ListItemHeader';
import { uuidValidations } from '@features/apl/components/ListItemHeader/utils';
import { NameDisplay } from '@features/apl/components/NameDisplay';
import { ValuePicker } from '@features/apl/components/ValuePicker';
import { AplProvider, useApl } from '@features/apl/context/AplContext';
import { copyAplClip } from '@features/apl/hooks/useAplClipboard';
import { useAplInput } from '@features/apl/hooks/useAplInput';
import { rowSource } from '@features/apl/model/row_source';
import type { InputConfig } from '@features/inputs/input';
import type { APLValue, APLValueVariable } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { renameAPLReference } from '@sim/proto/apl_utils';
import { Field } from '@ui-kit/Field';
import { Tree } from '@ui-kit/Tree';
import { useMemo, useRef, useState } from 'react';

export interface VariableItemProps {
	id: string;
	player: Player<any>;
	config: InputConfig<Player<any>, APLValueVariable>;
	onCopy: () => void;
	onDelete: () => void;
}

/** One named value variable: its name and validations in the title, and the value it stands for as the body. */
export const VariableItem = ({ id, player, config, onCopy, onDelete }: VariableItemProps) => {
	const { changeSource } = useApl();
	const { value: variable, disabled, frameProps } = useAplInput(player, config);
	const [renaming, setRenaming] = useState(false);

	const latest = useRef(config);
	latest.current = config;

	const label = i18n.t('rotation_tab.apl.variables.attributes.value');
	const labelTooltip = i18n.t('rotation_tab.apl.variables.attributes.valueTooltip');
	const row = useMemo(() => {
		const source = rowSource(changeSource(player), () => latest.current.getValue(player));
		const value: InputConfig<Player<any>, APLValue | undefined> = {
			label,
			labelTooltip,
			getValue: () => latest.current.getValue(player)?.value,
			setValue: (subject: Player<any>, newValue: APLValue | undefined) => {
				const current = latest.current.getValue(subject);
				if (!current) return;
				current.value = newValue;
				// The variable list owns the array; writing through it is what notifies.
				latest.current.setValue(subject, current);
			},
		};
		return { changeSource: () => source, value };
	}, [player, changeSource, label, labelTooltip]);
	const variableLabel = i18n.t('rotation_tab.apl.variables.name');

	return (
		<AplList.Item
			id={id}
			label={variableLabel}
			title={<NameDisplay name={variable?.name || ''} onRename={() => setRenaming(true)} />}
			aside={<AplValidations getValidations={subject => uuidValidations(subject, config.getValue(subject)?.value?.uuid?.value)} />}
			actions={
				<AplItemActions
					label={variableLabel}
					onCopy={onCopy}
					onCopyToClipboard={() => copyAplClip({ kind: 'variable', item: config.getValue(player) })}
					onDelete={onDelete}
				/>
			}
			testId="apl-variable-root">
			<Tree.Item.Body>
				<Field {...frameProps} size="xs" className={[frameProps.className, 'm-0']} testId="apl-list-item-picker-root" disabled={disabled}>
					<AplProvider changeSource={row.changeSource}>
						<ValuePicker player={player} config={row.value} />
					</AplProvider>
				</Field>
			</Tree.Item.Body>
			<AplNameDialog
				open={renaming}
				title={i18n.t('rotation_tab.apl.nameModal.rename', { itemName: variableLabel })}
				inputLabel={i18n.t('rotation_tab.apl.variables.attributes.name')}
				confirmLabel={i18n.t('rotation_tab.apl.nameModal.renameConfirm')}
				defaultValue={variable?.name}
				existingNames={(player.aplRotation.valueVariables || []).filter(other => other !== variable).map(other => other.name)}
				onSubmit={name =>
					player.modifyAplRotation(rotation => {
						if (!variable) return;
						renameAPLReference(rotation, { type: 'variable', oldName: variable.name, newName: name });
						variable.name = name;
					})
				}
				onClose={() => setRenaming(false)}
			/>
		</AplList.Item>
	);
};
