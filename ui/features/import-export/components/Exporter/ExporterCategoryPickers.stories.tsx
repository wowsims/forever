import type { Meta, StoryObj } from '@storybook/react-vite';
import { useReducer, useState } from 'react';
import { fn } from 'storybook/test';

import { defaultExportCategories, type ExportCategories } from '../../exporters';
import { ExporterCategoryPickers, type ExporterCategoryPickersProps } from './ExporterCategoryPickers';

// The pickers write the record in place, so the story re-renders on `onChange` to show the write.
const StatefulPickers = ({ categories: initial, onChange }: ExporterCategoryPickersProps) => {
	const [categories] = useState<ExportCategories>(() => ({ ...initial }));
	const [, rerender] = useReducer((revision: number) => revision + 1, 0);
	return (
		<ExporterCategoryPickers
			categories={categories}
			onChange={() => {
				rerender();
				onChange();
			}}
		/>
	);
};

const meta = {
	title: 'Features/ImportExport/Exporter/Components/ExporterCategoryPickers',
	component: ExporterCategoryPickers,
	tags: ['autodocs'],
	args: { categories: defaultExportCategories(), onChange: fn() },
	argTypes: { categories: { control: false } },
	render: args => <StatefulPickers {...args} />,
} satisfies Meta<typeof ExporterCategoryPickers>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Defaults: Story = {};

export const NoneSelected: Story = {
	args: { categories: Object.fromEntries(Object.keys(defaultExportCategories()).map(category => [category, false])) as ExportCategories },
};
