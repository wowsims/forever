import { SimHostProvider } from '@sim/context/SimHostContext';
import type { BulkMode } from '@sim/state/sim_store';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { makeBulkTopResultsHost } from '@testing/storybook/fixtures/bulk-top-results';
import { useState } from 'react';
import { fn } from 'storybook/test';

import { BulkTopResultsField, BulkTopResultsFieldView } from './BulkTopResultsField';

const meta = {
	title: 'Features/Bulk/BulkSettings/BulkTopResultsField',
	component: BulkTopResultsFieldView,
	tags: ['autodocs'],
	args: { value: 5, max: 5000, onChange: fn() },
} satisfies Meta<typeof BulkTopResultsFieldView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Five: Story = {};

export const All: Story = { args: { value: 0 } };

// The container over a real store, so ticking All and typing a number change the stored pick.
const Connected = ({ mode, topResults }: { mode: BulkMode; topResults: number }) => {
	const [host] = useState(() => makeBulkTopResultsHost(mode, topResults));
	return (
		<SimHostProvider host={host}>
			<BulkTopResultsField />
		</SimHostProvider>
	);
};

export const ConnectedGear: Story = { render: () => <Connected mode="gear" topResults={5} /> };

export const ConnectedBuilds: Story = { render: () => <Connected mode="builds" topResults={0} /> };
