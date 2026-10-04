import type { Meta, StoryObj } from '@storybook/react-vite';
import { delayRowItem } from '@testing/storybook/fixtures/results-views';

import { DelayTooltip } from './DelayTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/DelayTooltip',
	component: DelayTooltip,
	tags: ['autodocs'],
	args: { log: delayRowItem(8.5, 8.85).log },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof DelayTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const AutoAttackDelayed: Story = {};
