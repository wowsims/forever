import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedIconInput } from '@features/inputs/components/ConnectedIconInput';
import { ConnectedInput } from '@features/inputs/components/ConnectedInput';
import type { IconInputConfig } from '@features/inputs/icon_inputs';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import type { InputConfig } from '@sim/spec_config';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { IconRow } from '@ui-kit/IconRow';
import { type ReactNode, useMemo } from 'react';

import { professionInput, raceInput } from './utils';

export interface PlayerSettingsViewProps {
	iconPickers: ReadonlyArray<ReactNode>;
	race: ReactNode;
	inputs: ReadonlyArray<ReactNode>;
	professions: ReadonlyArray<ReactNode>;
}

export const PlayerSettingsView = ({ iconPickers, race, inputs, professions }: PlayerSettingsViewProps) => (
	<>
		{iconPickers.length > 0 && (
			<IconRow className="ui-player-icon-group" data-testid="player-icon-group">
				{iconPickers}
			</IconRow>
		)}
		{race}
		{inputs}
		<FieldGrid>{professions}</FieldGrid>
	</>
);

export interface PlayerSettingsProps {
	iconInputs: ReadonlyArray<IconInputConfig<Player<any>, any>>;
	inputs: ReadonlyArray<InputConfig<Player<any>>>;
}

export const PlayerSettings = ({ iconInputs, inputs }: PlayerSettingsProps) => {
	const player = usePlayer() as Player<any>;
	const race = useMemo(() => raceInput(player), [player]);
	const professions = useMemo(() => [professionInput(1), professionInput(2)] as const, []);

	return (
		<PlayerSettingsView
			iconPickers={iconInputs.map((config, index) => (
				<ConnectedIconInput key={index} modObject={player} config={config} />
			))}
			race={<ConnectedChoiceField modObject={player} config={race} />}
			inputs={inputs.map(config => (
				<ConnectedInput key={config.id} config={config} />
			))}
			professions={professions.map(config => (
				<ConnectedChoiceField key={config.id} modObject={player} config={config} />
			))}
		/>
	);
};
