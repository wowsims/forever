import { APLValidation } from '@generated/proto/api';
import { LogLevel } from '@generated/proto/common';
import type { Meta, StoryObj } from '@storybook/react-vite';

import { AplValidationsView } from './AplValidations';

const validation = (logLevel: LogLevel, text: string) => APLValidation.create({ logLevel, validation: text });

const meta = {
	title: 'Features/Apl/AplValidations',
	component: AplValidationsView,
	tags: ['autodocs'],
} satisfies Meta<typeof AplValidationsView>;

export default meta;
type Story = StoryObj<typeof meta>;

/** Nothing to say: the button is in the header but not displayed. */
export const None: Story = { args: { validations: [] } };

export const Information: Story = {
	args: { validations: [validation(LogLevel.Information, 'Bloodthirst is only cast while the target is above 20% health.')] },
};

export const Warnings: Story = {
	args: {
		validations: [
			validation(LogLevel.Information, 'Bloodthirst is only cast while the target is above 20% health.'),
			validation(LogLevel.Warning, 'Whirlwind is never castable: its cooldown is longer than the fight.'),
		],
	},
};

/** The highest level sets the icon and colour; the tooltip lists every level under its own header. */
export const Errors: Story = {
	args: {
		validations: [
			validation(LogLevel.Warning, 'Whirlwind is never castable: its cooldown is longer than the fight.'),
			validation(LogLevel.Error, "Group reference 'Execute' not found."),
			validation(LogLevel.Error, "Variable reference 'rage_pooled' not found."),
		],
	},
};
