import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { useEffect, useState } from 'react';
import { expect, userEvent, within } from 'storybook/test';

import { createToastManager } from './manager';
import { ToastArea, type ToastAreaProps } from './ToastArea';
import type { ToastVariant } from './types';

const BODIES: Record<ToastVariant, string> = {
	info: 'A newer version of the sim is available.',
	success: 'Import successful!',
	error: 'The sim failed to start.',
	warning: 'This spec is still in development.',
};

interface ToastStoryProps extends Omit<ToastAreaProps, 'manager'> {
	variant: ToastVariant;
}

const ToastStory = ({ variant, ...props }: ToastStoryProps) => {
	const [manager] = useState(createToastManager);
	useEffect(() => {
		const id = manager.add({ variant, body: BODIES[variant], autohide: false });
		return () => manager.close(id);
	}, [manager, variant]);
	return <ToastArea {...props} manager={manager} />;
};

const InlineStory = (props: ToastStoryProps) => (
	<div className="w-86">
		<p>The area renders in flow, inside this column:</p>
		<ToastStory {...props} inline />
	</div>
);

const StackStory = (props: Omit<ToastAreaProps, 'manager'>) => {
	const [manager] = useState(createToastManager);
	const [count, setCount] = useState(0);
	return (
		<>
			<Button
				onClick={() => {
					const next = count + 1;
					setCount(next);
					manager.add({ variant: 'info', body: `Toast ${next}`, autohide: false });
				}}>
				Add toast
			</Button>
			<ToastArea {...props} manager={manager} />
		</>
	);
};

const meta = {
	title: 'Kit/Toast',
	component: ToastStory,
	tags: ['autodocs'],
	args: { variant: 'info' },
	argTypes: {
		variant: { control: 'select', options: Object.keys(BODIES) },
	},
} satisfies Meta<typeof ToastStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Info: Story = {};

export const Success: Story = {
	args: { variant: 'success' },
};

export const Error: Story = {
	args: { variant: 'error' },
};

export const Warning: Story = {
	args: { variant: 'warning' },
};

export const Inline: Story = {
	render: args => <InlineStory {...args} />,
};

export const Stack: Story = {
	render: args => <StackStory {...args} />,
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		const add = canvas.getByRole('button', { name: 'Add toast' });
		await userEvent.click(add);
		await userEvent.click(add);
		await userEvent.click(add);
		const toasts = await canvas.findAllByTestId('sim-toast');
		await expect(toasts.map(toast => within(toast).getByTestId('sim-toast-body').textContent)).toEqual(['Toast 3', 'Toast 2', 'Toast 1']);
		const portal = canvasElement.querySelector('[data-testid="storybook-portal"]');
		await expect(toasts.every(toast => portal?.contains(toast))).toBe(true);
	},
};
