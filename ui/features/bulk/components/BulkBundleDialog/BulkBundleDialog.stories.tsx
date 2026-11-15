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

/** Every picker on Current keeps Save off. The empty name shows the label the picks give the bundle. */
export const NewBundle: Story = {};

export const EditBundle: Story = {
	args: bundleDialogProps({
		title: i18n.t('bulk_tab.builds.bundles.dialog_edit'),
		picks: { gearSet: 1, talents: 1, rotation: 1, settings: 1 },
		name: 'Fury raid',
	}),
};

/** Another bundle's label as the name keeps Save off and says so under the field. */
export const NameTaken: Story = {
	args: bundleDialogProps({ picks: { gearSet: 2, talents: 2, rotation: 2, settings: 2 }, name: 'Current bundle', taken: true }),
};
