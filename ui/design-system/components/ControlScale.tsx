import { Unstyled } from '@storybook/addon-docs/blocks';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField } from '@ui-kit/ChoiceField';
import type { ControlSize } from '@ui-kit/Field';
import { Input } from '@ui-kit/FormControl';
import { Fragment } from 'react';

const SIZES: ReadonlyArray<ControlSize> = ['xs', 'sm', 'md'];
const OPTIONS = [{ value: 0, label: 'Beast' }];

const HEIGHTS = [
	['xs', '1.625rem plus 2px', '24.75px', '28px'],
	['sm', '2.125rem plus 2px', '31.75px', '36px'],
	['md', '2.3125rem plus 2px', '34.38px', '39px'],
];

const PADDINGS = [
	['xs', '3.5px', '2.19px', '3.5px'],
	['sm', '7px', '5.69px', '7px'],
	['md', '8.31px', '7px', '7px'],
];

const Table = ({ head, rows }: { head: ReadonlyArray<string>; rows: ReadonlyArray<ReadonlyArray<string>> }) => (
	<Unstyled>
		<table className="mb-6 w-full text-sm">
			<thead>
				<tr className="border-b border-border">
					{head.map(cell => (
						<th key={cell} className="p-2">
							{cell}
						</th>
					))}
				</tr>
			</thead>
			<tbody>
				{rows.map(([size, ...cells]) => (
					<tr key={size} className="border-b border-border-muted">
						<td className="p-2">
							<code>{size}</code>
						</td>
						{cells.map((cell, index) => (
							<td key={index} className="p-2">
								{cell}
							</td>
						))}
					</tr>
				))}
			</tbody>
		</table>
	</Unstyled>
);

export const ControlHeights = () => <Table head={['Size', 'Height', 'At 14px', 'At 16px (fhd)']} rows={HEIGHTS} />;

export const ControlPaddings = () => (
	<Table head={['Size', 'Padding-y, inputs, at 14px', 'Padding-y, buttons, at 14px', 'Padding-x, at 14px']} rows={PADDINGS} />
);

export const ControlScale = () => (
	<Unstyled>
		<div className="mb-6 grid grid-cols-[auto_1fr_1fr_auto_auto] items-center gap-2">
			{SIZES.map(size => (
				<Fragment key={size}>
					<code>{size}</code>
					<Input size={size} aria-label={`Input ${size}`} defaultValue="0.1" />
					<ChoiceField size={size} ariaLabel={`Select ${size}`} options={OPTIONS} value={0} onChange={() => {}} />
					<Button size={size}>Apply</Button>
					<BooleanPicker size={size} value onChange={() => {}} />
				</Fragment>
			))}
		</div>
	</Unstyled>
);
