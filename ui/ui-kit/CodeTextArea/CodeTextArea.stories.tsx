import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';

import { CodeTextArea } from './CodeTextArea';

const JSON_TEXT = JSON.stringify({ player: { name: 'Player', race: 'RaceHuman', class: 'ClassWarrior', talentsString: '30305001302-05' } }, null, 2);
const SIML_TEXT = 'siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n  cast arcane_shot\n';
const itemIds = (line: number) => Array.from({ length: 24 }, (_, i) => 19000 + line * 24 + i).join(',');
const LONG_TEXT = Array.from({ length: 300 }, (_, line) => `{"line":${line + 1},"items":[${itemIds(line)}]}`).join('\n');

const meta = {
	title: 'Kit/CodeTextArea',
	component: CodeTextArea,
	tags: ['autodocs'],
	args: { value: JSON_TEXT, className: 'h-64 w-full' },
	argTypes: {
		readOnly: { control: 'boolean' },
		className: { control: 'text' },
	},
	render: function Render(args) {
		const [value, setValue] = useState(args.value);
		return <CodeTextArea {...args} value={value} onChange={setValue} />;
	},
} satisfies Meta<typeof CodeTextArea>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Siml: Story = { args: { value: SIML_TEXT } };
/** Hundreds of lines wider than the box: the numbers scroll with the text, and the position stays clear of both scrollbars. */
export const LongFile: Story = { args: { value: LONG_TEXT } };
export const Empty: Story = { args: { value: '' } };
