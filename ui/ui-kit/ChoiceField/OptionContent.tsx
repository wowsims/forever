import type { ChoiceOption } from './types';
import { UnitIcon } from './UnitIcon';

export interface OptionContentProps<V> {
	option: ChoiceOption<V>;
}

export const OptionContent = <V,>({ option: { icon, label } }: OptionContentProps<V>) => (
	<>
		{typeof icon === 'string' ? <UnitIcon icon={icon} /> : icon && <UnitIcon iconUrl={icon.iconUrl} />}
		{label}
	</>
);
