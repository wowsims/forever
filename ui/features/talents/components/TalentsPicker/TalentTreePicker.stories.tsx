import type { MageTalents } from '@generated/proto/mage';
import { mageTalentsConfig } from '@sim/talents/mage';
import type { TalentPoints } from '@sim/talents/talents_string';
import { parseTalentsString } from '@sim/talents/talents_string';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { buildTalentGraph } from '../../model/tree_graph';
import { TalentTreePicker, type TalentTreePickerProps } from './TalentTreePicker';

// A full 51-point Forever mage build: Arcane and Frost spent, Fire empty.
const MAGE_BUILD = '25502252231221--2555';
// The same Arcane tree with four points left in row 0, which breaks every talent below it.
const BROKEN_BUILD = '22002252231221--2555';

const graph = buildTalentGraph(mageTalentsConfig);

const StatefulTree = ({ points: initial, onChange, ...props }: TalentTreePickerProps<MageTalents>) => {
	const [points, setPoints] = useState(initial);
	const change = (next: TalentPoints) => {
		setPoints(next);
		onChange(next);
	};
	return <TalentTreePicker {...props} points={points} onChange={change} />;
};

const meta = {
	title: 'Features/Talents/TalentsPicker/TalentTreePicker',
	component: TalentTreePicker<MageTalents>,
	tags: ['autodocs'],
	args: { config: mageTalentsConfig, graph, points: parseTalentsString(mageTalentsConfig, ''), treeIdx: 0, onChange: fn() },
	argTypes: { config: { control: false }, graph: { control: false }, points: { control: false } },
	parameters: { smoke: false, smokeReason: 'Each talent fills its icon from the database through useActionId' },
	render: args => (
		<div className="flex w-fit">
			<StatefulTree {...args} />
		</div>
	),
} satisfies Meta<typeof TalentTreePicker<MageTalents>>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = { args: { points: parseTalentsString(mageTalentsConfig, MAGE_BUILD) } };

export const Broken: Story = { args: { points: parseTalentsString(mageTalentsConfig, BROKEN_BUILD) } };
