import type { Meta, StoryObj } from '@storybook/react-vite';
import { Tooltip } from '@ui-kit/Tooltip';

import { ResultReferenceDiff } from './ResultReferenceDiff';

const TOOLTIP_ID = 'story-reference-diff';

const meta = {
	title: 'Features/Results/ResultMetricList/Components/ResultReferenceDiff',
	component: ResultReferenceDiff,
	tags: ['autodocs'],
	args: {
		diff: { text: '+16.27 (5.26%)', tone: 'positive', significance: 'Difference is significantly different (Z = 4.212).' },
		tooltipId: TOOLTIP_ID,
	},
	decorators: [
		Story => (
			<>
				<Story />
				<Tooltip id={TOOLTIP_ID} />
			</>
		),
	],
} satisfies Meta<typeof ResultReferenceDiff>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Gain: Story = {};

export const Loss: Story = {
	args: { diff: { text: '-16.27 (5.00%)', tone: 'negative', significance: 'Difference is significantly different (Z = 4.212).' } },
};

export const NotSignificant: Story = {
	args: { diff: { text: '+0.41 (0.13%)', tone: null, significance: 'Difference is not significantly different (Z = 0.104).' } },
};
