import { formatDeltaText } from '@sim/utils/format';
import { toneTextClass } from '@ui-kit/utils/css';
import clsx from 'clsx';

export interface EpCellProps {
	ep: number;
	/** What the equipped item is worth; null when nothing is equipped to compare against. */
	equippedEP: number | null;
}

export const EpCell = ({ ep, equippedEP }: EpCellProps) => {
	const delta = equippedEP !== null && equippedEP !== ep ? formatDeltaText(equippedEP, ep, 0) : null;
	return (
		<div className="flex w-24 items-center in-data-hide-ep:hidden" data-testid="selector-modal-list-item-ep">
			<span className="inline-block text-right text-white" data-testid="selector-modal-list-item-ep-value">
				{ep < 9.95 ? ep.toFixed(1) : Math.round(ep).toString()}
			</span>
			<span
				className={clsx('ml-1 text-xs', toneTextClass(delta?.tone))}
				data-testid="selector-modal-list-item-ep-delta"
				data-sign={delta?.tone ?? undefined}>
				{delta?.text}
			</span>
		</div>
	);
};
