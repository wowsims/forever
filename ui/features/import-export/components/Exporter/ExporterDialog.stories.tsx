import { WOWHEAD_GEAR_PLANNER_URL } from '@sim/proto/wowhead';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { DialogTrigger } from '@testing/storybook/fixtures/import-export';
import { withSimHost } from '@testing/storybook/fixtures/models';
import { fn } from 'storybook/test';

import { type ExportCategories, JSON_EXPORTER, LINK_EXPORTER, selectedCategories, SIML_EXPORTER, WOWHEAD_GEAR_PLANNER_EXPORTER } from '../../exporters';
import { exporterDialog } from './ExporterDialog';

const listCategories = (_host: unknown, categories: ExportCategories) => JSON.stringify({ categories: selectedCategories(categories) }, null, 2);

const JsonExporterDialog = exporterDialog({ ...JSON_EXPORTER, getData: listCategories });
const LinkExporterDialog = exporterDialog({ ...LINK_EXPORTER, getData: listCategories });
const WowheadExporterDialog = exporterDialog({ ...WOWHEAD_GEAR_PLANNER_EXPORTER, getData: () => `${WOWHEAD_GEAR_PLANNER_URL}/warrior/human` });
const SimlExporterDialog = exporterDialog({
	...SIML_EXPORTER,
	getData: (_host, categories) => Promise.resolve(`siml 1\nspec hunter\n\n// ${selectedCategories(categories).length} categories ticked\n`),
});

const meta = {
	title: 'Features/ImportExport/Exporter/Components/ExporterDialog',
	component: JsonExporterDialog,
	tags: ['autodocs'],
	args: { open: false, onOpenChange: fn() },
	decorators: [withSimHost()],
	render: args => <DialogTrigger dialog={JsonExporterDialog} {...args} />,
} satisfies Meta<typeof JsonExporterDialog>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};

/** Category pickers and a download button; the text lists the categories ticked. */
export const Json: Story = { args: { open: true } };

/** Category pickers, no download. */
export const Link: Story = {
	args: { open: true },
	render: args => <DialogTrigger dialog={LinkExporterDialog} {...args} />,
};

/** Neither category pickers nor a download: the text and the copy button only. */
export const Wowhead: Story = {
	args: { open: true },
	render: args => <DialogTrigger dialog={WowheadExporterDialog} {...args} />,
};

/** Category pickers and a download, with the text a promise gives: it shows once the promise resolves. */
export const Siml: Story = {
	args: { open: true },
	render: args => <DialogTrigger dialog={SimlExporterDialog} {...args} />,
};
