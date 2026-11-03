import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SavedDataPanel, type SavedDataPanelProps } from './SavedDataPanel';
import type { SavedDataPanelEntry } from './types';

const entryOf = (name: string, isPreset = false, tooltip?: string): SavedDataPanelEntry<string> => ({ name, data: name, json: name, isPreset, tooltip });

const PRESETS = [entryOf('Pre-Raid', true), entryOf('Phase 1 Raid', true, 'Molten Core and Onyxia'), entryOf('Phase 3 Raid', true)];
const USER_DATA = [entryOf('Fire Resistance'), entryOf('Tank Swap')];

type StoryProps = SavedDataPanelProps<string>;

const SavedDataPanelStory = ({ userData: initialUserData, currentJson: initialJson, onLoad, onSave, onDelete, ...props }: StoryProps) => {
	const [userData, setUserData] = useState(initialUserData);
	const [currentJson, setCurrentJson] = useState(initialJson);
	return (
		<SavedDataPanel
			{...props}
			userData={userData}
			currentJson={currentJson}
			onLoad={entry => {
				setCurrentJson(entry.json);
				onLoad(entry);
			}}
			onSave={name => {
				setUserData(entries => [...entries.filter(entry => entry.name !== name), { ...entryOf(name), json: currentJson }]);
				onSave(name);
			}}
			onDelete={deleted => {
				setUserData(entries => entries.filter(entry => entry.name !== deleted.name));
				onDelete(deleted);
			}}
		/>
	);
};

const meta = {
	title: 'Kit/SavedDataPanel',
	component: SavedDataPanel<string>,
	tags: ['autodocs'],
	args: {
		title: 'Saved Gear Sets',
		label: 'Gear Set',
		presets: [],
		userData: [],
		currentJson: 'Phase 1 Raid',
		onLoad: fn(),
		onSave: fn(),
		onDelete: fn(),
	},
	argTypes: {
		headerChildren: { control: false },
		presets: { control: false },
		userData: { control: false },
		isActive: { control: false },
		loadOnly: { control: 'boolean' },
		className: { control: 'text' },
	},
	render: args => <SavedDataPanelStory {...args} />,
} satisfies Meta<StoryProps>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Empty: Story = {};
export const WithPresets: Story = { args: { presets: PRESETS } };
export const WithUserData: Story = { args: { presets: PRESETS, userData: USER_DATA } };
export const LoadOnly: Story = { args: { presets: PRESETS, userData: USER_DATA, loadOnly: true } };
export const WithHeaderChildren: Story = {
	args: {
		presets: PRESETS,
		userData: USER_DATA,
		headerChildren: (
			<Button variant="link" size="sm" className="ml-auto">
				Options
			</Button>
		),
	},
};
export const SaveFlow: Story = {
	args: { presets: PRESETS, userData: USER_DATA },
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		const name = canvas.getByTestId('saved-data-save-input');
		await userEvent.clear(name);
		await userEvent.type(name, 'Fire Resistance Farm');
		await userEvent.click(canvas.getByTestId('saved-data-save-button'));
		await expect(args.onSave).toHaveBeenCalledWith('Fire Resistance Farm');
		await expect(within(canvas.getByTestId('saved-data-custom')).getByRole('button', { name: 'Fire Resistance Farm' })).toBeInTheDocument();
	},
};
