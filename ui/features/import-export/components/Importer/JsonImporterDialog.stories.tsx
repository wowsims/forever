import type { Meta, StoryObj } from '@storybook/react-vite';
import { DialogTrigger } from '@testing/storybook/fixtures/import-export';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { JsonImporterDialog } from './JsonImporterDialog';

const meta = {
	title: 'Features/ImportExport/Importer/Components/JsonImporterDialog',
	component: JsonImporterDialog,
	tags: ['autodocs'],
	args: { open: false, onOpenChange: fn() },
	decorators: [withSimHost()],
	render: args => <DialogTrigger dialog={JsonImporterDialog} {...args} />,
} satisfies Meta<typeof JsonImporterDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

export const Open: Story = { args: { open: true } };
