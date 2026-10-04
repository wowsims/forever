import type { Meta, StoryObj } from '@storybook/react-vite';
import { DialogTrigger } from '@testing/storybook/fixtures/import-export';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { SixtyUpgradesImporterDialog } from './SixtyUpgradesImporterDialog';

const meta = {
	title: 'Features/ImportExport/Importer/Components/SixtyUpgradesImporterDialog',
	component: SixtyUpgradesImporterDialog,
	tags: ['autodocs'],
	args: { open: false, onOpenChange: fn() },
	decorators: [withSimHost()],
	render: args => <DialogTrigger dialog={SixtyUpgradesImporterDialog} {...args} />,
} satisfies Meta<typeof SixtyUpgradesImporterDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

export const Open: Story = { args: { open: true } };
