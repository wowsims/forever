import type { Meta, StoryObj } from '@storybook/react-vite';
import { StoryBoolean, StoryEnum, StoryNumber } from '@testing/storybook/fixtures/settings_reforge_bulk';
import { expect, fn, userEvent, within } from 'storybook/test';

import { ReforgeBreakpointLimitsView } from './ReforgeBreakpointLimits';
import { ReforgeFrozenSlotsView } from './ReforgeFrozenSlots';
import { ReforgeSettingsPanelView } from './ReforgeSettingsPanel';
import { ReforgeStatCapRowView } from './ReforgeStatCapRow';
import { ReforgeStatCapsView } from './ReforgeStatCaps';

const toggle = (id: string, label: string, initial = false) => (
	<StoryBoolean id={`story-${id}`} className="mb-2" label={label} layout="inline" initial={initial} />
);

const STAT_CAPS = (
	<ReforgeStatCapsView onReset={fn()}>
		{(['Hit', 'Expertise'] as const).map(statName => (
			<ReforgeStatCapRowView
				key={statName}
				statName={statName}
				percentage={<StoryNumber float showZeroes={false} positive inputClassName="pr-2 pl-2 text-right" initial={9} />}
				undershoot={<StoryBoolean label="" />}
			/>
		))}
	</ReforgeStatCapsView>
);

const SOFT_CAPS = (
	<>
		{toggle('soft-caps', 'Use soft cap breakpoints', true)}
		<ReforgeBreakpointLimitsView
			rows={[
				{
					key: 'hit',
					statName: 'Hit',
					picker: (
						<StoryEnum
							label=""
							values={[
								{ name: 'No limit set', value: 0 },
								{ name: '9%', value: 9 },
							]}
						/>
					),
				},
			]}
		/>
	</>
);

const FREEZE_SLOTS = (
	<>
		{toggle('freeze-slots', 'Freeze item slots', true)}
		<ReforgeFrozenSlotsView
			hidden={false}
			pickers={['Head', 'Neck', 'Shoulders', 'Back'].map(slot => (
				<StoryBoolean key={slot} id={`story-freeze-${slot}`} label={slot} layout="inline" />
			))}
		/>
	</>
);

const meta = {
	title: 'Features/Reforge/ReforgePanel/ReforgeSettingsPanel',
	component: ReforgeSettingsPanelView,
	tags: ['autodocs'],
	args: {
		customEpToggle: toggle('custom-ep', 'Use custom EP values', true),
		showEpHelp: false,
		statCaps: STAT_CAPS,
		softCaps: SOFT_CAPS,
		freezeSlots: FREEZE_SLOTS,
		savedEpWeights: null,
		onEditWeights: fn(),
	},
	argTypes: {
		customEpToggle: { control: false },
		statCaps: { control: false },
		softCaps: { control: false },
		freezeSlots: { control: false },
		savedEpWeights: { control: false },
	},
	decorators: [
		Story => (
			<div className="max-w-86">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof ReforgeSettingsPanelView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const CustomEpValues: Story = {
	play: async ({ args, canvasElement }) => {
		await userEvent.click(within(canvasElement).getByTestId('reforge-edit-weights'));
		await expect(args.onEditWeights).toHaveBeenCalledOnce();
	},
};

/** With the custom EP values off, the panel explains how to edit the weights and has no cap table. */
export const PresetEpValues: Story = {
	args: { customEpToggle: toggle('custom-ep-off', 'Use custom EP values'), showEpHelp: true, statCaps: null, softCaps: null },
};
