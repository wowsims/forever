import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { bundleDialogProps } from '@testing/storybook/fixtures/bulk-builds';

import { BulkBundleDialogView } from './BulkBundleDialog';

const meta = {
	title: 'Features/Bulk/BulkBundleDialog',
	component: BulkBundleDialogView,
	tags: ['autodocs'],
	args: bundleDialogProps(),
	argTypes: { pickers: { control: false } },
} satisfies Meta<typeof BulkBundleDialogView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Every picker on Current keeps Save off. */
export const NewBundle: Story = {};

export const EditBundle: Story = {
	args: bundleDialogProps({
		title: i18n.t('bulk_tab.builds.bundles.dialog_edit'),
		picks: { gearSet: 1, talents: 1, rotation: 1, settings: 1 },
	}),
};
