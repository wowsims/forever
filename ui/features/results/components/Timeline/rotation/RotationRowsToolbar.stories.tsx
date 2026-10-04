import type { Meta, StoryObj } from '@storybook/react-vite';
import { warriorRotationModel } from '@testing/storybook/fixtures/results-views';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { RotationRowsToolbar, type RotationRowsToolbarProps } from './RotationRowsToolbar';

const StatefulToolbar = ({ hidden: initial, onToggle, onShowAll, ...props }: RotationRowsToolbarProps) => {
	const [hidden, setHidden] = useState(initial);
	return (
		<RotationRowsToolbar
			{...props}
			hidden={hidden}
			onToggle={key => {
				setHidden(prev => {
					const next = new Set(prev);
					if (!next.delete(key)) next.add(key);
					return next;
				});
				onToggle(key);
			}}
			onShowAll={() => {
				setHidden(new Set());
				onShowAll();
			}}
		/>
	);
};

const meta = {
	title: 'Features/Results/Timeline/Components/RotationRowsToolbar',
	component: RotationRowsToolbar,
	tags: ['autodocs'],
	args: { model: warriorRotationModel(), hidden: new Set(), onToggle: fn(), onShowAll: fn() },
	argTypes: { model: { control: false }, hidden: { control: false } },
	render: args => <StatefulToolbar {...args} />,
} satisfies Meta<typeof RotationRowsToolbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const AllShown: Story = {};

export const RowsHidden: Story = { args: { hidden: new Set(['cast:Heroic Strike', 'cast:Deep Wounds', 'resource:rage', 'gcd:player']) } };

export const NoModel: Story = { args: { model: null } };
