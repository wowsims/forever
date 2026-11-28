import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Icon } from '@ui-kit/Icon';
import { useId, useState } from 'react';

import { type BulkSetupLabels, BulkSetupLines, bulkSetupText } from '../BulkSetupLabel';

export interface BulkBundleListProps {
	current: { label: string; checked: boolean; onCheckedChange(checked: boolean): void };
	/**
	 * A row with a `disabledReason` shows unchecked and cannot be ticked, whatever `checked` says. A named row's label is
	 * its name, and `parts` holds its four parts.
	 */
	rows: ReadonlyArray<{ key: string; label: string; checked: boolean; disabledReason?: string; parts?: BulkSetupLabels }>;
	onCheckedChange(key: string, checked: boolean): void;
	onEdit(key: string): void;
	onDelete(key: string): void;
	testIds: { list: string; row: string; current: string };
}

const t = (key: string, bundle: string) => i18n.t(`bulk_tab.builds.bundles.${key}`, { bundle });

// A disabled row's reason, and a named row's parts after it, show below the row, clear of the row's own buttons.
export const BulkBundleList = ({ current, rows, onCheckedChange, onEdit, onDelete, testIds }: BulkBundleListProps) => {
	const idPrefix = useId();
	const [confirming, setConfirming] = useState<string | null>(null);
	return (
		<div className="grid gap-2" data-testid={testIds.list}>
			<BooleanPicker
				id={`${idPrefix}-current`}
				className="wrap-anywhere"
				label={current.label}
				layout="inline"
				size="xs"
				value={current.checked}
				onChange={current.onCheckedChange}
				testId={testIds.current}
			/>
			{rows.map((row, index) => {
				const detailsId = row.disabledReason || row.parts ? `${idPrefix}-${index}-details` : undefined;
				return (
					<div key={row.key} className="flex items-center gap-2" data-testid={testIds.row}>
						<BooleanPicker
							id={`${idPrefix}-${index}`}
							className="min-w-0 flex-1 wrap-anywhere"
							label={row.label}
							// An element, since a string tooltip is read as markup and this one quotes saved names.
							labelTooltip={
								detailsId && (
									<>
										{row.disabledReason && <span className="block">{row.disabledReason}</span>}
										{row.parts && <BulkSetupLines labels={row.parts} />}
									</>
								)
							}
							labelTooltipPlace="bottom"
							layout="inline"
							size="xs"
							value={row.disabledReason ? false : row.checked}
							disabled={!!row.disabledReason}
							focusableWhenDisabled
							aria-describedby={detailsId}
							onChange={checked => onCheckedChange(row.key, checked)}
						/>
						{detailsId && (
							<span id={detailsId} className="sr-only">
								{[row.disabledReason, row.parts && bulkSetupText(row.parts)].filter(Boolean).join('. ')}
							</span>
						)}
						<Button iconOnly aria-label={t('edit', row.label)} aria-describedby={detailsId} onClick={() => onEdit(row.key)}>
							<Icon name="pencil" />
						</Button>
						<ConfirmPopover
							open={confirming === row.key}
							onOpenChange={open => setConfirming(open ? row.key : null)}
							trigger={<Icon name="xmark" />}
							triggerClassName="text-link-danger"
							triggerProps={{ 'aria-label': t('delete', row.label), 'aria-describedby': detailsId }}
							confirmLabel={i18n.t('common.delete')}
							onConfirm={() => onDelete(row.key)}>
							{t('delete_confirm', row.label)}
						</ConfirmPopover>
					</div>
				);
			})}
		</div>
	);
};
