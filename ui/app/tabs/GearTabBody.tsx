import { GearPicker } from '@features/gear/components/GearPicker';
import { SelectorModal } from '@features/gear/components/SelectorModal';
import { OpenSelectorModalContext, useSelectorModalState } from '@features/gear/hooks/useSelectorModal';
import { useSimReady } from '@sim/hooks/useSimReady';
import { TabPanelColumns } from '@ui-kit/TabPanelColumns';

import { GearPresets } from '../GearPresets';

export const GearTabBody = () => {
	const ready = useSimReady();
	const selector = useSelectorModalState();

	return (
		<OpenSelectorModalContext value={selector.openTab}>
			<TabPanelColumns.Left variant="auto-columns">
				<GearPicker ready={ready} />
			</TabPanelColumns.Left>
			<TabPanelColumns.Right testId="gear-presets">
				<GearPresets />
			</TabPanelColumns.Right>
			<SelectorModal state={selector} />
		</OpenSelectorModalContext>
	);
};
