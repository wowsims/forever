import type { Meta, StoryObj } from '@storybook/react-vite';

import { Table } from './Table';
import { TableCell } from './TableCell';
import { TableHeaderCell } from './TableHeaderCell';
import { TableRow } from './TableRow';

const ROWS = [
	{ name: 'Fireball', casts: 42, dps: 812.4 },
	{ name: 'Scorch', casts: 18, dps: 201.9 },
	{ name: 'Fire Blast', casts: 9, dps: 143.2 },
	{ name: 'Ignite', casts: 31, dps: 120.7 },
];

const meta = {
	title: 'Kit/Table',
	component: Table,
	tags: ['autodocs'],
	args: {
		striped: false,
		hover: false,
		dense: false,
		children: (
			<>
				<thead>
					<TableRow>
						<TableHeaderCell>Ability</TableHeaderCell>
						<TableHeaderCell>Casts</TableHeaderCell>
						<TableHeaderCell>DPS</TableHeaderCell>
					</TableRow>
				</thead>
				<tbody>
					{ROWS.map(row => (
						<TableRow key={row.name}>
							<TableCell>{row.name}</TableCell>
							<TableCell className="text-right">{row.casts}</TableCell>
							<TableCell className="text-right">{row.dps.toFixed(1)}</TableCell>
						</TableRow>
					))}
				</tbody>
			</>
		),
	},
	argTypes: {
		children: { control: false },
		className: { control: 'text' },
		striped: { control: 'select', options: [false, true, 'solid'] },
	},
} satisfies Meta<typeof Table>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
export const Striped: Story = { args: { striped: true } };
export const Hover: Story = { args: { hover: true } };
export const Dense: Story = { args: { dense: true } };
export const Plain: Story = { args: { variant: 'plain' } };
export const RowRulesOnly: Story = { args: { variant: 'rows' } };
export const Small: Story = { args: { size: 'sm' } };
export const Flush: Story = { args: { variant: 'plain', flush: true } };
export const WrappingHeader: Story = {
	args: {
		className: 'max-w-64',
		children: (
			<>
				<thead>
					<TableRow>
						<TableHeaderCell>Ability</TableHeaderCell>
						<TableHeaderCell wrap>Damage per second</TableHeaderCell>
						<TableHeaderCell>Casts per minute</TableHeaderCell>
					</TableRow>
				</thead>
				<tbody>
					{ROWS.map(row => (
						<TableRow key={row.name}>
							<TableCell>{row.name}</TableCell>
							<TableCell className="text-right">{row.dps.toFixed(1)}</TableCell>
							<TableCell className="text-right">{row.casts}</TableCell>
						</TableRow>
					))}
				</tbody>
			</>
		),
	},
};
