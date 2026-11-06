import type { Meta, StoryObj } from '@storybook/react-vite';
import { expect, fn, userEvent, within } from 'storybook/test';

import { PlannerSections, type PlannerSectionsProps, usePlannerSectionState } from './PlannerSections';

const Sections = (props: PlannerSectionsProps) => (
	<div className="w-80">
		<PlannerSections {...props} state={usePlannerSectionState()} />
	</div>
);

const meta = {
	title: 'Features/EncounterPlanner/PlannerSections',
	component: PlannerSections,
	tags: ['autodocs'],
	args: {
		testId: 'story-sections',
		state: { toggles: {}, onToggle: fn() },
		sections: [
			{ id: 'timing', title: 'Timing', summary: '72s–90s', children: <p>Start 72s, end 90s</p> },
			{ id: 'repeat', title: 'Repeat', summary: 'none', startsClosed: true, children: <p>Repeat every 0s</p> },
			{ id: 'damage', title: 'Damage', summary: 'Fire 500–700', forcedOpen: true, children: <p>Max damage 0</p> },
		],
	},
	argTypes: { sections: { control: false } },
	render: args => <Sections {...args} />,
} satisfies Meta<typeof PlannerSections>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

// A closed section names its values beside its title; an issue keeps Damage open after a click.
export const SummaryAndIssue: Story = {
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByRole('button', { name: /Repeat · none|panel\.section_summary$/ })).toBeVisible();
		await userEvent.click(canvas.getByRole('button', { name: 'Damage' }));
		await expect(canvas.getByText('Max damage 0')).toBeVisible();
	},
};
