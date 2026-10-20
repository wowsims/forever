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

const stringsOf = (group: 'encounters' | 'talents'): BulkCheckListProps['strings'] => {
	const t = (key: string) => i18n.t(`bulk_tab.settings.${group}.${key}`);
	return { description: t('description'), current: t('current'), presets: t('presets'), saved: t('saved'), empty: t('empty'), duplicate: t('duplicate') };
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

const meta = {
	title: 'Features/Bulk/BulkCheckList',
	component: BulkCheckList,
	tags: ['autodocs'],
	render: args => <StatefulCheckList {...args} />,
	args: {
		strings: stringsOf('encounters'),
		currentSummary: '3:00 ±5s · 1 target',
		presets: [presetRow('Patchwerk', '5:00 ±30s · 1 target', true)],
		saved: [savedRow('Council', '4:00 ±10s · 3 targets', true), savedRow('Execute test', 'Health · 1 target')],
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

export const Encounters: Story = {};

export const Talents: Story = {
	args: {
		strings: stringsOf('talents'),
		currentSummary: '30305001302',
		presets: [presetRow('Arms', '3030500130', true), presetRow('Fury', '-0505000540501')],
		saved: [savedRow('Arms, deep Cruelty', '3030500130-05')],
		testIds: { list: 'bulk-talents', row: 'bulk-talent', current: 'bulk-talent-current' },
	},
};

export const Duplicate: Story = {
	args: {
		saved: [
			savedRow('Council', '4:00 ±10s · 3 targets', true),
			savedRow('Council, saved twice', '4:00 ±10s · 3 targets', true, true),
			savedRow('Execute test', 'Health · 1 target'),
		],
	},
};

export const Empty: Story = { args: { presets: [], saved: [] } };
