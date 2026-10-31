import { ConnectedBooleanPicker } from '@features/inputs/components/ConnectedBooleanPicker';
import { ConnectedChoiceField } from '@features/inputs/components/ConnectedChoiceField';
import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import { Target as TargetProto } from '@generated/proto/common';
import { useDisplayMetrics } from '@sim/hooks/useDisplayMetrics';
import { useStoreSubscribe } from '@sim/hooks/useStoreSubscribe';
import type { Encounter } from '@sim/raid/encounter';
import { subscribeEncounterField } from '@sim/state/subscriptions';
import { Field } from '@ui-kit/Field';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { type ReactNode, useMemo } from 'react';

import { TargetInputsPicker } from './TargetInputsPicker';
import {
	aiConfig,
	ALL_TARGET_STATS,
	bleedImmuneConfig,
	booleanConfigs,
	creatureFamiliesOf,
	creatureFamilyConfig,
	levelConfig,
	mobTypeConfig,
	npcConfig,
	numberConfigs,
	spellSchoolConfig,
	statConfig,
	tankIndexConfig,
	type TargetFieldContext,
} from './utils/configs';

export interface TargetPickerViewProps {
	/** Who the target is: NPC preset, AI, level, mob type, creature family, tank, bleed immunity and the target's own inputs. */
	identity: ReactNode;
	stats: ReactNode;
	/** The fields only a threat sim reads; without them an empty column keeps the three-column grid. */
	threat?: ReactNode;
}

export const TargetPickerView = ({ identity, stats, threat }: TargetPickerViewProps) => (
	<Field testId="target-picker-root" className="grid grid-cols-1 gap-2 xl:grid-cols-3">
		<FieldGrid data-testid="target-picker-section">{identity}</FieldGrid>
		<FieldGrid data-testid="target-picker-section">{stats}</FieldGrid>
		{threat ? (
			<FieldGrid data-testid="target-picker-section">{threat}</FieldGrid>
		) : (
			<div className="hidden xl:block" data-testid="target-picker-section-placeholder" aria-hidden="true" />
		)}
	</Field>
);

export interface TargetPickerProps {
	encounter: Encounter;
	targetIndex: number;
}

export const TargetPicker = ({ encounter, targetIndex }: TargetPickerProps) => {
	const { threat: showThreatMetrics } = useDisplayMetrics(encounter.sim);
	const context = useMemo(
		(): TargetFieldContext => ({
			encounter,
			targetIndex,
			getTarget: () => encounter.getTarget(targetIndex) || TargetProto.create(),
		}),
		[encounter, targetIndex],
	);

	const npc = useMemo(() => npcConfig(context), [context]);
	const ai = useMemo(() => aiConfig(context), [context]);
	const level = useMemo(() => levelConfig(context), [context]);
	const mobType = useMemo(() => mobTypeConfig(context), [context]);
	const targetMobType = useStoreSubscribe(subscribeEncounterField(encounter, 'targets'), () => context.getTarget().mobType);
	const creatureFamily = useMemo(() => creatureFamilyConfig(context, targetMobType), [context, targetMobType]);
	const tankIndex = useMemo(() => tankIndexConfig(context), [context]);
	const stats = useMemo(
		() => ALL_TARGET_STATS.filter(entry => showThreatMetrics || !entry.threatMetrics).map(entry => statConfig(context, entry.stat, entry.tooltip)),
		[context, showThreatMetrics],
	);
	const numbers = useMemo(() => numberConfigs(context), [context]);
	const booleans = useMemo(() => booleanConfigs(context), [context]);
	const bleedImmune = useMemo(() => bleedImmuneConfig(context), [context]);
	const spellSchool = useMemo(() => spellSchoolConfig(context), [context]);

	return (
		<TargetPickerView
			identity={
				<>
					<ConnectedChoiceField modObject={null} config={npc} />
					<ConnectedChoiceField modObject={null} config={ai} />
					<ConnectedChoiceField modObject={null} config={mobType} />
					{creatureFamiliesOf(targetMobType).length > 0 && <ConnectedChoiceField modObject={null} config={creatureFamily} />}
					<ConnectedChoiceField modObject={null} config={level} />
					{showThreatMetrics && <ConnectedChoiceField modObject={null} config={tankIndex} />}
					<ConnectedBooleanPicker modObject={null} config={bleedImmune} />
					<TargetInputsPicker encounter={encounter} targetIndex={targetIndex} />
				</>
			}
			stats={stats.map(config => (
				<ConnectedNumberField key={config.id} modObject={null} config={config} />
			))}
			threat={
				showThreatMetrics && (
					<>
						{numbers.map(config => (
							<ConnectedNumberField key={config.id} modObject={null} config={config} />
						))}
						{booleans.map(config => (
							<ConnectedBooleanPicker key={config.id} modObject={null} config={config} />
						))}
						<ConnectedChoiceField modObject={null} config={spellSchool} />
					</>
				)
			}
		/>
	);
};
