import { EncounterDesigner } from '@features/encounter-designer';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { Encounter as EncounterProto } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { Button } from '@ui-kit/Button';
import { ContentBlock } from '@ui-kit/ContentBlock';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { Icon } from '@ui-kit/Icon';
import { type ReactNode, useMemo, useState } from 'react';

import { useEncounterChoice } from '../../hooks/useEncounterChoice';
import { newEncounterProto } from '../../model/my_encounters';
import { AdvancedEncounterModal } from '../AdvancedEncounterModal';
import { TargetInputsPicker } from '../TargetsPicker';
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
	targetInputs: ReactNode;
	onAdvancedClick: () => void;
	onDesignClick: () => void;
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
	targetInputs,
	onAdvancedClick,
	onDesignClick,
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
			<div className="flex w-full gap-2">
				<Button className="flex-1" data-testid="design-encounter-button" onClick={onDesignClick}>
					{i18n.t('settings_tab.encounter.designer.open')}
				</Button>
				<Button className="flex-1" data-testid="advanced-button" onClick={onAdvancedClick}>
					{i18n.t('settings_tab.encounter.advanced')}
				</Button>
			</div>
			<FieldGrid>{duration}</FieldGrid>
			{execute && <FieldGrid className="w-full">{execute}</FieldGrid>}
			{zone}
			{areaTypes}
			{allies}
			{minBaseDamage}
			{targetInputs}
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

	const duration = useMemo(() => durationConfigs(), []);
	const execute = useMemo(() => (showExecuteProportion ? executeConfigs() : []), [showExecuteProportion]);
	const allies = useMemo(() => numAlliesConfig(player), [player]);
	const minBaseDamage = useMemo(() => minBaseDamageConfig(), []);
	const zone = useMemo(() => zoneConfig(encounter), [encounter]);

	const [advancedOpen, setAdvancedOpen] = useState(false);
	const [designerOpen, setDesignerOpen] = useState(false);
	const [replaced, setReplaced] = useState<EncounterProto>();

	const startNew = () => {
		setReplaced(encounter.toProto());
		encounter.fromProto(newEncounterProto());
		setDesignerOpen(true);
	};
	const onDesignerOpenChange = (open: boolean) => {
		setDesignerOpen(open);
		if (!open) setReplaced(undefined);
	};

	return (
		<>
			<EncounterPickerView
				duration={duration.map(config => (
					<ConnectedNumberField key={config.id} modObject={encounter} config={config} />
				))}
				execute={showExecuteProportion && execute.map(config => <ConnectedNumberField key={config.id} modObject={encounter} config={config} />)}
				zone={<ConnectedChoiceField modObject={encounter} config={zone} />}
				areaTypes={<AreaTypesPicker encounter={encounter} />}
				preset={showDamage && <ConnectedChoiceField {...choice} />}
				allies={player.canEnableTargetDummies() && <ConnectedNumberField modObject={host.sim.raid} config={allies} />}
				minBaseDamage={player.getPlayerSpec().isTankSpec && <ConnectedNumberField modObject={encounter} config={minBaseDamage} />}
				targetInputs={<TargetInputsPicker encounter={encounter} targetIndex={0} />}
				onAdvancedClick={() => setAdvancedOpen(true)}
				onDesignClick={() => setDesignerOpen(true)}
				onNewClick={showDamage ? startNew : undefined}
			/>
			<AdvancedEncounterModal open={advancedOpen} onOpenChange={setAdvancedOpen} />
			<EncounterDesigner open={designerOpen} onOpenChange={onDesignerOpenChange} replaced={replaced} />
		</>
	);
};
