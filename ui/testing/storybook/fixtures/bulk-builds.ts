import type { BuildBundlesViewProps } from '@features/bulk/components/BuildBundles';
import type { BulkBundleDialogViewProps } from '@features/bulk/components/BulkBundleDialog';
import type { BulkBundleListProps } from '@features/bulk/components/BulkBundleList';
import type { BulkCheckRow } from '@features/bulk/components/BulkCheckList';
import type { EncounterListViewProps } from '@features/bulk/components/EncounterList';
import type { SavedRef } from '@features/bulk/model/builds/saved_refs';
import i18n from '@i18n/config';
import { fn } from 'storybook/test';

const ref = (name: string, preset = false, group?: string): SavedRef => ({ name, preset, group });

const CURRENT = ref('');

const checkRow = (name: string, summary: string, preset: boolean, checked: boolean, duplicate = false): BulkCheckRow => ({
	ref: ref(name, preset),
	name,
	summary,
	checked,
	duplicate,
});

export const encounterListProps = (overrides: Partial<EncounterListViewProps> = {}): EncounterListViewProps => ({
	current: { label: i18n.t('bulk_tab.settings.encounters.current'), summary: '3:00 ±5s · 1 target', checked: true, onCheckedChange: fn() },
	presets: [checkRow('Patchwerk', '5:00 ±5s · 1 target', true, true), checkRow('Three targets', '3:00 ±5s · 3 targets', true, false)],
	saved: [
		checkRow('Long fight', '8:00 ±10s · 1 target', false, true),
		{
			...checkRow('Patchwerk copy', '5:00 ±5s · 1 target', false, true),
			disabledReason: i18n.t('bulk_tab.settings.encounters.duplicate', { encounter: 'Patchwerk' }),
		},
	],
	onCheckedChange: fn(),
	...overrides,
});

/** A saved encounter stored ticked that sims what the current encounter does: disabled while the current one is ticked. */
export const sameAsCurrentEncounterRow = (currentChecked: boolean): BulkCheckRow => ({
	...checkRow('Three minutes', '3:00 ±5s · 1 target', false, true),
	disabledReason: currentChecked
		? i18n.t('bulk_tab.settings.encounters.duplicate', { encounter: i18n.t('bulk_tab.settings.encounters.current') })
		: undefined,
});

const FURY_BUNDLE = 'Pre-BiS (Fury) · Fury · Fury APL · Pre-BiS - Fury';

/** The last row holds a saved copy of the Fury APL, so it sims what the first does. */
export const DUPLICATE_BUNDLE = 'Pre-BiS (Fury) · Fury · My Fury APL · Pre-BiS - Fury';

export const bundleListProps = (overrides: Partial<BulkBundleListProps> = {}): BulkBundleListProps => ({
	current: { label: i18n.t('bulk_tab.builds.bundles.current'), checked: true, onCheckedChange: fn() },
	rows: [
		{ key: '0', label: FURY_BUNDLE, checked: true },
		{ key: '1', label: 'Pre-BiS (Arms) · Arms · Arms APL · Pre-BiS - Arms', checked: true },
		{ key: '2', label: DUPLICATE_BUNDLE, checked: true, disabledReason: i18n.t('bulk_tab.builds.bundles.duplicate', { bundle: FURY_BUNDLE }) },
	],
	onCheckedChange: fn(),
	onEdit: fn(),
	onDelete: fn(),
	testIds: { list: 'bulk-bundles', row: 'bulk-bundle', current: 'bulk-bundle-current' },
	...overrides,
});

const PICKER_OPTIONS: Record<BulkBundleDialogViewProps['pickers'][number]['id'], Array<{ value: SavedRef; label: string }>> = {
	gearSet: [
		{ value: CURRENT, label: i18n.t('bulk_tab.results.current_gear') },
		{ value: ref('Pre-BiS', true, 'Fury'), label: 'Pre-BiS (Fury)' },
		{ value: ref('Pre-BiS', true, 'Arms'), label: 'Pre-BiS (Arms)' },
		{ value: ref('My raid set'), label: 'My raid set' },
	],
	talents: [
		{ value: CURRENT, label: i18n.t('bulk_tab.results.current_talents') },
		{ value: ref('Fury', true), label: 'Fury' },
		{ value: ref('Arms', true), label: 'Arms' },
	],
	rotation: [
		{ value: CURRENT, label: i18n.t('bulk_tab.results.current_rotation') },
		{ value: ref('Fury APL', true), label: 'Fury APL' },
		{ value: ref('Arms APL', true), label: 'Arms APL' },
		{ value: ref('My Fury APL'), label: 'My Fury APL' },
	],
	settings: [
		{ value: CURRENT, label: i18n.t('bulk_tab.results.current_settings') },
		{ value: ref('Pre-BiS - Fury', true), label: 'Pre-BiS - Fury' },
		{ value: ref('Pre-BiS - Arms', true), label: 'Pre-BiS - Arms' },
	],
};

const PICKER_LABEL_KEYS = { gearSet: 'gear_set', talents: 'talents', rotation: 'rotation', settings: 'settings' } as const;

/** `picks` holds the index of each picker's option, Current (0) when left out. */
export const bundleDialogProps = ({
	title = i18n.t('bulk_tab.builds.bundles.dialog_new'),
	picks = {},
}: { title?: string; picks?: Partial<Record<keyof typeof PICKER_OPTIONS, number>> } = {}): BulkBundleDialogViewProps => {
	const pickers = (Object.keys(PICKER_OPTIONS) as Array<keyof typeof PICKER_OPTIONS>).map(id => ({
		id,
		label: i18n.t(`bulk_tab.builds.parts.${PICKER_LABEL_KEYS[id]}`),
		options: PICKER_OPTIONS[id],
		value: PICKER_OPTIONS[id][picks[id] ?? 0].value,
		onChange: fn(),
	}));
	return {
		open: true,
		onOpenChange: fn(),
		title,
		pickers,
		canSave: pickers.some(picker => picker.value !== CURRENT),
		onSave: fn(),
	};
};

export const buildBundlesProps = (overrides: Partial<BuildBundlesViewProps> = {}): BuildBundlesViewProps => ({
	list: bundleListProps(),
	onNew: fn(),
	dialog: null,
	enchantPicks: [checkRow('Agility cloak', '1 slot.', false, true), checkRow('Crusader weapons', '2 slots.', false, false)],
	onEnchantPickCheckedChange: fn(),
	onNewEnchantPreset: fn(),
	enchantPresetDialog: null,
	...overrides,
});
