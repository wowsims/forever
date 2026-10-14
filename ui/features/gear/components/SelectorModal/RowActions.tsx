import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';

export interface RowActionsProps {
	favourited: boolean;
	favouriteTooltipId: string;
	onToggleFavourite: () => void;
	/** The batch toggle; the items tab alone offers it. */
	compare?: { inBatch: boolean; tooltipId: string; onToggle: () => void };
}

// The list's two shared tooltips read `data-favourited` and `data-in-batch` off the hovered button.
export const RowActions = ({ favourited, favouriteTooltipId, onToggleFavourite, compare }: RowActionsProps) => (
	<>
		<Button
			iconOnly
			aria-label={i18n.t('gear_tab.gear_picker.row_actions.favorite')}
			className={clsx('relative z-1 w-8', favourited && 'text-brand')}
			data-testid="selector-modal-list-item-favorite"
			data-tooltip-id={favouriteTooltipId}
			data-favourited={String(favourited)}
			onClick={event => {
				event.stopPropagation();
				onToggleFavourite();
			}}>
			<Icon name="star" style={favourited ? 'solid' : 'regular'} size="xl" />
		</Button>
		{compare && (
			<Button
				iconOnly
				aria-label={i18n.t('gear_tab.gear_picker.row_actions.compare')}
				className={clsx('relative z-1 w-8', compare.inBatch && 'text-brand')}
				data-testid="selector-modal-list-item-compare"
				data-tooltip-id={compare.tooltipId}
				data-in-batch={String(compare.inBatch)}
				onClick={compare.onToggle}>
				<Icon name="arrow-right-arrow-left" size="xl" />
			</Button>
		)}
	</>
);
