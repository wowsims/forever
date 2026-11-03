import { ActionPicker } from '@features/apl/components/ActionPicker';
import { AplProvider, useApl } from '@features/apl/context/AplContext';
import { useAplInput } from '@features/apl/hooks/useAplInput';
import { rowSource } from '@features/apl/model/row_source';
import type { InputConfig } from '@features/inputs/input';
import type { APLAction } from '@generated/proto/apl';
import type { Player } from '@sim/player/player';
import { Field } from '@ui-kit/Field';
import { type ReactNode, useMemo, useRef } from 'react';

/** The shape the three action lists share: a row that can be hidden and holds one action. */
export interface HideableAction {
	hide: boolean;
	action?: APLAction;
}

export interface AplListItemProps<T extends HideableAction> {
	player: Player<any>;
	config: InputConfig<Player<any>, T>;
	leading?: ReactNode;
	actionFullWidth?: boolean;
}

/**
 * One row of the priority list, a group's action list or the pre-pull list.
 *
 * A hidden row renders `disabled` rather than being unmounted — that is all `hide` means here.
 */
export const AplListItem = <T extends HideableAction>({ player, config, leading, actionFullWidth }: AplListItemProps<T>) => {
	const { changeSource } = useApl();
	const { hidden, disabled, frameProps } = useAplInput(player, { ...config, enableWhen: () => !config.getValue(player)?.hide });

	const item = useRef(() => config.getValue(player));
	item.current = () => config.getValue(player);

	const row = useMemo(() => {
		const source = rowSource(changeSource(player), () => item.current());
		const action: InputConfig<Player<any>, APLAction> = {
			getValue: () => item.current().action!,
			setValue: (subject: Player<any>, newValue: APLAction) => {
				const current = item.current();
				if (!current) return;
				current.action = newValue;
				subject.touchRotation();
			},
		};
		return { changeSource: () => source, action };
	}, [player, changeSource]);

	return (
		<AplProvider changeSource={row.changeSource}>
			{hidden ? null : (
				<Field
					{...frameProps}
					size="xs"
					className={['m-0 flex flex-row items-baseline gap-2', frameProps.className]}
					testId="apl-list-item-picker-root"
					disabled={disabled}>
					{leading}
					<ActionPicker player={player} config={row.action} stacked fullWidth={actionFullWidth} />
				</Field>
			)}
		</AplProvider>
	);
};
