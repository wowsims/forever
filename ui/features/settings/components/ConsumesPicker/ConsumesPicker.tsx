import { ConnectedIconEnumPicker } from '@features/inputs/components/ConnectedIconEnumPicker';
import { ConnectedIconPicker } from '@features/inputs/components/ConnectedIconPicker';
import { ConnectedMultiIconPicker } from '@features/inputs/components/ConnectedMultiIconPicker';
import type { MultiIconPickerConfig } from '@features/inputs/configs';
import { CONSUMABLE_PICKER_FIELDS, consumableColumns } from '@features/settings/model/consumables';
import { clearMultiIconInputs } from '@features/settings/model/multi_icon';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { useStoreField } from '@sim/hooks/useStoreField';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { Database } from '@sim/proto/database';
import { CONSUMABLE_ORDER_FIELDS, consumableOrderOf, orderedByEp } from '@sim/settings/consumables';
import { IconEnumPickerDirection } from '@ui-kit/IconEnumPicker';
import { IconRow } from '@ui-kit/IconRow';
import { useMemo } from 'react';

import { ConsumeRow } from './ConsumeRow';
import { type ConsumeConfig, consumeConfigs, type ConsumeToggle } from './utils';

export interface ConsumesPickerProps {
	// Potions, explosives and the combat-only miscellany matter inside an encounter; a gear planner
	// never runs one and passes false.
	encounterConsumes?: boolean;
}

const useOrderedToggles = (player: Player<any>, toggles: ReadonlyArray<ConsumeToggle>) => {
	const order = useStoreSubscribe(useStoreField(CONSUMABLE_ORDER_FIELDS)!, () => consumableOrderOf(player));
	return useMemo(() => orderedByEp(toggles, toggle => toggle.consumable, order), [toggles, order]);
};

const Toggles = ({ player, toggles }: { player: Player<any>; toggles: ReadonlyArray<ConsumeToggle> }) =>
	useOrderedToggles(player, toggles).map(config => <ConnectedIconPicker key={config.actionId.itemId} modObject={player} config={config} />);

// The toggles behind one icon, which opens below on a menu of them, as the pickers beside it do.
interface MultiToggleProps {
	player: Player<any>;
	toggles: ReadonlyArray<ConsumeToggle>;
	tooltip: string;
	testId: string;
}

const MultiToggle = ({ player, toggles, tooltip, testId }: MultiToggleProps) => {
	const subscribe = useStoreField(CONSUMABLE_PICKER_FIELDS)!;
	const ordered = useOrderedToggles(player, toggles);
	const config = useMemo<MultiIconPickerConfig<Player<any>>>(() => {
		const offered = (p: Player<any>) => ordered.filter(toggle => !toggle.showWhen || toggle.showWhen(p)).length;
		return {
			inputs: ordered,
			tooltip,
			direction: IconEnumPickerDirection.Vertical,
			numColumns: p => consumableColumns(offered(p)),
			showWhen: p => offered(p) > 0,
		};
	}, [ordered, tooltip]);
	return (
		<ConnectedMultiIconPicker
			modObject={player}
			config={config}
			subscribe={subscribe}
			onClear={() => clearMultiIconInputs(player, config)}
			testId={testId}
		/>
	);
};

const Pickers = ({ player, configs }: { player: Player<any>; configs: ReadonlyArray<ConsumeConfig> }) =>
	configs.map((config, i) => <ConnectedIconEnumPicker key={i} modObject={player} config={config} />);

export interface ImbueRowsProps {
	player: Player<any>;
	/** The main hand's and the off hand's pick. */
	imbues: [ConsumeConfig, ConsumeConfig];
	weaponImbues: [ConsumeConfig, ConsumeConfig];
	imbueTestId: string;
	weaponImbueTestId: string;
}

export const ImbueRows = ({ player, imbues, weaponImbues, imbueTestId, weaponImbueTestId }: ImbueRowsProps) => (
	<>
		<ConsumeRow name="imbue" configs={imbues}>
			<IconRow className="justify-end" data-testid={imbueTestId}>
				<Pickers player={player} configs={imbues} />
			</IconRow>
		</ConsumeRow>
		<ConsumeRow name="weaponImbue" configs={weaponImbues}>
			<IconRow className="justify-end" data-testid={weaponImbueTestId}>
				<Pickers player={player} configs={weaponImbues} />
			</IconRow>
		</ConsumeRow>
	</>
);

export const ConsumesPicker = ({ encounterConsumes = true }: ConsumesPickerProps) => {
	const player = usePlayer() as Player<any>;
	const configs = useMemo(() => consumeConfigs(player, Database.getSync()), [player]);

	return (
		<div className="grid gap-2 max-lg:grid-cols-3 max-md:grid-cols-1">
			{encounterConsumes && (
				<ConsumeRow name="potions" configs={[configs.offensivePotion, configs.defensivePotion, configs.resourcePotion, configs.conjured]}>
					<IconRow className="justify-end" data-testid="consumes-potions">
						<ConnectedIconEnumPicker modObject={player} config={configs.offensivePotion} />
						<ConnectedIconEnumPicker modObject={player} config={configs.defensivePotion} />
						<ConnectedIconEnumPicker modObject={player} config={configs.resourcePotion} />
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
			<ConsumeRow name="food" configs={[configs.food, ...(configs.dragonbreathChili ? [configs.dragonbreathChili] : []), ...configs.foodGroups]}>
				<IconRow className="justify-end" data-testid="consumes-food">
					<ConnectedIconEnumPicker modObject={player} config={configs.food} />
					{configs.dragonbreathChili && <ConnectedIconPicker modObject={player} config={configs.dragonbreathChili} />}
					<Pickers player={player} configs={configs.foodGroups} />
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
			<ImbueRows
				player={player}
				imbues={[configs.mhImbue, configs.ohImbue]}
				weaponImbues={[configs.mhWeaponImbue, configs.ohWeaponImbue]}
				imbueTestId="consumes-imbue"
				weaponImbueTestId="consumes-weapon-imbue"
			/>
			<ConsumeRow name="holiday" configs={[...configs.holidayGroups, ...configs.holiday]}>
				<IconRow className="justify-end" data-testid="consumes-holiday">
					<Pickers player={player} configs={configs.holidayGroups} />
					<Toggles player={player} toggles={configs.holiday} />
				</IconRow>
			</ConsumeRow>
			<ConsumeRow
				name="miscellaneous"
				configs={[...configs.miscGroups, ...(encounterConsumes ? [...configs.miscOffensive, ...configs.miscDefensive] : [])]}>
				<IconRow className="justify-end" data-testid="consumes-misc">
					<Pickers player={player} configs={configs.miscGroups} />
					{encounterConsumes && (
						<>
							<MultiToggle
								player={player}
								toggles={configs.miscOffensive}
								tooltip={i18n.t('settings_tab.consumables.miscellaneous.offensive')}
								testId="consumes-misc-offensive"
							/>
							<MultiToggle
								player={player}
								toggles={configs.miscDefensive}
								tooltip={i18n.t('settings_tab.consumables.miscellaneous.defensive')}
								testId="consumes-misc-defensive"
							/>
						</>
					)}
				</IconRow>
			</ConsumeRow>
		</div>
	);
};
