import { useOnboardingRestart } from '@features/onboarding';
import i18n from '@i18n/config';
import { setLang, supportedLanguages } from '@i18n/locale_service';
import { useSim } from '@sim/context/SimHostContext';
import { useSimStore } from '@sim/hooks/useSimStore';
import { useUiStore } from '@sim/hooks/useUiStore';
import { BooleanPicker } from '@ui-kit/BooleanPicker';
import { Button } from '@ui-kit/Button';
import { ChoiceField, type ChoiceOption } from '@ui-kit/ChoiceField';
import { Dialog } from '@ui-kit/Dialog';
import { FieldGrid } from '@ui-kit/FieldGrid';
import { HelpText } from '@ui-kit/FormControl';
import { NumberField } from '@ui-kit/TextField';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import { useEffect, useId, useMemo, useState } from 'react';

import { trackEvent } from '../../tracking/utils';
import type { SimHostObject } from '../individual_sim_ui';
import { restoreDefaults } from '../restore_defaults';

export interface SettingsDialogViewProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	fixedRngSeed: number;
	onFixedRngSeedChange: (seed: number) => void;
	lastUsedRngSeed: number;
	/** A language code; one not in `supportedLanguages` shows as English. */
	language: string;
	onLanguageChange: (language: string) => void;
	showThreatMetrics: boolean;
	onShowThreatMetricsChange: (show: boolean) => void;
	showExperimental: boolean;
	onShowExperimentalChange: (show: boolean) => void;
	showPowerTools: boolean;
	onShowPowerToolsChange: (show: boolean) => void;
	showQuickSwap: boolean;
	onShowQuickSwapChange: (show: boolean) => void;
	wasmConcurrency: number;
	onWasmConcurrencyChange: (workers: number) => void;
	/** False hides the worker picker; the row stays mounted. */
	showConcurrency: boolean;
	maxWorkers: number;
	/** Shows the memory note under the worker picker. */
	firefox: boolean;
	onApplyDefaults: () => void;
	/** Closes the dialog and opens the guided tour from its first card; absent where the page has no tour. */
	onRestartTour?: () => void;
}

const LANGUAGE_OPTIONS: Array<ChoiceOption<string>> = Object.entries(supportedLanguages).map(([language, name]) => ({ value: language, label: name }));

const LastUsedRngSeed = ({ seed }: { seed: number }) => (
	<HelpText as="div">
		<span>{i18n.t('info.options.fixed_rng_seed.last_used')}</span>&nbsp;{seed}
	</HelpText>
);

const WorkerNote = ({ shown }: { shown: boolean }) => (
	<HelpText as="div" hidden={!shown}>
		{shown ? i18n.t('info.options.use_multiple_cpu_cores.firefox_note') : ''}
	</HelpText>
);

export const SettingsDialogView = ({
	open,
	onOpenChange,
	fixedRngSeed,
	onFixedRngSeedChange,
	lastUsedRngSeed,
	language,
	onLanguageChange,
	showThreatMetrics,
	onShowThreatMetricsChange,
	showExperimental,
	onShowExperimentalChange,
	showPowerTools,
	onShowPowerToolsChange,
	showQuickSwap,
	onShowQuickSwapChange,
	wasmConcurrency,
	onWasmConcurrencyChange,
	showConcurrency,
	maxWorkers,
	firefox,
	onApplyDefaults,
	onRestartTour,
}: SettingsDialogViewProps) => {
	const restoreTooltipId = useId();

	const workerOptions = useMemo(() => {
		const options: Array<ChoiceOption<number>> = [{ value: 0, label: i18n.t('info.options.use_multiple_cpu_cores.off') }];
		for (let workers = 2; workers <= maxWorkers; workers++) {
			options.push({ value: workers, label: workers.toString() });
		}
		return options;
	}, [maxWorkers]);

	return (
		<Dialog
			open={open}
			onOpenChange={onOpenChange}
			title={i18n.t('info.options.title')}
			footer={
				<>
					{onRestartTour && (
						<Button variant="outline-primary" className="mr-2" data-testid="restart-tour" onClick={onRestartTour}>
							{i18n.t('info.options.restart_tour')}
						</Button>
					)}
					<Button variant="primary" data-testid="restore-defaults" {...tooltipAnchorProps(restoreTooltipId)} onClick={onApplyDefaults}>
						{i18n.t('info.options.restore_defaults.button')}
					</Button>
					<Tooltip id={restoreTooltipId} content={i18n.t('info.options.restore_defaults.tooltip')} />
				</>
			}>
			<div className="flex flex-col gap-2">
				<FieldGrid>
					<div>
						<NumberField
							id="simui-fixed-rng-seed"
							label={i18n.t('info.options.fixed_rng_seed.label')}
							labelTooltip={i18n.t('info.options.fixed_rng_seed.tooltip')}
							value={fixedRngSeed}
							onChange={onFixedRngSeedChange}
						/>
						<LastUsedRngSeed seed={lastUsedRngSeed} />
					</div>
					<ChoiceField
						id="simui-language-picker"
						label={i18n.t('info.options.language.label')}
						labelTooltip={i18n.t('info.options.language.tooltip')}
						options={LANGUAGE_OPTIONS}
						value={LANGUAGE_OPTIONS.some(option => option.value === language) ? language : 'en'}
						onChange={onLanguageChange}
					/>
				</FieldGrid>
				<div className="w-1/2 pr-2" data-testid="show-threat-metrics-picker">
					<BooleanPicker
						id="simui-show-threat-metrics"
						label={i18n.t('info.options.feature_toggles.show_threat_metrics')}
						labelTooltip={i18n.t('info.options.feature_toggles.show_threat_metrics_tooltip')}
						layout="inline"
						value={showThreatMetrics}
						onChange={onShowThreatMetricsChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="show-experimental-picker">
					<BooleanPicker
						id="simui-show-experimental"
						label={i18n.t('info.options.feature_toggles.show_experimental')}
						labelTooltip={i18n.t('info.options.feature_toggles.show_experimental_tooltip')}
						layout="inline"
						value={showExperimental}
						onChange={onShowExperimentalChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="show-power-tools-picker">
					<BooleanPicker
						id="simui-show-power-tools"
						label={i18n.t('info.options.feature_toggles.show_power_tools')}
						labelTooltip={i18n.t('info.options.feature_toggles.show_power_tools_tooltip')}
						layout="inline"
						value={showPowerTools}
						onChange={onShowPowerToolsChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="show-quick-swap-picker">
					<BooleanPicker
						id="simui-show-quick-swap"
						label={i18n.t('info.options.feature_toggles.show_quick_swap')}
						labelTooltip={i18n.t('info.options.feature_toggles.show_quick_swap_tooltip')}
						layout="inline"
						value={showQuickSwap}
						onChange={onShowQuickSwapChange}
					/>
				</div>
				<div className="w-1/2 pr-2" data-testid="use-concurrency-container" hidden={!showConcurrency}>
					<ChoiceField
						id="simui-concurrent-workers-picker"
						label={i18n.t('info.options.use_multiple_cpu_cores.label')}
						labelTooltip={i18n.t('info.options.use_multiple_cpu_cores.tooltip')}
						options={workerOptions}
						value={wasmConcurrency}
						onChange={onWasmConcurrencyChange}
					/>
					<WorkerNote shown={firefox} />
				</div>
			</div>
		</Dialog>
	);
};

export interface SettingsDialogProps {
	open: boolean;
	onOpenChange: (open: boolean) => void;
	host: SimHostObject<any>;
}

export const SettingsDialog = ({ open, onOpenChange, host }: SettingsDialogProps) => {
	const sim = useSim();
	const fixedRngSeed = useSimStore('fixedRngSeed');
	const lastUsedRngSeed = useSimStore('lastUsedRngSeed');
	const language = useUiStore('language');
	const showThreatMetrics = useUiStore('showThreatMetrics');
	const showExperimental = useUiStore('showExperimental');
	const showPowerTools = useUiStore('showPowerTools');
	const showQuickSwap = useUiStore('showQuickSwap');
	const wasmConcurrency = useUiStore('wasmConcurrency');
	const restartTour = useOnboardingRestart();

	// Local sim has native threading, so the row is only meaningful on wasm. It starts shown because
	// the answer is a promise.
	const [isWasm, setIsWasm] = useState(true);
	useEffect(() => {
		let live = true;
		sim.isWasm().then(wasm => {
			if (live) setIsWasm(wasm);
		});
		return () => {
			live = false;
		};
	}, [sim]);

	const setLanguage = (next: string) => {
		trackEvent({ action: 'settings', category: 'language', label: 'update', value: next });
		sim.setLanguage(next);
		setLang(next);
		// Every string on the page was read at render, so the new language only lands on a reload.
		setTimeout(() => location.reload(), 300);
	};

	const setShowExperimental = (show: boolean) => {
		trackEvent({ action: 'settings', category: 'show-experimental', label: 'update', value: show });
		sim.setShowExperimental(show);
	};

	const setShowPowerTools = (show: boolean) => {
		trackEvent({ action: 'settings', category: 'show-power-tools', label: 'update', value: show });
		sim.setShowPowerTools(show);
	};

	const setWasmConcurrency = (workers: number) => {
		trackEvent({ action: 'settings', category: 'concurrency', label: 'update', value: workers });
		sim.setWasmConcurrency(workers);
	};

	return (
		<SettingsDialogView
			open={open}
			onOpenChange={onOpenChange}
			fixedRngSeed={fixedRngSeed}
			onFixedRngSeedChange={seed => sim.setFixedRngSeed(seed)}
			lastUsedRngSeed={lastUsedRngSeed}
			language={language}
			onLanguageChange={setLanguage}
			showThreatMetrics={showThreatMetrics}
			onShowThreatMetricsChange={show => sim.setShowThreatMetrics(show)}
			showExperimental={showExperimental}
			onShowExperimentalChange={setShowExperimental}
			showPowerTools={showPowerTools}
			onShowPowerToolsChange={setShowPowerTools}
			showQuickSwap={showQuickSwap}
			onShowQuickSwapChange={show => sim.setShowQuickSwap(show)}
			wasmConcurrency={wasmConcurrency}
			onWasmConcurrencyChange={setWasmConcurrency}
			showConcurrency={isWasm}
			maxWorkers={navigator.hardwareConcurrency}
			firefox={navigator.userAgent.toLowerCase().includes('firefox')}
			onApplyDefaults={() => restoreDefaults(host)}
			onRestartTour={
				restartTour &&
				(() => {
					onOpenChange(false);
					restartTour();
				})
			}
		/>
	);
};
