import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { MORTAL_STRIKE } from '@testing/storybook/fixtures/results-views';
import { Tooltip } from '@ui-kit/Tooltip';
import { WowheadIcon } from '@ui-kit/WowheadIcon';
import { expect, fn, userEvent, within } from 'storybook/test';

import { RotationRowLabel } from './RotationRowLabel';

const HIDE_TOOLTIP_ID = 'story-rotation-label-hide';

const meta = {
	title: 'Features/Results/Timeline/Components/RotationRowLabel',
	component: RotationRowLabel,
	tags: ['autodocs'],
	args: {
		text: MORTAL_STRIKE.name,
		icon: <WowheadIcon className="ui-timeline-row-icon" iconUrl={MORTAL_STRIKE.iconUrl} />,
		onHide: fn(),
		hideTooltipId: HIDE_TOOLTIP_ID,
	},
	argTypes: { icon: { control: false } },
	decorators: [
		Story => (
			<div className="text-white [--label-w:12rem]">
				<Story />
				<Tooltip id={HIDE_TOOLTIP_ID} />
			</div>
		),
	],
} satisfies Meta<typeof RotationRowLabel>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Hideable: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(
			within(canvasElement).getByRole('button', { name: i18n.t('results_tab.details.timeline.hide_row_named', { row: MORTAL_STRIKE.name }) }),
		);
		await expect(args.onHide).toHaveBeenCalledOnce();
	},
};

export const SectionHeader: Story = { args: { text: 'Player', icon: undefined, onHide: undefined, header: true } };

export const Measuring: Story = { args: { measuring: true } };
