import { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateWeaponType } from '@i18n/localization';
import { getBulkFreezeWeaponTypes } from '@sim/bulk/utils';
import { usePlayer } from '@sim/context/SimHostContext';
import { BooleanPicker } from '@ui-kit/BooleanPicker';

import { trackEvent } from '../../../../tracking/utils';
import { useBulkState } from '../../hooks/useBulkState';
import { setBulkWeaponTypeFilter } from '../../model/gear/settings';

export interface FreezeWeaponTypesProps {
	slot: ItemSlot.ItemSlotMainHand | ItemSlot.ItemSlotOffHand;
}

export const FreezeWeaponTypes = ({ slot }: FreezeWeaponTypesProps) => {
	const player = usePlayer();
	const frozenWeaponSlot = useBulkState(slice => slice.frozenWeaponSlot);
	const weaponTypeFilters = useBulkState(slice => slice.weaponTypeFilters);
	const weaponTypes = getBulkFreezeWeaponTypes(player, slot);

	return (
		<>
			{!!weaponTypes.length && frozenWeaponSlot !== slot && (
				<div className="flex flex-col gap-2">
					<h6 className="mb-0 text-sm font-semibold">
						{slot === ItemSlot.ItemSlotMainHand
							? i18n.t('bulk_tab.settings.freeze_weapon_types.mainhand_label')
							: i18n.t('bulk_tab.settings.freeze_weapon_types.offhand_label')}
					</h6>
					<div className="text-ui">{i18n.t('bulk_tab.settings.freeze_weapon_types.tooltip')}</div>
					<div className="grid grid-cols-2 gap-1">
						{weaponTypes.map(weaponType => (
							<BooleanPicker
								key={weaponType}
								id={`bulk-${slot}-weapon-type-${weaponType}`}
								label={translateWeaponType(weaponType)}
								layout="inline"
								value={weaponTypeFilters.get(slot)!.includes(weaponType)}
								onChange={newValue => {
									const filter = weaponTypeFilters.get(slot)!;
									setBulkWeaponTypeFilter(player, slot, newValue ? [...filter, weaponType] : filter.filter(type => type !== weaponType));
									trackEvent({ action: 'settings', category: 'batch_sim', label: `freeze_${slot}_weapon_type`, value: newValue });
								}}
							/>
						))}
					</div>
				</div>
			)}
		</>
	);
};
