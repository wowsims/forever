import type { Meta, StoryObj } from '@storybook/react-vite';
import { castRotationRow, castRowItem, MORTAL_STRIKE, rotationTrackStyle } from '@testing/storybook/fixtures/results-views';

import { CastItem } from './CastItem';
import { GcdExtensionItem } from './GcdExtensionItem';

const meta = {
	title: 'Features/Results/Timeline/Components/GcdExtensionItem',
	component: GcdExtensionItem,
	tags: ['autodocs'],
	args: { item: { kind: 'gcdExtension', start: 1.5, end: 3 } },
	decorators: [
		Story => (
			<div className="ui-timeline-row" style={rotationTrackStyle(castRotationRow(MORTAL_STRIKE, []), 6)}>
				<div className="ui-timeline-row-track">
					<Story />
				</div>
			</div>
		),
	],
} satisfies Meta<typeof GcdExtensionItem>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Alone: Story = {};

export const AfterItsCast: Story = {
	render: args => (
		<>
			<CastItem item={castRowItem(0.5, { end: 1.5 })} index={0} iconUrl={MORTAL_STRIKE.iconUrl} />
			<GcdExtensionItem {...args} />
		</>
	),
};
