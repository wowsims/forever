import { MultiComboBox } from '@ui-kit/MultiComboBox';

import { useConnectedField } from '../../hooks/useConnectedField';
import type { MultiComboBoxConfig } from './types';

export interface ConnectedMultiComboBoxProps<M> {
	modObject: M;
	config: MultiComboBoxConfig<M>;
	className?: string;
	testId?: string;
}

export const ConnectedMultiComboBox = <M,>({ modObject, config, className, testId }: ConnectedMultiComboBoxProps<M>) => {
	const {
		value,
		setValue,
		hidden,
		disabled,
		frameProps: { className: frameClassName, ...frameProps },
	} = useConnectedField(modObject, config);
	if (hidden) return null;
	return (
		<MultiComboBox
			{...frameProps}
			className={[frameClassName, className]}
			testId={testId}
			value={value}
			onChange={setValue}
			disabled={disabled}
			values={config.values}
			placeholder={config.placeholder}
			emptyText={config.emptyText}
			removeLabel={config.removeLabel}
			openLabel={config.openLabel}
		/>
	);
};
