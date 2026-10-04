import { useAplInput } from '@features/apl/hooks/useAplInput';
import type { InputConfig } from '@features/inputs/input';
import i18n from '@i18n/config';
import type { Player } from '@sim/player/player';
import { Field } from '@ui-kit/Field';

import { HeaderButton } from './HeaderButton';

export interface HidePickerProps {
	player: Player<any>;
	config: InputConfig<Player<any>, boolean>;
}

/** The eye that disables one APL list item. */
export const HidePicker = ({ player, config }: HidePickerProps) => {
	const { value, setValue, hidden, disabled, frameProps } = useAplInput(player, config);

	if (hidden) return null;
	return (
		<Field {...frameProps} className={['m-0 w-auto', frameProps.className]} testId="hide-picker-root" disabled={disabled}>
			<HeaderButton
				icon={value ? 'eye-slash' : 'eye'}
				pressed={value}
				label={i18n.t('rotation_tab.apl.helpers.buttons.enable_disable')}
				testId="hide-picker-button"
				onClick={() => setValue(!value)}
			/>
		</Field>
	);
};
