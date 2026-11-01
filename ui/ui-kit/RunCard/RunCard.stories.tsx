import type { Meta, StoryObj } from '@storybook/react-vite';
import { Button } from '@ui-kit/Button';
import { ButtonGroup } from '@ui-kit/ButtonGroup';
import { Icon } from '@ui-kit/Icon';
import { Pill } from '@ui-kit/Pill';
import { ProgressTrackerBar } from '@ui-kit/ProgressTrackerDialog';
import { NumberField } from '@ui-kit/TextField';

import { RunCard } from './RunCard';

const Iterations = () => (
	<div className="flex w-full items-end gap-2">
		<NumberField id="story-iterations" label="Iterations" size="md" value={12500} onChange={() => {}} className="min-w-0 flex-1" />
		<Button variant="primary" size="md" aria-pressed>
			12.5K
		</Button>
		<Button variant="outline-primary" size="md" aria-pressed={false}>
			25K
		</Button>
	</div>
);

const RunRow = ({ running = false }: { running?: boolean }) => (
	<ButtonGroup gap className="w-full justify-end">
		{running ? (
			<Button variant="cancel" className="w-full">
				<Icon name="xmark" style="base" />
				Stop
			</Button>
		) : (
			<Button className="w-full">Simulate</Button>
		)}
		<Button variant="outline-primary" aria-label="Stat Weights">
			<Icon name="scale-balanced" />
		</Button>
	</ButtonGroup>
);

const LastRunBody = ({ reference = true }: { reference?: boolean }) => (
	<div className="flex w-full flex-col gap-2">
		<div className="flex items-center justify-between gap-2 text-sm">
			<span>Last run</span>
			<span className="flex items-center gap-2 text-success">
				<span className="size-2 rounded-full bg-current" aria-hidden="true" />
				Up to date
			</span>
		</div>
		<div className="flex flex-col items-start gap-2">
			<div className="text-left font-bold">
				<span className="text-2xl leading-none">
					311.87<span className="text-ui font-normal"> DPS</span>
				</span>
				<span className="text-ui font-normal">
					(<Icon name="plus-minus" size="xs" />
					8)
				</span>
				{reference && (
					<Pill tone="positive" className="flex">
						<span>+4.21 (1.37%)</span>vs ref
					</Pill>
				)}
			</div>
		</div>
		<div className="flex flex-wrap items-center justify-between gap-2 font-normal">
			{reference ? (
				<ButtonGroup size="sm">
					<Button variant="outline-primary">
						<Icon name="arrows-rotate" />
						Swap
					</Button>
					<Button variant="outline-primary">
						<Icon name="xmark" style="base" />
						Unset
					</Button>
				</ButtonGroup>
			) : (
				<Button variant="outline-primary" size="sm">
					<Icon name="map-pin" style="base" />
					Set as reference
				</Button>
			)}
			<Button variant="link" size="sm">
				Results
				<Icon name="arrow-right" />
			</Button>
		</div>
	</div>
);

const RunningBody = () => (
	<div className="flex w-full flex-col gap-2">
		<div className="flex items-center justify-between gap-2 text-sm">
			<span>Simulating</span>
			<span className="tabular-nums">00:44 remaining</span>
		</div>
		<div className="text-left font-bold">
			<span className="text-2xl leading-none">60.81</span>
			<span className="text-ui font-normal"> DPS</span>
		</div>
		<ProgressTrackerBar running progress={0.36} readout="4,500 / 12,500" className="w-full max-w-full" />
	</div>
);

const meta = {
	title: 'Kit/RunCard',
	component: RunCard,
	tags: ['autodocs'],
	args: {
		controls: <Iterations />,
		action: <RunRow />,
	},
	argTypes: {
		controls: { control: false },
		action: { control: false },
		children: { control: false },
	},
	render: args => (
		<div className="max-w-100">
			<RunCard {...args} />
		</div>
	),
} satisfies Meta<typeof RunCard>;

export default meta;
type Story = StoryObj<typeof meta>;

/** The sim sidebar after a run, before a reference is set. */
export const LastRun: Story = {
	args: {
		children: <LastRunBody reference={false} />,
	},
};

/** The sim sidebar after a run, with a reference set: its delta, and Swap and Unset. */
export const LastRunWithReference: Story = {
	args: {
		children: <LastRunBody />,
	},
};

/** The sim sidebar mid-run: Simulate turns into Stop, and the body shows live numbers. */
export const Running: Story = {
	args: {
		action: <RunRow running />,
		children: <RunningBody />,
	},
};

/** The Batch tab's run card: no iterations field, the batch picks its own. */
export const Batch: Story = {
	args: {
		controls: undefined,
		action: <Button className="w-full">Simulate Batch</Button>,
		children: (
			<dl className="mb-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2">
				<dt className="font-normal">Combinations</dt>
				<dd className="mb-0 text-right font-bold tabular-nums">12</dd>
				<dt className="font-normal">Minimum iterations</dt>
				<dd className="mb-0 flex items-center justify-end gap-2 font-bold tabular-nums">120,000</dd>
			</dl>
		),
	},
};

export const ControlsOnly: Story = {};

export const Loading: Story = { args: { loading: true, children: <LastRunBody /> } };
