import { EnchantPresetsView } from '@features/enchant-presets/components/EnchantPresets/EnchantPresets';
import { ItemSlot } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryBoolean } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { BulkSettingsView, type BulkSettingsViewProps } from './BulkSettings';
import { BulkTopResultsFieldView } from './BulkTopResultsField';
import { CombinationsCountView } from './CombinationsCount';

const pair = (id: string, label: string, first: [string, ItemSlot], second: [string, ItemSlot]) => ({
	id,
	label,
	options: [
		{ value: -1, label: 'None' },
		{ value: first[1], label: first[0] },
		{ value: second[1], label: second[0] },
	],
	value: -1,
	onChange: fn(),
});

const WEAPON_TYPES = (
	<div>
		<h6>Main hand weapon types</h6>
		<div className="grid grid-cols-2 gap-1">
			{['Sword', 'Axe', 'Mace', 'Dagger'].map(type => (
				<StoryBoolean key={type} id={`story-weapon-type-${type}`} label={type} layout="inline" initial={type === 'Sword'} />
			))}
		</div>
	</div>
);

const StatefulBulkSettings = (props: BulkSettingsViewProps) => {
	const [openGroups, setOpenGroups] = useState(props.openGroups);
	return (
		<BulkSettingsView
			{...props}
			openGroups={openGroups}
			onOpenGroupsChange={groups => {
				setOpenGroups(groups);
				props.onOpenGroupsChange(groups);
			}}
		/>
	);
};

const meta = {
	title: 'Features/Bulk/BulkSettings/BulkSettings',
	component: BulkSettingsView,
	tags: ['autodocs'],
	render: args => <StatefulBulkSettings {...args} />,
	args: {
		topResults: null,
		combinationsCount: <CombinationsCountView pending={false} ready combinations={48} iterations={96000} iterationsLimit={1000000} builds={null} />,
		canRun: true,
		onRun: fn(),
		openGroups: ['freezes'],
		onOpenGroupsChange: fn(),
		groups: true,
		enchants: (
			<EnchantPresetsView
				presets={[
					{ name: 'Undead', checked: true },
					{ name: 'Outdoor', checked: false },
				]}
				onCheckedChange={fn()}
				onNew={fn()}
				onEdit={fn()}
				onDelete={fn()}
				dialog={null}
			/>
		),
		frozenItems: [
			pair('story-freeze-ring', 'Freeze ring', ['Finger 1', ItemSlot.ItemSlotFinger1], ['Finger 2', ItemSlot.ItemSlotFinger2]),
			pair('story-freeze-trinket', 'Freeze trinket', ['Trinket 1', ItemSlot.ItemSlotTrinket1], ['Trinket 2', ItemSlot.ItemSlotTrinket2]),
		],
		weaponFreeze: null,
		onFrozenWeaponSlotChange: fn(),
	},
	argTypes: {
		topResults: { control: false },
		combinationsCount: { control: false },
		enchants: { control: false },
		frozenItems: { control: false },
		weaponFreeze: { control: false },
	},
} satisfies Meta<typeof BulkSettingsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const DualWield: Story = { args: { weaponFreeze: { value: -1, weaponTypes: WEAPON_TYPES } } };

export const CannotRun: Story = { args: { canRun: false } };

export const PowerTools: Story = { args: { topResults: <BulkTopResultsFieldView value={5} max={5000} onChange={fn()} /> } };

export const Builds: Story = {
	args: {
		topResults: <BulkTopResultsFieldView value={0} max={5000} onChange={fn()} />,
		groups: false,
		combinationsCount: (
			<CombinationsCountView
				pending={false}
				ready
				combinations={0}
				iterations={0}
				iterationsLimit={1000000}
				builds={{ setups: 6, bundles: 3, encounters: 2, runs: 6, builds: 12, totalIterations: 240000, largestRequestIterations: 60000 }}
			/>
		),
	},
};
