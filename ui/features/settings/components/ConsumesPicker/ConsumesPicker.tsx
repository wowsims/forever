import { ConnectedIconEnumPicker } from '@features/inputs/components/ConnectedIconEnumPicker';
import { ConnectedIconPicker } from '@features/inputs/components/ConnectedIconPicker';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { Database } from '@sim/proto/database';
import { IconRow } from '@ui-kit/IconRow';
import { useMemo } from 'react';

import { ConsumeRow } from './ConsumeRow';
import { type ConsumeConfig, consumeConfigs, type ConsumeToggle } from './utils';

export interface ConsumesPickerProps {
	// Potions, explosives and the combat-only miscellany matter inside an encounter; a gear planner
	// never runs one and passes false.
	encounterConsumes?: boolean;
}

const Toggles = ({ player, toggles }: { player: Player<any>; toggles: ReadonlyArray<ConsumeToggle> }) =>
	toggles.map(config => <ConnectedIconPicker key={config.actionId.itemId} modObject={player} config={config} />);

const Pickers = ({ player, configs }: { player: Player<any>; configs: ReadonlyArray<ConsumeConfig> }) =>
	configs.map((config, i) => <ConnectedIconEnumPicker key={i} modObject={player} config={config} />);

export const ConsumesPicker = ({ encounterConsumes = true }: ConsumesPickerProps) => {
	const player = usePlayer() as Player<any>;
	const configs = useMemo(() => consumeConfigs(player, Database.getSync()), [player]);

	return (
		<div className="grid gap-2 max-lg:grid-cols-3 max-md:grid-cols-1">
			{encounterConsumes && (
				<ConsumeRow name="potions" configs={[configs.potion, configs.conjured]}>
					<IconRow className="justify-end" data-testid="consumes-potions">
						<ConnectedIconEnumPicker modObject={player} config={configs.potion} />
						<ConnectedIconEnumPicker modObject={player} config={configs.conjured} />
					</IconRow>
				</ConsumeRow>
			)}
			<ConsumeRow name="elixirs" configs={[configs.flask, ...configs.elixirs]}>
				<IconRow className="justify-end">
					<ConnectedIconEnumPicker modObject={player} config={configs.flask} testId="consumes-flasks" />
					<Pickers player={player} configs={configs.elixirs} />
				</IconRow>
			</ConsumeRow>
			<ConsumeRow name="scrolls" configs={configs.scrolls}>
				<IconRow className="justify-end" data-testid="consumes-scrolls">
					<Pickers player={player} configs={configs.scrolls} />
				</IconRow>
			</ConsumeRow>
			<ConsumeRow name="food" configs={[configs.food, ...configs.groups]}>
				<IconRow className="justify-end" data-testid="consumes-food">
					<ConnectedIconEnumPicker modObject={player} config={configs.food} />
					<Pickers player={player} configs={configs.groups} />
				</IconRow>
			</ConsumeRow>
			{encounterConsumes && (
				<ConsumeRow name="engineering" configs={[configs.explosive, ...configs.explosiveExtras]}>
					<IconRow className="justify-end" data-testid="consumes-engi">
						<ConnectedIconEnumPicker modObject={player} config={configs.explosive} />
						<Toggles player={player} toggles={configs.explosiveExtras} />
					</IconRow>
				</ConsumeRow>
			)}
			<ConsumeRow name="imbue" configs={[configs.mhImbue, configs.ohImbue]}>
				<IconRow className="justify-end" data-testid="consumes-imbue">
					<ConnectedIconEnumPicker modObject={player} config={configs.mhImbue} />
					<ConnectedIconEnumPicker modObject={player} config={configs.ohImbue} />
				</IconRow>
			</ConsumeRow>
			<ConsumeRow name="weaponImbue" configs={[configs.mhWeaponImbue, configs.ohWeaponImbue]}>
				<IconRow className="justify-end" data-testid="consumes-weapon-imbue">
					<ConnectedIconEnumPicker modObject={player} config={configs.mhWeaponImbue} />
					<ConnectedIconEnumPicker modObject={player} config={configs.ohWeaponImbue} />
				</IconRow>
			</ConsumeRow>
			<ConsumeRow name="holiday" configs={[...configs.holidayGroups, ...configs.holiday]}>
				<IconRow className="justify-end" data-testid="consumes-holiday">
					<Pickers player={player} configs={configs.holidayGroups} />
					<Toggles player={player} toggles={configs.holiday} />
				</IconRow>
			</ConsumeRow>
			{encounterConsumes && (
				<ConsumeRow name="miscellaneous" configs={configs.miscellaneous}>
					<IconRow className="justify-end" data-testid="consumes-misc">
						<Toggles player={player} toggles={configs.miscellaneous} />
					</IconRow>
				</ConsumeRow>
			)}
		</div>
	);
};
