import type { IconEnumPickerConfig } from './types';

export const iconEnumPickerShown = <ModObject, T>(config: IconEnumPickerConfig<ModObject, T>, modObject: ModObject): boolean =>
	(!config.showWhen || config.showWhen(modObject)) &&
	config.values.some(valueConfig => !!valueConfig.actionId && (!valueConfig.showWhen || valueConfig.showWhen(modObject)));
