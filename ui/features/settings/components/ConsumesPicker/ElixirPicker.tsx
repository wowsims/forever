import { ConnectedIconPicker } from '@features/inputs/components/ConnectedIconPicker';
import { CONSUMABLE_PICKER_FIELDS, elixirGroups } from '@features/settings/model/consumables';
import i18n from '@i18n/config';
import { useStoreField } from '@sim/hooks/useStoreField';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { Field } from '@ui-kit/Field';
import { FieldLabel } from '@ui-kit/FormControl';
import { Icon } from '@ui-kit/Icon';
import { IconRow } from '@ui-kit/IconRow';
import { Popover } from '@ui-kit/Popover';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { type ReactNode, useId, useMemo } from 'react';

import { useOrderedToggles } from './useOrderedToggles';
import type { ConsumeToggle } from './utils';

export interface ElixirPickerGroup {
	key: string;
	label: string;
	children: ReactNode;
}

export interface ElixirPickerViewProps {
	/** The picked elixirs' swatches, each unpicked by a click. */
	picks: ReactNode;
	/** The add menu's lines, one per stat, each holding the elixirs offered under it. */
	groups: ReadonlyArray<ElixirPickerGroup>;
}

const GroupLine = ({ label, children }: Omit<ElixirPickerGroup, 'key'>) => {
	const labelId = useId();
	return (
		<Field
			layout="inline"
			role="group"
			aria-labelledby={labelId}
			testId="consumes-elixir-group"
			leading={
				<FieldLabel as="span" id={labelId}>
					{label}
				</FieldLabel>
			}>
			<IconRow className="justify-end">{children}</IconRow>
		</Field>
	);
};

export const ElixirPickerView = ({ picks, groups }: ElixirPickerViewProps) => {
	const tooltipId = useId();
	const add = i18n.t('settings_tab.consumables.elixirs.add');
	return (
		<>
			{picks}
			<Popover
				trigger={<Icon name="plus" />}
				triggerClassName="flex size-10 min-w-10 items-center justify-center border border-border bg-grey text-white hover:border-success"
				triggerTestId="consumes-elixirs-add"
				triggerProps={{ 'aria-label': add, ...tooltipAnchorProps(tooltipId) }}
				side="bottom"
				align="end"
				title={add}
				className="w-96"
				testId="consumes-elixirs-menu">
				<div className="grid gap-2">
					{groups.map(group => (
						<GroupLine key={group.key} label={group.label}>
							{group.children}
						</GroupLine>
					))}
				</div>
			</Popover>
			<Tooltip id={tooltipId} content={add} />
		</>
	);
};

export interface ElixirPickerProps {
	player: Player<any>;
	toggles: ReadonlyArray<ConsumeToggle>;
}

// The picked elixirs, and the menu that adds and removes them. A picked one stays in the row whether the menu
// offers it or not, so its picker can zero it once the player may no longer use it.
export const ElixirPicker = ({ player, toggles }: ElixirPickerProps) => {
	const ordered = useOrderedToggles(player, toggles);
	const groups = useMemo(() => elixirGroups(ordered, toggle => toggle.consumable), [ordered]);
	const offered = (toggle: ConsumeToggle) => !toggle.showWhen || toggle.showWhen(player);
	// Renders again whenever a pick or an offer changes; the reads below take the configs from the render.
	useStoreSubscribe(useStoreField(CONSUMABLE_PICKER_FIELDS)!, () =>
		ordered.map(toggle => `${Number(toggle.getValue(player))}${Number(offered(toggle))}`).join(''),
	);

	return (
		<ElixirPickerView
			picks={ordered
				.filter(toggle => toggle.getValue(player))
				.map(toggle => (
					<ConnectedIconPicker key={toggle.consumable.id} modObject={player} config={toggle} />
				))}
			groups={groups.flatMap(group => {
				const shown = group.items.filter(offered);
				if (!shown.length) return [];
				return [
					{
						key: group.key,
						label: group.label,
						children: shown.map(toggle => <ConnectedIconPicker key={toggle.consumable.id} modObject={player} config={toggle} />),
					},
				];
			})}
		/>
	);
};
