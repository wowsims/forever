import { ConnectedIconEnumPicker } from '@features/inputs/components/ConnectedIconEnumPicker';
import { ConnectedIconPicker } from '@features/inputs/components/ConnectedIconPicker';
import { ConnectedMultiIconPicker } from '@features/inputs/components/ConnectedMultiIconPicker';
import { clearMultiIconInputs } from '@features/settings/model/multi_icon';
import type { IconEnumPickerStatOption, MultiIconPickerStatOption, RenderableStatOptions } from '@features/settings/model/stat_options';
import { usePlayer, useSim } from '@sim/context/SimHostContext';
import { subscribeSimChange } from '@sim/state/subscriptions';

export interface StatOptionIconsProps {
	options: ReadonlyArray<RenderableStatOptions>;
}

const isMultiIcon = (option: RenderableStatOptions): option is MultiIconPickerStatOption => 'inputs' in option.config;
const isIconEnum = (option: RenderableStatOptions): option is IconEnumPickerStatOption => 'values' in option.config;

export const StatOptionIcons = ({ options }: StatOptionIconsProps) => {
	const player = usePlayer();
	const sim = useSim();
	const subscribe = subscribeSimChange(sim);

	return (
		<>
			{options.map((option, index) => {
				if (isMultiIcon(option)) {
					return (
						<ConnectedMultiIconPicker
							key={index}
							modObject={player}
							config={option.config}
							subscribe={subscribe}
							onClear={() => clearMultiIconInputs(player, option.config)}
						/>
					);
				}
				if (isIconEnum(option)) {
					return <ConnectedIconEnumPicker key={index} modObject={player} config={option.config} />;
				}
				return <ConnectedIconPicker key={index} modObject={player} config={option.config} />;
			})}
		</>
	);
};
