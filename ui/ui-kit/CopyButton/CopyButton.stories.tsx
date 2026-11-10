import type { Meta, StoryObj } from '@storybook/react-vite';
import { ToastArea, toastManager } from '@ui-kit/Toast';
import { expect, userEvent, waitFor, within } from 'storybook/test';

import { CopyButton } from './CopyButton';

const meta = {
	title: 'Kit/CopyButton',
	component: CopyButton,
	tags: ['autodocs'],
	args: {
		getContent: () => '{"player":{"name":"Player"}}',
	},
	argTypes: {
		variant: { control: 'select', options: [null, 'primary', 'outline-primary', 'link'] },
		text: { control: 'text' },
		tooltip: { control: 'text' },
		className: { control: 'text' },
		disabled: { control: 'boolean' },
	},
} satisfies Meta<typeof CopyButton>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const OutlinePrimary: Story = { args: { variant: 'outline-primary', text: 'Copy Export' } };
export const WithTooltip: Story = { args: { variant: 'primary', tooltip: 'Copies the whole sim as JSON.' } };
/** Nothing to copy yet, such as an export whose text is still being made. */
export const Disabled: Story = { args: { variant: 'primary', disabled: true } };

/** An insecure origin has no `navigator.clipboard`: the click toasts the payload for a manual copy. */
export const NoClipboard: Story = {
	args: { variant: 'primary' },
	beforeEach: () => {
		const own = Object.getOwnPropertyDescriptor(navigator, 'clipboard');
		Object.defineProperty(navigator, 'clipboard', { value: undefined, configurable: true });
		return () => {
			if (own) Object.defineProperty(navigator, 'clipboard', own);
			else Reflect.deleteProperty(navigator, 'clipboard');
			toastManager.close();
		};
	},
	render: args => (
		<>
			<CopyButton {...args} />
			<ToastArea manager={toastManager} inline />
		</>
	),
	play: async ({ canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('copy-button'));
		const payload = await within(canvasElement.ownerDocument.body).findByText('{"player":{"name":"Player"}}');
		await waitFor(() => expect(payload).toBeVisible());
	},
};
