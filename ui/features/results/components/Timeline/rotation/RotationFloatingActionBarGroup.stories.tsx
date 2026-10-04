import type { Meta, StoryObj } from '@storybook/react-vite';
import { warriorRotationModel } from '@testing/storybook/fixtures/results-views';
import { useState } from 'react';
import { fn } from 'storybook/test';

import type { ContentRow } from '../../../model/timeline/rotation';
import { RotationFloatingActionBarGroup, type RotationFloatingActionBarGroupProps } from './RotationFloatingActionBarGroup';

const MODEL = warriorRotationModel();
const ROWS = MODEL.rows.filter((row): row is ContentRow => row.kind !== 'header' && row.kind !== 'separator');

const StatefulGroup = ({ hidden: initial, onToggle, ...props }: RotationFloatingActionBarGroupProps) => {
	const [hidden, setHidden] = useState(initial);
	return (
		<RotationFloatingActionBarGroup
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
		/>
	);
};

const meta = {
	title: 'Features/Results/Timeline/Components/RotationFloatingActionBarGroup',
	component: RotationFloatingActionBarGroup,
	tags: ['autodocs'],
	args: { title: 'Player', rows: ROWS, hidden: new Set(), onToggle: fn() },
	argTypes: { rows: { control: false }, hidden: { control: false } },
	render: args => <StatefulGroup {...args} />,
} satisfies Meta<typeof RotationFloatingActionBarGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const AllShown: Story = {};

export const TwoHidden: Story = { args: { hidden: new Set(['cast:Heroic Strike', 'gcd:player']) } };
