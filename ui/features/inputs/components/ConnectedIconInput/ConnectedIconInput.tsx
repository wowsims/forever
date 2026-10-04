import { ConnectedIconEnumPicker } from '../ConnectedIconEnumPicker';
import { ConnectedIconPicker } from '../ConnectedIconPicker';
import type { ConnectedIconInputConfig } from './types';

export interface ConnectedIconInputProps<ModObject> {
	modObject: ModObject;
	config: ConnectedIconInputConfig<ModObject, any>;
	testId?: string;
}

export const ConnectedIconInput = <ModObject,>({ modObject, config, testId }: ConnectedIconInputProps<ModObject>) =>
	config.type === 'icon' ? (
		<ConnectedIconPicker modObject={modObject} config={config} testId={testId} />
	) : (
		<ConnectedIconEnumPicker modObject={modObject} config={config} testId={testId} />
	);
