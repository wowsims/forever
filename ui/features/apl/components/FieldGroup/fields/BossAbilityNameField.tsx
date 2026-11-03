import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedStringField } from '@features/inputs/components/ConnectedTextField';
import type { InputConfig } from '@features/inputs/input';
import i18n from '@i18n/config';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Player } from '@sim/player/player';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { useMemo } from 'react';

export interface BossAbilityNameFieldProps {
	player: Player<any>;
	config: InputConfig<Player<any>, string> & { id: string };
}

/**
 * A dropdown over the names of the encounter design's boss abilities, each once.
 *
 * The sim finds an ability by its exact name and an unnamed one by none, so a design without a named ability gets a
 * text field instead, for a name the design is still to give.
 */
export const BossAbilityNameField = ({ player, config }: BossAbilityNameFieldProps) => {
	const encounter = player.sim.encounter;
	const names = useStoreSubscribe(subscribeEncounterField(encounter, 'design'), () => [
		...new Set(encounter.getDesign().bossAbilities.flatMap(ability => (ability.name ? [ability.name] : []))),
	]);
	const options = useMemo(() => names.map(name => ({ value: name, label: name })), [names]);

	if (!names.length) return <ConnectedStringField modObject={player} config={config} />;
	return (
		<ConnectedChoiceField<Player<any>, string>
			modObject={player}
			config={config}
			options={options}
			variant="link"
			placeholder={i18n.t('rotation_tab.apl.helpers.select_boss_ability')}
		/>
	);
};
