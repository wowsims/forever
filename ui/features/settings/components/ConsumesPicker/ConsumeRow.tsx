import { iconEnumPickerShown } from '@features/inputs/components/ConnectedIconEnumPicker';
import type { IconEnumPickerConfig, IconPickerConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { subscribePlayerChange } from '@sim/state/subscriptions';
import { Field } from '@ui-kit/Field';
import { FieldLabel } from '@ui-kit/FormControl';
import { type ReactNode, useId } from 'react';

export interface ConsumeRowViewProps {
	label: string;
	hidden: boolean;
	children: ReactNode;
}

// The caption is a <span> that names the row's picker group, since no single control owns it.
export const ConsumeRowView = ({ label, hidden, children }: ConsumeRowViewProps) => {
	const labelId = useId();
	return (
		<Field
			layout="inline"
			role="group"
			aria-labelledby={labelId}
			testId="consumes-row"
			hidden={hidden}
			leading={
				<FieldLabel as="span" id={labelId}>
					{label}
				</FieldLabel>
			}>
			{children}
		</Field>
	);
};

export type ConsumeRowConfig = IconEnumPickerConfig<Player<any>, any> | IconPickerConfig<Player<any>, any>;

const rowConfigShown = (config: ConsumeRowConfig, player: Player<any>): boolean =>
	'values' in config ? iconEnumPickerShown(config, player) : !config.showWhen || config.showWhen(player);

export interface ConsumeRowProps {
	name: 'potions' | 'elixirs' | 'food' | 'holiday' | 'engineering' | 'imbue' | 'weaponImbue' | 'scrolls' | 'miscellaneous';
	configs: ReadonlyArray<ConsumeRowConfig>;
	children: ReactNode;
}

export const ConsumeRow = ({ name, configs, children }: ConsumeRowProps) => {
	const player = usePlayer();
	const shown = useStoreSubscribe(subscribePlayerChange(player), () => configs.some(config => rowConfigShown(config, player)));

	// Hide, never unmount: a hidden ConnectedIconEnumPicker is what zeroes the field it is bound to
	// when the player can no longer use it, and restores the value if they can again. Returning
	// null here skips that effect entirely, so a stale selection survives in the proto.
	return (
		<ConsumeRowView label={i18n.t(`settings_tab.consumables.${name}.title`)} hidden={!shown}>
			{children}
		</ConsumeRowView>
	);
};
