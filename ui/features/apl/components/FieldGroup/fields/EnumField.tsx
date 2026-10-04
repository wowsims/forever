import type { EnumTable } from '@features/apl/model/field_specs';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import type { InputConfig } from '@features/inputs/input';
import type { Player } from '@sim/player/player';

export interface EnumFieldProps {
	player: Player<any>;
	config: InputConfig<Player<any>, number> & { id: string };
	table: EnumTable;
}

/** A field whose values are a fixed proto enum. The lists themselves are data, in `model/field_specs.ts`. */
export const EnumField = ({ player, config, table }: EnumFieldProps) => (
	<ConnectedChoiceField<Player<any>, number> modObject={player} config={config} options={table.options} variant="link" placeholder={table.defaultLabel} />
);
