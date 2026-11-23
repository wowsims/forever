import { iconEnumPickerShown } from '@features/inputs/components/ConnectedIconEnumPicker';
import type { IconEnumPickerConfig, IconPickerConfig } from '@features/inputs/configs';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreField } from '@sim/hooks/useStoreField';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { Field } from '@ui-kit/Field';
import { FieldLabel } from '@ui-kit/FormControl';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { CONSUMABLE_PICKER_FIELDS } from '../../model/consumables';

export interface ConsumeRowViewProps {
	label: string;
	hidden: boolean;
	/** Keeps the caption on one line, the pickers wrapping instead. */
	nowrap?: boolean;
	children: ReactNode;
}

// The caption is a <span> that names the row's picker group, since no single control owns it. It stands as tall as
// one line of swatches and the row aligns to its top, so the caption stays level with the first line when the
// swatches wrap.
export const ConsumeRowView = ({ label, hidden, nowrap, children }: ConsumeRowViewProps) => {
	const labelId = useId();
	return (
		<Field
			layout="inline"
			className="items-start"
			role="group"
			aria-labelledby={labelId}
			testId="consumes-row"
			hidden={hidden}
			leading={
				<FieldLabel as="span" id={labelId} className={clsx('flex min-h-10 items-center', nowrap && 'whitespace-nowrap')}>
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
	const shown = useStoreSubscribe(useStoreField(CONSUMABLE_PICKER_FIELDS)!, () => configs.some(config => rowConfigShown(config, player)));

	// Hide, never unmount: a hidden ConnectedIconEnumPicker is what zeroes the field it is bound to
	// when the player can no longer use it, and restores the value if they can again. Returning
	// null here skips that effect entirely, so a stale selection survives in the proto.
	return (
		<ConsumeRowView label={i18n.t(`settings_tab.consumables.${name}.title`)} hidden={!shown} nowrap={name === 'elixirs'}>
			{children}
		</ConsumeRowView>
	);
};
