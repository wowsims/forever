import { AplProvider, useApl } from '@features/apl/context/AplContext';
import { useAplInput } from '@features/apl/hooks/useAplInput';
import { actionKinds, type APLActionKind } from '@features/apl/model/action_kinds';
import { actionListGroup } from '@features/apl/model/apply_move';
import { newKindImpl } from '@features/apl/model/field_specs';
import { actionKindOptions } from '@features/apl/model/kind_options';
import { swapActionKind } from '@features/apl/model/kind_swap';
import { pathChoices } from '@features/apl/model/submenus';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import type { InputConfig } from '@features/inputs/input';
import { APLAction, type APLValue } from '@generated/proto/apl';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Field } from '@ui-kit/Field';
import clsx from 'clsx';
import { memo, useId, useMemo } from 'react';

import { FieldGroup } from '../FieldGroup';
import { ValuePicker } from '../ValuePicker';

export interface ActionPickerProps {
	player: Player<any>;
	config: InputConfig<Player<any>, APLAction>;
	stacked?: boolean;
	fullWidth?: boolean;
}

type ValidAPLActionKind = NonNullable<APLActionKind>;

const SEQUENCE_KINDS: ReadonlySet<ValidAPLActionKind> = new Set(['sequence', 'strictSequence', 'schedule']);

/**
 * One APL action: an optional condition, a kind, and that kind's fields.
 *
 * The condition is a whole `ValuePicker`, and a sequence's field list holds more `ActionPicker`s —
 * so this is the action half of the same cycle `AplField` documents. An action in the pre-pull list,
 * or inside a sequence, a strict sequence or a schedule, has no condition picker.
 *
 * The kind type is the non-nullable one: the option list is built from the kind table alone, with
 * no empty entry — unlike the value picker's — so "no kind" is not a selection the menu can make.
 *
 * Memoised because a list row hands it a config that lives as long as the row, so a rotation change
 * that leaves the row's content alone stops here instead of re-rendering the row's whole tree.
 */
export const ActionPicker = memo(({ player, config, stacked, fullWidth }: ActionPickerProps) => {
	const kindId = useId();
	const { isPrepull, inSequence, changeSource } = useApl();

	const { value, hidden, disabled, frameProps } = useAplInput(player, config);
	const kind = value?.action.oneofKind;

	const choices = useMemo(() => pathChoices(actionKindOptions(player, isPrepull), Object.is), [player, isPrepull]);

	const conditionConfig: InputConfig<Player<any>, APLValue | undefined> = {
		label: i18n.t('rotation_tab.apl.priority_list.if_label'),
		getValue: () => config.getValue(player)?.condition,
		setValue: (subject: Player<any>, newValue: APLValue | undefined) => {
			const source = config.getValue(subject);
			if (source) {
				source.condition = newValue;
				subject.touchRotation();
			} else {
				config.setValue(subject, APLAction.create({ condition: newValue }));
			}
		},
	};

	const kindConfig: InputConfig<Player<any>, ValidAPLActionKind> & { id: string } = {
		id: kindId,
		storeSubscribe: changeSource,
		getValue: () => config.getValue(player)?.action.oneofKind as ValidAPLActionKind,
		setValue: (subject: Player<any>, newKind: ValidAPLActionKind) => {
			const source = config.getValue(subject);
			if (source?.action.oneofKind == newKind) return;
			const next = swapActionKind(source, newKind, newKindImpl(actionKinds[newKind]));
			if (source) source.action = next.action;
			else config.setValue(subject, next);
			subject.touchRotation();
		},
	};

	const implConfig = (implKind: ValidAPLActionKind): InputConfig<Player<any>, any> => ({
		getValue: () => (config.getValue(player)?.action as any)?.[implKind] || actionKinds[implKind].newValue(),
		setValue: (subject: Player<any>, newValue: any) => {
			const source = config.getValue(subject);
			if (source) (source.action as any)[implKind] = newValue;
			subject.touchRotation();
		},
	});

	if (hidden) return null;
	return (
		<Field
			{...frameProps}
			className={[clsx('m-0 min-w-0 grow gap-2', stacked ? 'flex-col' : 'flex-row items-baseline', fullWidth && 'w-full'), frameProps.className]}
			testId="apl-action-picker-root"
			disabled={disabled}>
			{!isPrepull && !inSequence && <ValuePicker player={player} config={conditionConfig} testId="apl-action-condition" />}
			<div className="flex w-full flex-row flex-wrap items-baseline gap-2" data-kind={kind}>
				<ConnectedChoiceField<Player<any>, ValidAPLActionKind>
					modObject={player}
					config={kindConfig}
					options={choices.options}
					optionTree={choices.optionTree}
					variant="link"
					placeholder={i18n.t('rotation_tab.apl.priority_list.item_label')}
				/>
				{kind && value && (
					<AplProvider inSequence={SEQUENCE_KINDS.has(kind)} listGroup={actionListGroup(value)}>
						<FieldGroup key={kind} player={player} config={implConfig(kind)} fields={actionKinds[kind].fields} />
					</AplProvider>
				)}
			</div>
		</Field>
	);
});
