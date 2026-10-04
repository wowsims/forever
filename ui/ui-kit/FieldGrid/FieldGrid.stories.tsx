import type { Meta, StoryObj } from '@storybook/react-vite';
import { Field } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';

import { FieldGrid } from './FieldGrid';

const FIELDS = [
	{ id: 'field-grid-story-duration', label: 'Duration', value: '180' },
	{ id: 'field-grid-story-variation', label: 'Duration +/-', value: '5' },
	{ id: 'field-grid-story-execute', label: 'Execute Phase %', value: '20' },
];

const meta = {
	title: 'Kit/FieldGrid',
	component: FieldGrid,
	tags: ['autodocs'],
	argTypes: {
		className: { control: 'text' },
		children: { control: false },
	},
} satisfies Meta<typeof FieldGrid>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	args: {
		children: FIELDS.map(({ id, label, value }) => (
			<Field key={id} id={id} label={label}>
				<Input id={id} defaultValue={value} />
			</Field>
		)),
	},
};
