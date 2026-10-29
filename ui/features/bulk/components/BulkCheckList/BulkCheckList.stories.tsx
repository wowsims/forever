import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import type { SavedRef } from '../../model/saved_refs';
import { BulkCheckList, type BulkCheckListProps, type BulkCheckRow } from './BulkCheckList';

const StatefulCheckList = (props: BulkCheckListProps) => {
	const [presets, setPresets] = useState(props.presets);
	const [saved, setSaved] = useState(props.saved);
	const toggle = (rows: readonly BulkCheckRow[], ref: SavedRef, checked: boolean) => rows.map(row => (row.ref === ref ? { ...row, checked } : row));
	return (
		<BulkCheckList
			{...props}
			presets={presets}
			saved={saved}
			onCheckedChange={(ref, checked) => {
				if (ref.preset) setPresets(toggle(presets, ref, checked));
				else setSaved(toggle(saved, ref, checked));
				props.onCheckedChange(ref, checked);
			}}
		/>
	);
};

const encounterStrings: BulkCheckListProps['strings'] = {
	presets: i18n.t('bulk_tab.settings.encounters.presets'),
	saved: i18n.t('bulk_tab.settings.encounters.saved'),
	empty: i18n.t('bulk_tab.settings.encounters.empty'),
};

const enchantPickStrings: BulkCheckListProps['strings'] = {
	saved: i18n.t('bulk_tab.settings.enchant_picks.saved'),
	empty: i18n.t('bulk_tab.settings.enchant_picks.empty'),
	duplicate: i18n.t('bulk_tab.settings.enchant_picks.duplicate'),
};

const rowOf =
	(preset: boolean) =>
	(name: string, summary: string, checked = false, duplicate = false): BulkCheckRow => ({
		ref: { name, preset },
		name,
		summary,
		checked,
		duplicate,
	});
const presetRow = rowOf(true);
const savedRow = rowOf(false);

const currentEncounter = (checked: boolean) => ({
	label: i18n.t('bulk_tab.settings.encounters.current'),
	summary: '3:00 ±5s · 1 target',
	checked,
	onCheckedChange: fn(),
});

const meta = {
	title: 'Features/Bulk/BulkCheckList',
	component: BulkCheckList,
	tags: ['autodocs'],
	render: args => <StatefulCheckList {...args} />,
	args: {
		strings: encounterStrings,
		current: currentEncounter(true),
		presets: [presetRow('Patchwerk', '5:00 ±5s · 1 target', true), presetRow('Three targets', '3:00 ±5s · 3 targets')],
		saved: [savedRow('Long fight', '8:00 ±10s · 1 target', true)],
		onCheckedChange: fn(),
		testIds: { list: 'bulk-encounters', row: 'bulk-encounter', current: 'bulk-encounter-current' },
	},
	argTypes: {
		strings: { control: false },
		testIds: { control: false },
	},
	decorators: [
		Story => (
			<div className="w-80">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof BulkCheckList>;

export default meta;
type Story = StoryObj<typeof meta>;

export const WithoutCurrentRow: Story = {
	args: {
		strings: enchantPickStrings,
		current: null,
		presets: [],
		saved: [savedRow('Agility cloak', '1 slot.', true), savedRow('Crusader weapons', '2 slots.')],
		testIds: { list: 'bulk-enchant-picks', row: 'bulk-enchant-pick' },
	},
};

export const Duplicate: Story = {
	args: {
		strings: enchantPickStrings,
		current: null,
		presets: [],
		saved: [
			savedRow('Agility cloak', '1 slot.', true),
			savedRow('Agility cloak, saved twice', '1 slot.', true, true),
			savedRow('Crusader weapons', '2 slots.'),
		],
		testIds: { list: 'bulk-enchant-picks', row: 'bulk-enchant-pick' },
	},
};

export const Empty: Story = { args: { presets: [], saved: [] } };

export const CurrentUnticked: Story = { args: { current: currentEncounter(false) } };
