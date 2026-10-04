import type { Meta, StoryObj } from '@storybook/react-vite';
import { Dialog } from '@ui-kit/Dialog';
import { expect, userEvent, within } from 'storybook/test';

import { type ImportExportDialogProps, ImportExportKind } from '../import_export';
import { ImportExportItem } from './ImportExportItem';
import { ImportExportMenu } from './ImportExportMenu';

const stubDialog = (title: string) => {
	const StubDialog = ({ open, onOpenChange }: ImportExportDialogProps) => (
		<Dialog open={open} onOpenChange={onOpenChange} title={title}>
			<p>{title} dialog body</p>
		</Dialog>
	);
	return StubDialog;
};

const meta = {
	title: 'App/Header/Components/ImportExportMenu',
	component: ImportExportMenu,
	tags: ['autodocs'],
	args: {
		kind: ImportExportKind.Import,
		icon: 'download',
		title: 'Import',
		children: [
			<ImportExportItem key="json" label="JSON" dialog={stubDialog('JSON Import')} />,
			<ImportExportItem key="60u" label="60U" dialog={stubDialog('60U Import')} />,
			<ImportExportItem key="addon" label="Addon" dialog={stubDialog('Addon Import')} />,
		],
	},
	argTypes: { children: { control: false } },
	decorators: [
		Story => (
			<div className="flex list-none items-end pl-0 font-bold">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ImportExportMenu>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Import: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('import-link'));
		const items = await within(canvasElement.ownerDocument.body).findAllByTestId('sim-dropdown-item');
		await expect(items.map(item => item.textContent)).toEqual(['JSON', '60U', 'Addon']);
	},
};

export const Export: Story = {
	args: {
		kind: ImportExportKind.Export,
		icon: 'right-from-bracket',
		title: 'Export',
		children: [
			<ImportExportItem key="link" label="Link" dialog={stubDialog('Link Export')} />,
			<ImportExportItem key="json" label="JSON" dialog={stubDialog('JSON Export')} />,
		],
	},
};
