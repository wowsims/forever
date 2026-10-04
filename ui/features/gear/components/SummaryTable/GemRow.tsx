import type { UIGem as Gem } from '@generated/proto/ui';
import { ActionId } from '@sim/proto/action_id';
import { SummaryTableRow } from '@ui-kit/SummaryTableRow';
import { itemQualityClassName } from '@ui-kit/utils/css';
import { WowheadAnchor } from '@ui-kit/WowheadAnchor';
import clsx from 'clsx';

export interface GemRowProps {
	gem: Gem;
	count: number;
	/** Empty until the gem's icon has resolved. */
	iconUrl?: string;
}

export const GemRow = ({ gem, count, iconUrl }: GemRowProps) => {
	const href = ActionId.makeItemUrl(gem.id);
	return (
		<SummaryTableRow>
			<WowheadAnchor className={clsx('flex items-center', itemQualityClassName(gem.quality))} target="_blank" href={href} data-quality={gem.quality}>
				<img
					className="ui-summary-table-gem-icon static inset-gem z-1 inline-block size-gem-inner cursor-pointer rounded-none bg-cover bg-center bg-no-repeat"
					data-testid="gem-icon"
					src={iconUrl || undefined}
					alt=""
				/>
				{gem.name}
			</WowheadAnchor>
			<div className="ml-auto">{count.toFixed(0)}</div>
		</SummaryTableRow>
	);
};
