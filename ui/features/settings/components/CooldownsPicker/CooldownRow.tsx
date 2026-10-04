import { ConnectedIconEnumPicker } from '@features/inputs/components/ConnectedIconEnumPicker';
import { ConnectedNumberListField } from '@features/inputs/components/ConnectedTextField';
import type { ActionID as ActionIdProto } from '@generated/proto/common';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useActionId } from '@sim/hooks/useActionId';
import { ActionId } from '@sim/proto/action_id';
import { Button } from '@ui-kit/Button';
import { FieldLabel } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useMemo } from 'react';

import { actionPickerConfig, deleteCooldown, timingsPickerConfig } from './utils';

export interface CooldownRowViewProps {
	name: string | undefined;
	/** The trailing row, which picks a cooldown that does not exist yet. */
	isAdd: boolean;
	deleteTooltipId: string;
	onDelete: () => void;
	actionPicker: ReactNode;
	timingsPicker: ReactNode;
}

export const CooldownRowView = ({ name, isAdd, deleteTooltipId, onDelete, actionPicker, timingsPicker }: CooldownRowViewProps) => (
	<div className="mb-2 flex items-center justify-between [&>*:not(:last-child)]:mr-2" data-testid="cooldown-picker" data-add={isAdd ? '' : undefined}>
		{actionPicker}
		<FieldLabel as="span" className="min-w-[30%] overflow-hidden text-ellipsis" testId="cooldown-picker-label">
			{name}
		</FieldLabel>
		{timingsPicker}
		<Button
			variant="unstyled"
			aria-label={i18n.t('rotation_tab.cooldowns.delete_tooltip')}
			className={clsx('text-link-danger', isAdd && 'invisible')}
			data-testid="delete-cooldown"
			data-hidden={isAdd ? '' : undefined}
			onClick={onDelete}
			{...tooltipAnchorProps(deleteTooltipId)}>
			<Icon name="xmark" style="base" size="xl" />
		</Button>
	</div>
);

export interface CooldownRowProps {
	index: number;
	id: ActionIdProto | undefined;
	available: ReadonlyArray<ActionId>;
	/** The trailing row, which picks a cooldown that does not exist yet. */
	isAdd: boolean;
	deleteTooltipId: string;
}

export const CooldownRow = ({ index, id, available, isAdd, deleteTooltipId }: CooldownRowProps) => {
	const player = usePlayer();
	const actionConfig = useMemo(() => actionPickerConfig(available, index), [available, index]);
	const timingsConfig = useMemo(() => timingsPickerConfig(index, isAdd), [index, isAdd]);
	const actionId = id ? ActionId.fromProto(id) : undefined;
	const { name } = useActionId(actionId);

	return (
		<CooldownRowView
			name={name}
			isAdd={isAdd}
			deleteTooltipId={deleteTooltipId}
			onDelete={() => deleteCooldown(player, index)}
			actionPicker={<ConnectedIconEnumPicker modObject={player} config={actionConfig} />}
			timingsPicker={<ConnectedNumberListField modObject={player} config={timingsConfig} />}
		/>
	);
};
