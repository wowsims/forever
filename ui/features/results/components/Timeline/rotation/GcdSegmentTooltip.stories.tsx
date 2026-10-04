import type { Meta, StoryObj } from '@storybook/react-vite';
import { castLog } from '@testing/storybook/fixtures/results-views';

import { GcdSegmentTooltip } from './GcdSegmentTooltip';

const meta = {
	title: 'Features/Results/Timeline/Components/GcdSegmentTooltip',
	component: GcdSegmentTooltip,
	tags: ['autodocs'],
	args: { log: castLog(6) },
	argTypes: { log: { control: false } },
	decorators: [
		Story => (
			<div className="ui-timeline-hover-tooltip relative w-fit">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof GcdSegmentTooltip>;

export default meta;
type Story = StoryObj<typeof meta>;

export const FullGcd: Story = {};

export const ShortenedGcd: Story = { args: { log: castLog(6, { gcd: 1 }) } };
