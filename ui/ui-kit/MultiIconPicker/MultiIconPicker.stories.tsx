import type { Meta, StoryObj } from '@storybook/react-vite';
import { KEPT_TOOLTIP } from '@testing/storybook/fixtures/kept';
import type { ResolvedIcon } from '@ui-kit/Field/types';
import { IconEnumPickerDirection } from '@ui-kit/IconEnumPicker';
import { IconPicker } from '@ui-kit/IconPicker';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { MultiIconPicker, type MultiIconPickerProps } from './MultiIconPicker';

const ICONS = 'https://wow.zamimg.com/images/wow/icons/large';
const icon = (file: string, name: string, spellId: number): ResolvedIcon => ({
	iconUrl: `${ICONS}/${file}.jpg`,
	name,
	href: `https://www.wowhead.com/classic/spell=${spellId}`,
});

const BUFFS = [
	icon('spell_holy_fistofjustice', 'Blessing of Might', 25291),
	icon('spell_holy_sealofwisdom', 'Blessing of Wisdom', 25290),
	icon('spell_nature_lightningshield', 'Trueshot Aura', 20906),
];

const MultiIconPickerStory = ({
	onClear,
	keptTooltip,
	direction,
	numColumns,
	initial,
}: Pick<MultiIconPickerProps, 'onClear' | 'keptTooltip' | 'direction' | 'numColumns'> & { initial: Array<number> }) => {
	const [values, setValues] = useState(initial);
	const first = values.findIndex(value => value > 0);
	return (
		<MultiIconPicker
			label="Misc"
			icon={first >= 0 ? BUFFS[first] : undefined}
			active={first >= 0}
			keptTooltip={keptTooltip}
			direction={direction}
			numColumns={numColumns}
			onClear={() => {
				setValues(values.map(() => 0));
				onClear();
			}}>
			{BUFFS.map((buff, index) => (
				<IconPicker
					key={buff.name}
					states={2}
					icon={buff}
					value={values[index]}
					onChange={next => setValues(values.map((value, i) => (i === index ? next : value)))}
				/>
			))}
		</MultiIconPicker>
	);
};

const meta = {
	title: 'Kit/MultiIconPicker',
	component: MultiIconPicker,
	tags: ['autodocs'],
	args: {
		label: 'Misc',
		active: false,
		onClear: fn(),
		children: null,
	},
	argTypes: {
		icon: { control: false },
		children: { control: false },
	},
	render: ({ onClear }) => <MultiIconPickerStory onClear={onClear} initial={[0, 0, 0]} />,
} satisfies Meta<typeof MultiIconPicker>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Closed: Story = {};
export const Open: Story = {
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		const trigger = canvas.getByTestId('multi-icon-picker-root');
		await userEvent.hover(trigger);
		const body = within(canvasElement.ownerDocument.body);
		const options = await body.findAllByTestId('icon-picker-root');
		await expect(options).toHaveLength(BUFFS.length);
		await userEvent.click(options[1]);
		await expect(trigger.style.backgroundImage).toContain('spell_holy_sealofwisdom');
		await userEvent.click(body.getByTestId('icon-picker-clear'));
		await expect(args.onClear).toHaveBeenCalledOnce();
	},
};
export const Active: Story = {
	render: ({ onClear }) => <MultiIconPickerStory onClear={onClear} initial={[0, 1, 1]} />,
};
export const Kept: Story = {
	args: { keptTooltip: KEPT_TOOLTIP },
	render: ({ onClear, keptTooltip }) => <MultiIconPickerStory onClear={onClear} keptTooltip={keptTooltip} initial={[0, 1, 0]} />,
};
export const Vertical: Story = {
	args: { direction: IconEnumPickerDirection.Vertical, numColumns: 2 },
	render: ({ onClear, direction, numColumns }) => (
		<MultiIconPickerStory onClear={onClear} direction={direction} numColumns={numColumns} initial={[1, 0, 0]} />
	),
};
