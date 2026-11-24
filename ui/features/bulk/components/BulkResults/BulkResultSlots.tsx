import { GearChangeIcon } from '@features/gear/components/GearChangeIcon';
import { ItemDetailCell } from '@features/gear/components/ItemCell';
import { ItemSpec } from '@generated/proto/common';
import { getBulkPlayerCanDualWield } from '@sim/bulk/utils';
import { useSimHost } from '@sim/context/SimHostContext';
import type { Gear } from '@sim/proto/gear';
import type { ReactNode } from 'react';

import { getSwappableItemSlotPair, itemSpecPairsEqualUnordered } from '../../model/gear/swappable_slots';

export interface BulkResultSlotsProps {
	gear: Gear;
	currentGear: Gear;
	/** Drawn after the slots, such as the name of the item a Drops row adds. */
	label?: ReactNode;
	/** Draws only the slots whose item or enchant differs from the current gear's in that same slot, for a row that changes one item. */
	changedOnly?: boolean;
}

/** A result row's gear as one cell per slot, filled where it differs from the current gear. */
export const BulkResultSlots = ({ gear, currentGear, label, changedOnly = false }: BulkResultSlotsProps) => {
	const host = useSimHost();
	const canDualWield = getBulkPlayerCanDualWield(host.player);
	const resultAsSpec = gear.asSpec();
	const originalEquipmentSpec = currentGear.asSpec();

	const cells = resultAsSpec.items.map((spec, idx) => {
		if (changedOnly && ItemSpec.equals(spec, originalEquipmentSpec.items[idx])) return null;
		const swappableItemSlotPair = getSwappableItemSlotPair(idx, canDualWield);
		const itemChanged = swappableItemSlotPair
			? !itemSpecPairsEqualUnordered(resultAsSpec.items, originalEquipmentSpec.items, swappableItemSlotPair)
			: !ItemSpec.equals(spec, originalEquipmentSpec.items[idx]);

		// Three display states per slot:
		// 1. Unchanged -  slots stay empty
		// 2. Same - Item with a different enchant shows the compact change icon (enchant marker);
		// 3. New - The item appearing at all already says the slot changed.
		return itemChanged && spec.id !== 0 && spec.id === originalEquipmentSpec.items[idx]?.id ? (
			<div key={idx} className="ui-bulk-result-item ui-bulk-item-cell" data-testid="bulk-result-item">
				<GearChangeIcon
					slot={idx}
					item={host.sim.db.lookupItemSpec(spec) ?? undefined}
					previousItem={currentGear.getEquippedItem(idx) ?? undefined}
					showEnchant
				/>
			</div>
		) : (
			<ItemDetailCell
				key={idx}
				testId="bulk-result-item"
				className="ui-bulk-result-item ui-bulk-item-cell mb-0 p-0 [&_.ui-item-picker-labels-container]:hidden"
				slot={idx}
				item={itemChanged && spec.id !== 0 ? host.sim.db.lookupItemSpec(spec) : null}
			/>
		);
	});
	if (!label) return cells;
	return (
		<>
			{cells}
			<div className="flex min-w-0 items-center pl-1">{label}</div>
		</>
	);
};
