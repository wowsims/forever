import i18n from '@i18n/config';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { ProgressTrackerDialog, type ProgressTrackerHandle } from '@ui-kit/ProgressTrackerDialog';

import { BulkProgressSteps } from './BulkProgressSteps';

const CURRENT_ENCOUNTER = { encounter: i18n.t('bulk_tab.settings.encounters.current'), encounterSummary: '3:00 ±5s · 1 target' };
const DEFAULT_60 = { encounter: 'Default - 60', encounterSummary: '5:00 ±30s · 3 targets' };
const CURRENT_GROUP = {
	settings: i18n.t('bulk_tab.results.current_settings'),
	talents: i18n.t('bulk_tab.results.current_talents'),
	rotation: i18n.t('bulk_tab.results.current_rotation'),
};
const FURY_GROUP = { settings: 'Pre-BiS - Fury', talents: 'Fury', rotation: 'Fury' };
const ARMS_GROUP = { settings: 'Pre-BiS - Arms', talents: 'Arms', rotation: 'Arms - Autocast on use' };
const ROWS = [CURRENT_ENCOUNTER, DEFAULT_60].flatMap(encounter => [CURRENT_GROUP, FURY_GROUP, ARMS_GROUP].map(group => ({ ...encounter, ...group })));

const pushProgress = (handle: ProgressTrackerHandle | null) =>
	handle?.setProgress({ title: i18n.t('bulk_tab.progress.high_iteration_rounds'), current: 10, total: 26 });

const meta = {
	title: 'Features/Bulk/BulkProgressSteps',
	component: BulkProgressSteps,
	tags: ['autodocs'],
	args: { rows: ROWS, running: 1, progress: 0.4 },
} satisfies Meta<typeof BulkProgressSteps>;

export default meta;
type Story = StoryObj<typeof meta>;

export const MidRun: Story = {};

/** A Builds batch in the progress dialog, a row per encounter and bundle group: one done, one 40% through, the rest pending. */
export const BuildsDialog: Story = {
	render: args => (
		<ProgressTrackerDialog
			open
			title={i18n.t('bulk_tab.progress.title')}
			state={{ stage: 'running' }}
			hasProgressBar
			belowBar={<BulkProgressSteps {...args} />}
			onCancel={() => {}}
			ref={pushProgress}
		/>
	),
};
