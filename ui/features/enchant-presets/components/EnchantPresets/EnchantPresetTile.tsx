import { useEnchantDescription } from '@features/gear/hooks/useEnchantDescription';
import { enchantActionId } from '@features/gear/model/item_data';
import type { ItemSlot } from '@generated/proto/common';
import type { UIEnchant as Enchant } from '@generated/proto/ui';
import { useActionId } from '@sim/hooks/useActionId';
import { useActionIdWowheadDataset } from '@sim/hooks/useActionIdWowheadDataset';
import type { WowheadDatasetProps } from '@sim/hooks/useWowheadDataset';
import { Icon } from '@ui-kit/Icon';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import { memo, useCallback, useMemo } from 'react';

export interface EnchantPresetTileViewProps {
	name: string;
	effect?: string;
	iconUrl?: string;
	href?: string;
	wowhead?: WowheadDatasetProps;
	checked: boolean;
	onToggle: () => void;
}

export const EnchantPresetTileView = ({ name, effect, iconUrl, href, wowhead, checked, onToggle }: EnchantPresetTileViewProps) => (
	<WowheadAnchor
		className="flex min-w-0 items-center gap-2 border border-border p-2 text-foreground hover:bg-surface-hover hover:text-foreground aria-checked:border-primary"
		testId="bulk-enchant-tile"
		href={href || undefined}
		role="checkbox"
		aria-checked={checked}
		onActivate={onToggle}
		{...wowhead}>
		<span className="ui-boolean-picker-input" data-size="xs" data-checked={checked ? '' : undefined} aria-hidden>
			{checked && (
				<span className="ui-boolean-picker-indicator">
					<Icon name="check" className="text-primary-foreground" />
				</span>
			)}
		</span>
		<WowheadIcon as="div" className="size-8 shrink-0" iconUrl={iconUrl || undefined} />
		<span className="flex min-w-0 flex-col">
			<span className="truncate">{name}</span>
			{effect && <span className="truncate text-xs text-gray-400">{effect}</span>}
		</span>
	</WowheadAnchor>
);

export interface EnchantPresetTileProps {
	slot: ItemSlot;
	enchant: Enchant;
	checked: boolean;
	onToggle: (slot: ItemSlot, effectId: number) => void;
}

export const EnchantPresetTile = memo(({ slot, enchant, checked, onToggle }: EnchantPresetTileProps) => {
	const actionId = useMemo(() => enchantActionId(enchant), [enchant]);
	const { iconUrl, href } = useActionId(actionId);
	const wowhead = useActionIdWowheadDataset(actionId);
	const description = useEnchantDescription(enchant);
	const toggle = useCallback(() => onToggle(slot, enchant.effectId), [onToggle, slot, enchant]);

	return (
		<EnchantPresetTileView
			name={enchant.name}
			effect={description === enchant.name ? undefined : description}
			iconUrl={iconUrl}
			href={href}
			wowhead={wowhead}
			checked={checked}
			onToggle={toggle}
		/>
	);
});
