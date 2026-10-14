import type { Meta, StoryObj } from '@storybook/react-vite';
import { useState } from 'react';
import { expect, fn, userEvent, within } from 'storybook/test';

import { SettingsDialogView } from './SettingsDialog';

const meta = {
	title: 'App/SettingsDialog/SettingsDialog',
	component: SettingsDialogView,
	tags: ['autodocs'],
	args: {
		open: true,
		onOpenChange: fn(),
		fixedRngSeed: 0,
		onFixedRngSeedChange: fn(),
		lastUsedRngSeed: 1839204711,
		language: 'en',
		onLanguageChange: fn(),
		showThreatMetrics: false,
		onShowThreatMetricsChange: fn(),
		showExperimental: false,
		onShowExperimentalChange: fn(),
		showPowerTools: false,
		onShowPowerToolsChange: fn(),
		showQuickSwap: true,
		onShowQuickSwapChange: fn(),
		wasmConcurrency: 4,
		onWasmConcurrencyChange: fn(),
		showConcurrency: true,
		maxWorkers: 8,
		firefox: false,
		onApplyDefaults: fn(),
	},
} satisfies Meta<typeof SettingsDialogView>;

export default meta;
type Story = StoryObj<typeof meta>;

type ViewProps = Parameters<typeof SettingsDialogView>[0];

const LiveSettingsDialog = (args: ViewProps) => {
	const [open, setOpen] = useState(args.open);
	const [fixedRngSeed, setFixedRngSeed] = useState(args.fixedRngSeed);
	const [language, setLanguage] = useState(args.language);
	const [showThreatMetrics, setShowThreatMetrics] = useState(args.showThreatMetrics);
	const [showExperimental, setShowExperimental] = useState(args.showExperimental);
	const [showPowerTools, setShowPowerTools] = useState(args.showPowerTools);
	const [showQuickSwap, setShowQuickSwap] = useState(args.showQuickSwap);
	const [wasmConcurrency, setWasmConcurrency] = useState(args.wasmConcurrency);
	const relay =
		<T,>(set: (next: T) => void, report: (next: T) => void) =>
		(next: T) => {
			set(next);
			report(next);
		};
	return (
		<SettingsDialogView
			{...args}
			open={open}
			onOpenChange={relay(setOpen, args.onOpenChange)}
			fixedRngSeed={fixedRngSeed}
			onFixedRngSeedChange={relay(setFixedRngSeed, args.onFixedRngSeedChange)}
			language={language}
			onLanguageChange={relay(setLanguage, args.onLanguageChange)}
			showThreatMetrics={showThreatMetrics}
			onShowThreatMetricsChange={relay(setShowThreatMetrics, args.onShowThreatMetricsChange)}
			showExperimental={showExperimental}
			onShowExperimentalChange={relay(setShowExperimental, args.onShowExperimentalChange)}
			showPowerTools={showPowerTools}
			onShowPowerToolsChange={relay(setShowPowerTools, args.onShowPowerToolsChange)}
			showQuickSwap={showQuickSwap}
			onShowQuickSwapChange={relay(setShowQuickSwap, args.onShowQuickSwapChange)}
			wasmConcurrency={wasmConcurrency}
			onWasmConcurrencyChange={relay(setWasmConcurrency, args.onWasmConcurrencyChange)}
		/>
	);
};

export const Default: Story = {
	args: { firefox: true },
	render: args => <LiveSettingsDialog {...args} />,
	play: async ({ args, canvasElement }) => {
		const frame = within(canvasElement.ownerDocument.body);
		const threat = await frame.findByTestId('show-threat-metrics-picker');
		await userEvent.click(within(threat).getByRole('checkbox'));
		await expect(args.onShowThreatMetricsChange).toHaveBeenCalledWith(true);
		await expect(within(threat).getByRole('checkbox')).toBeChecked();
		await userEvent.click(frame.getByTestId('restore-defaults'));
		await expect(args.onApplyDefaults).toHaveBeenCalledOnce();
	},
};
