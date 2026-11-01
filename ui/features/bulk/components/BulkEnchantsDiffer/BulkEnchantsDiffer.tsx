import i18n from '@i18n/config';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useId } from 'react';

export interface BulkEnchantsDifferProps {
	/** Each as "Slot: Enchant", listed on hover and focus. */
	enchants: readonly string[];
}

// The sr-only copy of the list describes the count for screen readers, which the tooltip does not reach.
export const BulkEnchantsDiffer = ({ enchants }: BulkEnchantsDifferProps) => {
	const tooltipId = useId();
	return (
		<span className="block text-xs text-muted">
			<span
				tabIndex={0}
				className="cursor-default focus-visible:focus-ring"
				aria-describedby={`${tooltipId}-list`}
				data-testid="bulk-enchants-differ"
				{...tooltipAnchorProps(tooltipId)}>
				{i18n.t('bulk_tab.compare.enchants_differ', { count: enchants.length })}
			</span>
			<span id={`${tooltipId}-list`} className="sr-only">
				{enchants.join(', ')}
			</span>
			<Tooltip
				id={tooltipId}
				align="start"
				content={
					<ul className="m-0 list-none p-0">
						{enchants.map((enchant, index) => (
							<li key={index}>{enchant}</li>
						))}
					</ul>
				}
			/>
		</span>
	);
};
