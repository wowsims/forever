import { Spec } from '@generated/proto/common';
import { PlayerSpecs } from '@sim/player/specs/index';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, userEvent, within } from 'storybook/test';

import { SimTitleDropdown } from './SimTitleDropdown';

const meta = {
	title: 'App/Header/Components/SimTitleDropdown',
	component: SimTitleDropdown,
	tags: ['autodocs'],
	args: { currentSpec: PlayerSpecs.fromProto(Spec.SpecDpsWarrior) },
	argTypes: { currentSpec: { control: false } },
	decorators: [
		Story => (
			<div className="w-80 bg-background">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof SimTitleDropdown>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	play: async ({ canvasElement }) => {
		await userEvent.click(within(within(canvasElement).getByTestId('sim-link-dropdown')).getByTestId('sim-link'));
		await expect(await within(canvasElement.ownerDocument.body).findByTestId('sim-title-popup')).toBeTruthy();
	},
};
