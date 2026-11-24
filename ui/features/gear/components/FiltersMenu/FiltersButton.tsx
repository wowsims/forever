import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';

export interface FiltersButtonProps {
	/** How many filters narrow the list; above zero, the label shows it. */
	active: number;
	onClick?: () => void;
	testId?: string;
}

export const FiltersButton = ({ active, onClick, testId }: FiltersButtonProps) => (
	<Button data-testid={testId} onClick={onClick}>
		{active > 0 ? i18n.t('gear_tab.gear_picker.filters_button_active', { active }) : i18n.t('gear_tab.gear_picker.filters_button')}
	</Button>
);
