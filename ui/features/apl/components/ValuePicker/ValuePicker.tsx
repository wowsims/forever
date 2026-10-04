import { AplProvider, useApl } from '@features/apl/context/AplContext';
import { useAplInput } from '@features/apl/hooks/useAplInput';
import { valueListGroup } from '@features/apl/model/apply_move';
import { newKindImpl, resolveField } from '@features/apl/model/field_specs';
import { valueKindOptions } from '@features/apl/model/kind_options';
import { swapValueKind } from '@features/apl/model/kind_swap';
import { pathChoices } from '@features/apl/model/submenus';
import { type APLValueKind, type ValidAPLValueKind, valueKinds } from '@features/apl/model/value_kinds';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import type { InputConfig } from '@features/inputs/input';
import type { APLValue } from '@generated/proto/apl';
import { UUID } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { randomUUID } from '@sim/utils/misc';
import { Field } from '@ui-kit/Field';
import { memo, useId, useMemo } from 'react';

import { FieldGroup } from '../FieldGroup';

export interface ValuePickerProps {
	player: Player<any>;
	config: InputConfig<Player<any>, APLValue | undefined>;
	testId?: string;
}

/**
 * One APL value: a kind, and the fields that kind has.
 *
 * The recursion lives in those fields — a comparison holds two values, an `and` holds a list of
 * them — so this component reaches itself through `FieldGroup` and `AplField`. See `AplField` for
 * why that cycle is safe and what would break it.
 *
 * Nothing here holds a copy of the value: each picker writes its own property of the live message,
 * and the rotation notification is what re-renders.
 */
export const ValuePicker = memo(({ player, config, testId }: ValuePickerProps) => {
	const kindId = useId();
	const { isPrepull, isGroup, changeSource } = useApl();

	const { value, hidden, disabled, frameProps } = useAplInput(player, config);
	const kind = value?.value.oneofKind;

	// A value loaded from a preset or an import may carry no uuid, and the sim keys its per-value
	// validations on one, as `valueListGroup` keys this value's operand list. Minting it is a silent
	// write during render, deliberately without a notification: it changes nothing the user can see,
	// and the list below reads the group in this same render.
	if (value && !value.uuid?.value) value.uuid = UUID.create({ value: randomUUID() });

	// An operand list takes the whole line, so its rows lay out by the row's width, not their own.
	const ownsList = useMemo(() => !!kind && valueKinds[kind].fields.some(field => resolveField(field).kind === 'valueList'), [kind]);

	const choices = useMemo(
		() =>
			pathChoices<APLValueKind>(
				[{ value: undefined, label: i18n.t('rotation_tab.apl.values.none') }, ...valueKindOptions(player, isPrepull, isGroup)],
				Object.is,
			),
		[player, isPrepull, isGroup],
	);

	const kindConfig: InputConfig<Player<any>, APLValueKind> & { id: string } = {
		id: kindId,
		storeSubscribe: changeSource,
		getValue: () => config.getValue(player)?.value.oneofKind,
		setValue: (subject: Player<any>, newKind: APLValueKind) => {
			const source = config.getValue(subject);
			if (source?.value.oneofKind == newKind) return;
			if (newKind) {
				const next = swapValueKind(source, newKind, newKindImpl(valueKinds[newKind]));
				if (source) source.value = next.value;
				else config.setValue(subject, next);
			} else {
				config.setValue(subject, undefined);
			}
			subject.touchRotation();
		},
	};

	const implConfig = (implKind: ValidAPLValueKind): InputConfig<Player<any>, any> => ({
		getValue: () => {
			const source = config.getValue(player);
			return (source && (source.value as any)[implKind]) || valueKinds[implKind].newValue();
		},
		setValue: (subject: Player<any>, newValue: any) => {
			const source = config.getValue(subject);
			if (source) (source.value as any)[implKind] = newValue;
			subject.touchRotation();
		},
	});

	if (hidden) return null;
	return (
		<Field
			{...frameProps}
			className={['m-0 flex-row items-baseline gap-2', ownsList && 'w-full', frameProps.className]}
			testId={testId ?? 'apl-value-picker-root'}
			disabled={disabled}>
			<ConnectedChoiceField<Player<any>, APLValueKind>
				modObject={player}
				config={kindConfig}
				options={choices.options}
				optionTree={choices.optionTree}
				variant="link"
				placeholder={i18n.t('rotation_tab.apl.values.no_condition')}
			/>
			{kind && value && (
				<AplProvider listGroup={valueListGroup(value)}>
					<FieldGroup key={kind} player={player} config={implConfig(kind)} fields={valueKinds[kind].fields} />
				</AplProvider>
			)}
		</Field>
	);
});
