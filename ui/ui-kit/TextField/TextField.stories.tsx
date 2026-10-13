import type { Meta, StoryObj } from '@storybook/react-vite';
import { Fragment, useState } from 'react';
import { expect, fn, userEvent, waitFor, within } from 'storybook/test';

import { NumberField, NumberListField, SearchField, StringField } from './TextField';
import type { StringFieldProps, TextFieldCommit } from './types';

const useStoryValue = <V,>(initial: NoInfer<V>, onChange: (next: V) => void) => {
	const [value, setValue] = useState(initial);
	return {
		value,
		onChange: (next: V) => {
			setValue(next);
			onChange(next);
		},
	};
};

const TextFieldStory = ({ value, onChange, ...props }: StringFieldProps) => <StringField {...props} {...useStoryValue(value, onChange)} />;

const meta = {
	title: 'Kit/TextField',
	component: StringField,
	tags: ['autodocs'],
	args: {
		id: 'text-field-story',
		label: 'Rotation Name',
		value: 'Default',
		onChange: fn(),
	},
	argTypes: {
		layout: { control: 'select', options: [undefined, 'inline', 'split'] },
		disabled: { control: 'boolean' },
		value: { control: false },
		className: { control: 'text' },
	},
	render: args => <TextFieldStory {...args} />,
} satisfies Meta<typeof StringField>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};

export const MediumSize: Story = { args: { inputSize: 'md' } };

const InvalidStory = ({ onChange }: { onChange: (next: number) => void }) => (
	<div className="flex flex-col gap-2">
		<StringField id="string-field-invalid" label="Rotation Name" invalid {...useStoryValue('Default', () => {})} />
		<NumberField id="number-field-invalid" label="Duration (s)" invalid {...useStoryValue(180, onChange)} />
	</div>
);

export const Invalid: Story = {
	render: args => <InvalidStory onChange={args.onChange as (next: number) => void} />,
	play: async ({ canvasElement }) => {
		const canvas = within(canvasElement);
		await expect(canvas.getByLabelText('Rotation Name').getAttribute('aria-invalid')).toBe('true');
		await expect(canvas.getByLabelText('Duration (s)').getAttribute('aria-invalid')).toBe('true');
	},
};

const NumberStory = ({ onChange }: { onChange: (next: number) => void }) => (
	<div className="flex flex-col gap-2">
		<NumberField id="number-field-int" label="Iterations" {...useStoryValue(12500, onChange)} />
		<NumberField id="number-field-float" label="Swing Speed" float {...useStoryValue(2.6, onChange)} />
		<NumberField id="number-field-positive" label="Distance From Target" float positive {...useStoryValue(5, onChange)} />
		<NumberField id="number-field-zero" label="Min ILvl" showZeroes={false} {...useStoryValue(0, onChange)} />
		<NumberField id="number-field-split" label="Reaction Time" layout="split" {...useStoryValue(150, onChange)} />
	</div>
);

export const NumberPreset: Story = {
	render: args => <NumberStory onChange={args.onChange as (next: number) => void} />,
	play: async ({ args, canvasElement }) => {
		const field = within(canvasElement).getByLabelText('Iterations');
		await userEvent.clear(field);
		await userEvent.type(field, '3000');
		await expect(args.onChange).not.toHaveBeenCalled();
		await userEvent.tab();
		await expect(args.onChange).toHaveBeenLastCalledWith(3000);
	},
};

const StringStory = ({ onChange }: { onChange: (next: string) => void }) => (
	<StringField id="string-field-story" label="Schedule" {...useStoryValue('0s, 60s', onChange)} />
);

export const StringPreset: Story = {
	render: args => <StringStory onChange={args.onChange} />,
};

const NumberListStory = ({ onChange }: { onChange: (next: Array<number>) => void }) => (
	<NumberListField id="number-list-field-story" placeholder="20, 40, ..." {...useStoryValue([20, 40], onChange)} />
);

export const NumberListPreset: Story = {
	render: args => <NumberListStory onChange={args.onChange as unknown as (next: Array<number>) => void} />,
};

const SearchStory = ({ onChange }: { onChange: (next: string) => void }) => (
	<SearchField id="search-field-story" placeholder="Search items" debounceMs={150} clear={{ label: 'Clear' }} {...useStoryValue('', onChange)} />
);

export const SearchPreset: Story = {
	render: args => <SearchStory onChange={args.onChange} />,
	play: async ({ args, canvasElement }) => {
		const canvas = within(canvasElement);
		await userEvent.type(canvas.getByRole('textbox'), 'Ashkandi');
		await expect(args.onChange).not.toHaveBeenCalled();
		await waitFor(() => expect(args.onChange).toHaveBeenCalledWith('Ashkandi'));
		await userEvent.click(canvas.getByRole('button', { name: 'Clear' }));
		await expect(args.onChange).toHaveBeenLastCalledWith('');
	},
};

const COMMITS: ReadonlyArray<[string, TextFieldCommit]> = [
	['blur', 'blur'],
	['input', 'input'],
	['debounced', { debounceMs: 300 }],
];

const MatrixCell = ({ label, commit, onChange }: { label?: string; commit: TextFieldCommit; onChange: (next: string) => void }) => {
	const value = useStoryValue('Typhoon', onChange);
	if (commit === 'blur') return <StringField label={label} {...value} />;
	return <SearchField label={label} debounceMs={commit === 'input' ? undefined : commit.debounceMs} {...value} />;
};

export const Matrix: Story = {
	render: args => (
		<div className="grid grid-cols-[auto_1fr_1fr] items-center gap-2">
			<span />
			<span>labelled</span>
			<span>unlabelled</span>
			{COMMITS.map(([name, commit]) => (
				<Fragment key={name}>
					<span>{name}</span>
					<MatrixCell label="Rotation Name" commit={commit} onChange={args.onChange} />
					<MatrixCell commit={commit} onChange={args.onChange} />
				</Fragment>
			))}
		</div>
	),
};
