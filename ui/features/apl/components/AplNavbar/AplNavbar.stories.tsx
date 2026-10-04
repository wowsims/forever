import { Tabs } from '@base-ui/react/tabs';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withRotationTypeHost } from '@testing/storybook/fixtures/apl';

import { AplNavbar } from './AplNavbar';

const meta = {
	title: 'Features/Apl/AplNavbar',
	component: AplNavbar,
	tags: ['autodocs'],
	decorators: [
		Story => (
			<Tabs.Root defaultValue="apl-priority-list">
				<Story />
			</Tabs.Root>
		),
	],
} satisfies Meta<typeof AplNavbar>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
	decorators: [withRotationTypeHost({ type: APLRotationType.TypeAPL, hasSimple: true })],
};
