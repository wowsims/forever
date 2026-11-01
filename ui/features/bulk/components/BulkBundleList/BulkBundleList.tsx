import i18n from '@i18n/config';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Icon } from '@ui-kit/Icon';
import { useId, useState } from 'react';

export interface BulkBundleListProps {
	current: { label: string; checked: boolean; onCheckedChange(checked: boolean): void };
	/** A row with a `disabledReason` shows unchecked and cannot be ticked, whatever `checked` says. */
	rows: ReadonlyArray<{ key: string; label: string; checked: boolean; disabledReason?: string }>;
	onCheckedChange(key: string, checked: boolean): void;
	onEdit(key: string): void;
	onDelete(key: string): void;
	testIds: { list: string; row: string; current: string };
}

const t = (key: string, bundle: string) => i18n.t(`bulk_tab.builds.bundles.${key}`, { bundle });

// The Current bundle heads the list with no pencil and no ×. A disabled row is unchecked and its box stays in the tab
// order. Its reason shows below the row on the label's hover and the box's focus, clear of the row's own buttons, and
// describes the box and the row's buttons for keyboard and screen reader users.
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
				const reasonId = row.disabledReason ? `${idPrefix}-${index}-reason` : undefined;
				return (
					<div key={row.key} className="flex items-center gap-2" data-testid={testIds.row}>
						<BooleanPicker
							id={`${idPrefix}-${index}`}
							className="min-w-0 flex-1 wrap-anywhere"
							label={row.label}
							labelTooltip={row.disabledReason}
							labelTooltipPlace="bottom"
							layout="inline"
							size="xs"
							value={row.disabledReason ? false : row.checked}
							disabled={!!row.disabledReason}
							focusableWhenDisabled
							aria-describedby={reasonId}
							onChange={checked => onCheckedChange(row.key, checked)}
						/>
						{reasonId && (
							<span id={reasonId} className="sr-only">
								{row.disabledReason}
							</span>
						)}
						<Button iconOnly aria-label={t('edit', row.label)} aria-describedby={reasonId} onClick={() => onEdit(row.key)}>
							<Icon name="pencil" />
						</Button>
						<ConfirmPopover
							open={confirming === row.key}
							onOpenChange={open => setConfirming(open ? row.key : null)}
							trigger={<Icon name="xmark" />}
							triggerClassName="text-link-danger"
							triggerProps={{ 'aria-label': t('delete', row.label), 'aria-describedby': reasonId }}
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
