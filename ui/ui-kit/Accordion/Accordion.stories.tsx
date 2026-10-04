import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fireEvent, within } from 'storybook/test';

import { Accordion, type AccordionVariant } from './Accordion';
import { AccordionItem } from './AccordionItem';

interface AccordionStoryProps {
	initial: Array<string>;
	variant?: AccordionVariant;
}

const AccordionStory = ({ initial, variant }: AccordionStoryProps) => {
	const [open, setOpen] = useState(initial);
	return (
		<Accordion value={open} onValueChange={setOpen} variant={variant}>
			<AccordionItem value="consumes" title="Consumables" testId="story-consumes">
				<p>Flask of Supreme Power, Elixir of Greater Firepower</p>
			</AccordionItem>
			<AccordionItem value="buffs" title="Raid buffs" testId="story-buffs">
				<p>Arcane Intellect, Mark of the Wild, Blessing of Kings</p>
			</AccordionItem>
			<AccordionItem value="debuffs" title="Debuffs" testId="story-debuffs">
				<p>Curse of Elements, Improved Scorch</p>
			</AccordionItem>
		</Accordion>
	);
};

const meta = {
	title: 'Kit/Accordion',
	component: AccordionStory,
	tags: ['autodocs'],
	args: { initial: ['consumes'] },
	argTypes: { variant: { control: 'inline-radio', options: ['divided', 'boxed'] } },
} satisfies Meta<typeof AccordionStory>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const Multiple: Story = { args: { initial: ['consumes', 'debuffs'] } };

export const Boxed: Story = {
	args: { variant: 'boxed' },
	render: args => (
		<div className="w-62">
			<AccordionStory {...args} />
		</div>
	),
};

export const Opens: Story = {
	args: { initial: [] },
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.queryByText(/Curse of Elements/)).toBeNull();

		await fireEvent.click(canvas.getByTestId('story-debuffs-trigger'));

		await expect(canvas.getByText(/Curse of Elements/)).toBeTruthy();
	},
};
