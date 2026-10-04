import { ConnectedIconInput } from '@features/inputs/components/ConnectedIconInput';
import type { IconInputConfig } from '@features/inputs/icon_inputs';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { IconRow } from '@ui-kit/IconRow';

export interface RotationIconGroupProps {
	inputs: ReadonlyArray<IconInputConfig<Player<any>, any>>;
}

/**
 * The rotation tab's icon row.
 *
 * `configureIconSection`'s `hide` branch is dropped: it fired only for an empty picker list, and
 * the one caller built the list from a `rotationIconInputs` array it had already checked was
 * non-empty, so that branch could never run.
 */
export const RotationIconGroup = ({ inputs }: RotationIconGroupProps) => {
	const player = usePlayer();

	return (
		<IconRow data-testid="rotation-icon-group">
			{inputs.map((config, index) => (
				<ConnectedIconInput key={index} modObject={player} config={config} />
			))}
		</IconRow>
	);
};
