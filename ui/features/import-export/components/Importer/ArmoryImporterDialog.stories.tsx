import type { Meta, StoryObj } from '@storybook/react-vite';
import { DialogTrigger } from '@testing/storybook/fixtures/import-export';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { ArmoryImporterDialog } from './ArmoryImporterDialog';

const meta = {
	title: 'Features/ImportExport/Importer/Components/ArmoryImporterDialog',
	component: ArmoryImporterDialog,
	tags: ['autodocs'],
	args: { open: false, onOpenChange: fn() },
	decorators: [withSimHost()],
	render: args => <DialogTrigger dialog={ArmoryImporterDialog} {...args} />,
} satisfies Meta<typeof ArmoryImporterDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

export const Open: Story = { args: { open: true } };
