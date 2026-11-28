import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { enchantTooltipAnchor } from '@sim/proto/tooltips';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { useMemo } from 'react';

import { useEnchantDescription } from '../../hooks/useEnchantDescription';

export interface EnchantLabelProps {
	className: string;
	testId?: string;
	enchant?: Enchant | null;
	onActivate?: () => void;
	tooltipId?: string;
}

export const EnchantLabel = ({ className, testId, enchant, onActivate, tooltipId }: EnchantLabelProps) => {
	const description = useEnchantDescription(enchant);

	const actionId = useMemo(
		() => (enchant ? (enchant.spellId ? ActionId.fromSpellId(enchant.spellId) : ActionId.fromItemId(enchant.itemId)) : undefined),
		[enchant],
	);

	const tooltip = useMemo(() => (enchant ? enchantTooltipAnchor(enchant) : {}), [enchant]);

	const href = actionId ? (actionId.spellId ? ActionId.makeSpellUrl(actionId.spellId) : ActionId.makeItemUrl(actionId.itemId)) : undefined;

	if (!enchant) return null;

	return (
		// A label that opens the quick enchant list on hover keeps it; the enchant's own tooltip shows on the others.
		<WowheadAnchor
			className={className}
			testId={testId}
			href={href}
			onActivate={onActivate}
			{...tooltip}
			data-tooltip-id={tooltipId ?? tooltip['data-tooltip-id']}>
			{description}
		</WowheadAnchor>
	);
};
