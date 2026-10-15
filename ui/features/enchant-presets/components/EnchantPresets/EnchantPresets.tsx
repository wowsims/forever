import type { BulkEnchantPreset } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ConfirmPopover } from '@ui-kit/ConfirmPopover';
import { Icon } from '@ui-kit/Icon';
import { type ReactNode, useId, useState } from 'react';

import { useEnchantPresets } from '../../hooks/useEnchantPresets';
import { deleteEnchantPreset, setEnchantPresetChecked } from '../../model/enchant_presets';
import { EnchantPresetDialog } from './EnchantPresetDialog';

export interface EnchantPresetRow {
	name: string;
	checked: boolean;
}

export interface EnchantPresetsViewProps {
	presets: readonly EnchantPresetRow[];
	onCheckedChange: (name: string, checked: boolean) => void;
	onNew: () => void;
	onEdit: (name: string) => void;
	onDelete: (name: string) => void;
	dialog: ReactNode;
}

export const EnchantPresetsView = ({ presets, onCheckedChange, onNew, onEdit, onDelete, dialog }: EnchantPresetsViewProps) => {
	const idPrefix = useId();
	const [confirming, setConfirming] = useState<string | null>(null);

	return (
		<div className="grid gap-2" data-testid="bulk-enchant-presets">
			<p className="mb-0 text-xs">{i18n.t('bulk_tab.settings.enchant_presets.description')}</p>
			{presets.length ? (
				<div className="grid gap-1">
					{presets.map((preset, index) => (
						<div key={preset.name} className="flex items-center gap-2" data-testid="bulk-enchant-preset">
							<BooleanPicker
								id={`${idPrefix}-${index}`}
								className="min-w-0 flex-1"
								label={preset.name}
								layout="inline"
								size="sm"
								value={preset.checked}
								onChange={checked => onCheckedChange(preset.name, checked)}
							/>
							<Button
								iconOnly
								aria-label={i18n.t('bulk_tab.settings.enchant_presets.edit', { name: preset.name })}
								onClick={() => onEdit(preset.name)}>
								<Icon name="pencil" />
							</Button>
							<ConfirmPopover
								open={confirming === preset.name}
								onOpenChange={open => setConfirming(open ? preset.name : null)}
								trigger={<Icon name="xmark" />}
								triggerClassName="text-link-danger"
								triggerProps={{ 'aria-label': i18n.t('bulk_tab.settings.enchant_presets.delete', { name: preset.name }) }}
								confirmLabel={i18n.t('common.delete')}
								onConfirm={() => onDelete(preset.name)}>
								{i18n.t('bulk_tab.settings.enchant_presets.delete_confirm', { name: preset.name })}
							</ConfirmPopover>
						</div>
					))}
				</div>
			) : (
				<p className="mb-0 text-gray-400">{i18n.t('bulk_tab.settings.enchant_presets.empty')}</p>
			)}
			<Button variant="outline-primary" size="sm" onClick={onNew}>
				<Icon name="plus" />
				{i18n.t('bulk_tab.settings.enchant_presets.new')}
			</Button>
			{dialog}
		</div>
	);
};

interface DialogState {
	open: boolean;
	preset: BulkEnchantPreset | null;
	// A fresh key per opening starts the dialog's draft from the preset again.
	key: number;
}

export const EnchantPresets = () => {
	const player = usePlayer();
	const presets = useEnchantPresets();
	const [dialog, setDialog] = useState<DialogState>({ open: false, preset: null, key: 0 });

	const openDialog = (preset: BulkEnchantPreset | null) => setDialog(current => ({ open: true, preset, key: current.key + 1 }));

	return (
		<EnchantPresetsView
			presets={presets}
			onCheckedChange={(name, value) => setEnchantPresetChecked(player, name, value)}
			onNew={() => openDialog(null)}
			onEdit={name => openDialog(presets.find(preset => preset.name === name) ?? null)}
			onDelete={name => deleteEnchantPreset(player, name)}
			dialog={
				dialog.open && (
					<EnchantPresetDialog key={dialog.key} open preset={dialog.preset} onClose={() => setDialog(current => ({ ...current, open: false }))} />
				)
			}
		/>
	);
};
