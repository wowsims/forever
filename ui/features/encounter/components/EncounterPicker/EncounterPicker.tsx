import { EncounterPlanner } from '@features/encounter-planner';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { Encounter as EncounterProto } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { Icon } from '@ui-kit/Icon';
import { type ReactNode, useMemo, useState } from 'react';

import { useEncounterChoice } from '../../hooks/useEncounterChoice';
import { newEncounterProto } from '../../model/my_encounters';
import { AreaTypesPicker } from './AreaTypesPicker';
import { durationConfigs, executeConfigs, minBaseDamageConfig, numAlliesConfig, zoneConfig } from './utils/configs';

export interface EncounterPickerViewProps {
	duration: ReactNode;
	execute?: ReactNode;
	zone: ReactNode;
	areaTypes: ReactNode;
	preset?: ReactNode;
	allies?: ReactNode;
	minBaseDamage?: ReactNode;
	onPlannerClick: () => void;
	/** Shows the + beside the title, which starts a new encounter of the player's own. */
	onNewClick?: () => void;
}

export const EncounterPickerView = ({
	duration,
	execute,
	zone,
	areaTypes,
	preset,
	allies,
	minBaseDamage,
	onPlannerClick,
	onNewClick,
}: EncounterPickerViewProps) => (
	<ContentBlock
		rootDataAttributes={{ 'data-block': 'encounter-settings' }}
		config={{ header: { title: i18n.t('settings_tab.encounter.title') } }}
		headerChildren={
			onNewClick && (
				<Button iconOnly className="ml-auto" aria-label={i18n.t('settings_tab.encounter.new_encounter')} onClick={onNewClick}>
					<Icon name="plus" />
				</Button>
			)
		}>
		<div className="flex flex-wrap gap-2" data-testid="encounter-picker-root">
			{preset}
			<Button className="w-full" data-testid="encounter-planner-button" onClick={onPlannerClick}>
				{i18n.t('settings_tab.encounter.planner.open')}
			</Button>
			<FieldGrid>{duration}</FieldGrid>
			{execute && <FieldGrid className="w-full">{execute}</FieldGrid>}
			{zone}
			{areaTypes}
			{allies}
			{minBaseDamage}
		</div>
	</ContentBlock>
);

export interface EncounterPickerProps {
	showExecuteProportion: boolean;
}

const CHOICE_FIELD = { id: 'encounter-preset-encouter', extraClassNames: ['w-full'] };

export const EncounterPicker = ({ showExecuteProportion }: EncounterPickerProps) => {
	const host = useSimHost();
	const encounter = host.sim.encounter;
	const player = host.player;
	const { damage: showDamage } = useDisplayMetrics(host.sim);
	const choice = useEncounterChoice(encounter, CHOICE_FIELD);
	// A fight with a plan has its execute proportions set in the planner.
	const plannerFight = useStoreSubscribe(subscribeEncounterField(encounter, 'plan'), () => encounter.hasPlan());

	const duration = useMemo(() => durationConfigs(), []);
	const execute = useMemo(() => (showExecuteProportion ? executeConfigs() : []), [showExecuteProportion]);
	const allies = useMemo(() => numAlliesConfig(player), [player]);
	const minBaseDamage = useMemo(() => minBaseDamageConfig(), []);
	const zone = useMemo(() => zoneConfig(encounter), [encounter]);

	const [plannerOpen, setPlannerOpen] = useState(false);
	const [replaced, setReplaced] = useState<EncounterProto>();

	const startNew = () => {
		setReplaced(encounter.toProto());
		encounter.fromProto(newEncounterProto());
		setPlannerOpen(true);
	};
	const onPlannerOpenChange = (open: boolean) => {
		setPlannerOpen(open);
		if (!open) setReplaced(undefined);
	};

	return (
		<>
			<EncounterPickerView
				duration={duration.map(config => (
					<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
				))}
				execute={
					showExecuteProportion &&
					!plannerFight &&
					execute.map(config => <ConnectedNumberField key={config.id} modObject={encounter} config={config} />)
				}
				zone={<ConnectedChoiceField modObject={encounter} config={zone} />}
				areaTypes={<AreaTypesPicker encounter={encounter} />}
				preset={showDamage && <ConnectedChoiceField {...choice} />}
				allies={player.canEnableTargetDummies() && <ConnectedNumberField modObject={host.sim.raid} config={allies} />}
				minBaseDamage={player.getPlayerSpec().isTankSpec && <ConnectedNumberField modObject={encounter} config={minBaseDamage} />}
				onPlannerClick={() => setPlannerOpen(true)}
				onNewClick={showDamage ? startNew : undefined}
			/>
			<EncounterPlanner open={plannerOpen} onOpenChange={onPlannerOpenChange} replaced={replaced} showExecuteProportion={showExecuteProportion} />
		</>
	);
};
