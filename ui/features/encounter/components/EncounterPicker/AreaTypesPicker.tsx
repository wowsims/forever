import { ConnectedMultiComboBox } from '@features/inputs/components/ConnectedMultiComboBox';
import type { MultiComboBoxConfig } from '@features/inputs/configs';
import { AreaType } from '@generated/proto/common';
import i18n from '@i18n/config';
import { getAreaTypeI18nKey } from '@i18n/entity_mapping';
import { translateAreaType } from '@i18n/localization';
import type { Encounter } from '@sim/raid/encounter';
import { useMemo } from 'react';

import { trackEvent } from '../../../../tracking/utils';

export const AREA_TYPES: ReadonlyArray<AreaType> = [
	AreaType.AreaTypeForestGrassland,
	AreaType.AreaTypeMountainous,
	AreaType.AreaTypeSnowy,
	AreaType.AreaTypeDesert,
	AreaType.AreaTypeSwamp,
	AreaType.AreaTypeWasteland,
	AreaType.AreaTypeHaunted,
	AreaType.AreaTypeCavernous,
	AreaType.AreaTypeVolcanic,
	AreaType.AreaTypeStrongholdsCities,
];

export const areaTypesConfig = (): MultiComboBoxConfig<Encounter> => ({
	id: 'encounter-area-types',
	label: i18n.t('settings_tab.encounter.area_types.label'),
	labelTooltip: i18n.t('settings_tab.encounter.area_types.tooltip'),
	placeholder: i18n.t('settings_tab.encounter.area_types.placeholder'),
	emptyText: i18n.t('settings_tab.encounter.area_types.empty'),
	removeLabel: i18n.t('settings_tab.encounter.area_types.remove'),
	openLabel: i18n.t('settings_tab.encounter.area_types.open'),
	values: AREA_TYPES.map(areaType => ({ name: translateAreaType(areaType), value: areaType })),
	storeField: 'encounter:*',
	getValue: (encounter: Encounter) => encounter.getAreaTypes(),
	setValue: (encounter: Encounter, next: Array<number>) => {
		const before = encounter.getAreaTypes();
		for (const areaType of AREA_TYPES) {
			const inArea = next.includes(areaType);
			if (inArea !== before.includes(areaType)) {
				trackEvent({ action: 'settings', category: 'area', label: getAreaTypeI18nKey(areaType), value: inArea });
			}
		}
		encounter.setAreaTypes(next as Array<AreaType>);
	},
});

export interface AreaTypesPickerProps {
	encounter: Encounter;
}

export const AreaTypesPicker = ({ encounter }: AreaTypesPickerProps) => {
	const config = useMemo(() => areaTypesConfig(), []);
	return <ConnectedMultiComboBox modObject={encounter} config={config} className="w-full flex-col items-stretch" testId="encounter-area-types" />;
};
