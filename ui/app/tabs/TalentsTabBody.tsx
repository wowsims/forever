import { SavedTalents } from '@features/talents/components/SavedTalents';
import { TalentsPicker } from '@features/talents/components/TalentsPicker';
import { PresetConfigurationCategory } from '@sim/constants/preset_categories';
import { useSimHost } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { classTalentsConfig } from '@sim/talents/factory';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';
import { useMemo } from 'react';

import { PresetConfigurationPicker } from '../PresetConfigurationPicker';

const TALENT_PRESETS = [PresetConfigurationCategory.Talents];

/** The tab's presets column, which the narrow screen's presets dialog also shows. */
export const TalentPresets = () => (
	<>
		<PresetConfigurationPicker categories={TALENT_PRESETS} />
		<SavedTalents />
	</>
);

export const TalentsTabBody = () => {
	const host = useSimHost();
	const player = host.player;

	const talentsConfig = useMemo(
		() => ({
			trees: classTalentsConfig[player.getClass()]!,
			storeField: 'talentsString' as const,
			getValue: (subject: Player<any>) => subject.getTalentsString(),
			setValue: (subject: Player<any>, newValue: string) => subject.setTalentsString(newValue),
		}),
		[player],
	);

	return (
		<>
			<TabPanelColumns.Left>
				<TalentsPicker config={talentsConfig} />
			</TabPanelColumns.Left>
			<TabPanelColumns.Right>
				<TalentPresets />
			</TabPanelColumns.Right>
		</>
	);
};
