import type { Meta, StoryObj } from '@storybook/react-vite';
import { DialogTrigger } from '@testing/storybook/fixtures/import-export';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { SimlImporterDialog, SimlRotationImporterDialog } from './SimlImporterDialog';

const meta = {
	title: 'Features/ImportExport/Importer/Components/SimlImporterDialog',
	component: SimlImporterDialog,
	tags: ['autodocs'],
	args: { open: false, onOpenChange: fn() },
	decorators: [withSimHost()],
	render: args => <DialogTrigger dialog={SimlImporterDialog} {...args} />,
} satisfies Meta<typeof SimlImporterDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

/** Any SIML: a full setup, or the part of one an uploaded file's name says it holds. */
export const Open: Story = { args: { open: true } };

/** The Rotation tab's importer, which reads the text as a rotation. */
export const Rotation: Story = {
	args: { open: true },
	render: args => <DialogTrigger dialog={SimlRotationImporterDialog} {...args} />,
};
