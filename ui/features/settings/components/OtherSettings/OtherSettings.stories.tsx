import type { Player } from '@sim/player/player';
import type { InputConfig } from '@sim/spec_config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { withSimHost } from '@testing/storybook/fixtures/models';

import { OtherSettings } from './OtherSettings';

const values = { inFrontOfTarget: false, healingModel: 0 };
const unsubscribed = () => () => () => {};

const INPUTS = [
	{
		type: 'boolean',
		id: 'story-in-front-of-target',
		label: 'In Front of Target',
		labelTooltip: 'Stand in front of the target, where attacks can be parried and blocked.',
		storeSubscribe: unsubscribed,
		getValue: () => values.inFrontOfTarget,
		setValue: (_player: unknown, next: boolean) => {
			values.inFrontOfTarget = next;
		},
	},
	{
		type: 'number',
		id: 'story-incoming-hps',
		label: 'Incoming HPS',
		storeSubscribe: unsubscribed,
		getValue: () => values.healingModel,
		setValue: (_player: unknown, next: number) => {
			values.healingModel = next;
		},
	},
] as unknown as ReadonlyArray<InputConfig<Player<any>>>;

const meta = {
	title: 'Features/Settings/OtherSettings',
	component: OtherSettings,
	tags: ['autodocs'],
	decorators: [withSimHost()],
	args: { inputs: INPUTS, itemSlots: [] },
	argTypes: { inputs: { control: false }, itemSlots: { control: false } },
} satisfies Meta<typeof OtherSettings>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The spec's other inputs, laid out split; a spec without item swap slots shows no swap picker. */
export const Default: Story = {};
