import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SimToolbarView } from './SimToolbar';

const meta = {
	title: 'App/Header/SimToolbar',
	component: SimToolbarView,
	tags: ['autodocs'],
	args: {
		knownIssues: [],
		outdatedNativeSim: false,
		onOpenSettings: fn(),
	},
	argTypes: { knownIssues: { control: false } },
	decorators: [
		Story => (
			<div className="flex list-none flex-nowrap items-end pl-0 text-(length:--text-ui) font-bold">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SimToolbarView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('sim-options'));
		await expect(args.onOpenSettings).toHaveBeenCalledOnce();
	},
};

export const KnownIssues: Story = {
	args: { knownIssues: ['Bloodrage costs no health.', 'Deep Wounds ticks on its own schedule.'] },
};

export const OutdatedNativeSim: Story = { args: { outdatedNativeSim: true } };

export const UpdatableNativeSim: Story = {
	args: { outdatedNativeSim: true, nativeSimUpdate: { latest: 'v1.2.3', status: { state: 'idle' }, onUpdate: fn() } },
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('native-sim-update'));
		const dialog = within(document.body);
		await expect(await dialog.findByTestId('native-sim-update-dialog')).toBeInTheDocument();
		await userEvent.click(dialog.getByRole('button', { name: i18n.t('info.native_sim_update_action') }));
		await expect(args.nativeSimUpdate?.onUpdate).toHaveBeenCalledOnce();
	},
};

export const UpdatingNativeSim: Story = {
	args: { outdatedNativeSim: true, nativeSimUpdate: { latest: 'v1.2.3', status: { state: 'updating' }, onUpdate: fn() } },
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('native-sim-update'));
		await expect(await within(document.body).findByTestId('native-sim-updating')).toBeInTheDocument();
	},
};

export const NativeSimUpdateFailed: Story = {
	args: {
		outdatedNativeSim: true,
		nativeSimUpdate: { latest: 'v1.2.3', status: { state: 'error', error: 'The update failed: no newer release' }, onUpdate: fn() },
	},
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('native-sim-update'));
		await expect(await within(document.body).findByTestId('native-sim-update-error')).toBeInTheDocument();
	},
};
