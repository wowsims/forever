import { useEnchantPresets } from '@features/enchant-presets/hooks/useEnchantPresets';
import type { ReforgeOptimizerModel, ReforgeOptimizerOptions } from '@features/reforge/model/reforge_optimizer';
import type { ItemSlot } from '@generated/proto/common';
import i18n from '@i18n/config';
import { useSimHost } from '@sim/context/SimHostContext';
import { useSimRun } from '@sim/hooks/useSimRun';
import type { EquippedItem } from '@sim/proto/equipped_item';
import { SimRunKind } from '@sim/state/sim_store';
import { isDevMode } from '@sim/utils/env';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Popover } from '@ui-kit/Popover';
import { ProgressTrackerDialog } from '@ui-kit/ProgressTrackerDialog';
import { toastManager, type ToastOptions } from '@ui-kit/Toast';
import { Tooltip, tooltipAnchorProps, type TooltipRefProps } from '@ui-kit/Tooltip';
import { type ReactNode, useCallback, useId, useRef, useState } from 'react';

import { trackEvent, trackPageView } from '../../../../tracking/utils';
import { useReforgeField } from '../../hooks/useReforgeField';
import { ReforgeDoneToast } from './ReforgeDoneToast';
import { ReforgeSettingsPanel } from './ReforgeSettingsPanel';
import { ReforgeSoftCapsTooltip } from './ReforgeSoftCapsTooltip';

export interface ReforgePanelProps {
	model: ReforgeOptimizerModel;
	options?: ReforgeOptimizerOptions;
}

/**
 * The optimizer's sidebar controls: the run button with its breakpoint tooltip, the settings
 * popover, the progress dialog and the run's outcome toasts. Everything below it is the model,
 * which `individual_sim_ui` owns and every caller still reads as `simUI.reforger`.
 */
export const ReforgePanel = ({ model, options }: ReforgePanelProps) => {
	const host = useSimHost();
	const player = host.player;
	const sim = host.sim;

	const [open, setOpen] = useState(false);
	const [progressOpen, setProgressOpen] = useState(false);
	const toastId = useRef<string | null>(null);

	const isCancelling = useRef(false);

	const { isRunning } = useSimRun(SimRunKind.ReforgeOptimize);
	const presetsOn = useReforgeField(model.settings, 'useEnchantPresets', () => model.settings.useEnchantPresets);
	const presets = useEnchantPresets();
	const needsPreset = presetsOn && !presets.some(preset => preset.checked);
	const softCapsTooltipId = useId();
	const settingsTooltipId = useId();
	const settingsTooltipRef = useRef<TooltipRefProps>(null);

	const hideToast = useCallback(() => {
		// Without the id this would close every toast in the standard area, not just this one.
		if (toastId.current) toastManager.close(toastId.current);
		toastId.current = null;
	}, []);

	const showToast = useCallback((node: ReactNode, toastOptions: Omit<ToastOptions, 'body'>) => {
		toastId.current = toastManager.add({ ...toastOptions, body: node });
	}, []);

	const onReforgeDone = useCallback(() => {
		const currentGear = player.getGear();
		const itemSlots = currentGear.getItemSlots();
		const changedSlots = new Map<ItemSlot, EquippedItem | undefined>();
		for (const slot of itemSlots) {
			const previous = model.previousGear?.getEquippedItem(slot);
			const current = currentGear?.getEquippedItem(slot);
			if ((!previous && current) || (previous && current && !previous?.equals(current))) changedSlots.set(slot, current);
		}

		trackEvent({ action: 'settings', category: 'reforging', label: 'suggest_success' });

		if (!changedSlots.size) {
			toastId.current = toastManager.add({
				variant: 'success',
				body: i18n.t('gear_tab.reforge_success.no_changes'),
				autohide: true,
				delay: 3000,
			});
			return;
		}

		showToast(<ReforgeDoneToast itemSlots={itemSlots} changedSlots={changedSlots} previousGear={model.previousGear} />, {
			variant: 'success',
			autohide: false,
			delay: 3000,
		});
	}, [model, player, showToast]);

	const onReforgeError = useCallback(
		(error: unknown) => {
			if (isDevMode()) console.log(error);
			// The gear is left as the optimizer set it, exactly as the vanilla panel does; only the
			// derived stats are recomputed from the gear the run started with.
			if (model.previousGear) void model.updateGear(model.previousGear);
			trackEvent({ action: 'settings', category: 'reforging', label: 'suggest_error', value: String(error) });

			showToast(
				<>
					{i18n.t('sidebar.buttons.suggest_reforges.reforge_optimization_failed')}
					<p />
					<p>
						<b>{i18n.t('sidebar.buttons.suggest_reforges.failure_reason')}</b> <i>{String(error)}</i>
					</p>
				</>,
				{ variant: 'error', delay: 10000 },
			);
		},
		[model, showToast],
	);

	const onReforgeFinally = useCallback(() => {
		setProgressOpen(false);
		performance.mark('reforge-optimization-end');
		const completionTimeInMs = performance.measure('reforge-optimization-measure', 'reforge-optimization-start', 'reforge-optimization-end').duration;
		if (isDevMode()) console.log('Gear optimization took:', `${completionTimeInMs.toFixed(2)}ms`);
		trackEvent({ action: 'settings', category: 'reforging', label: 'suggest_duration', value: Math.ceil(completionTimeInMs / 1000) });
	}, []);

	const onOptimize = async () => {
		if (sim.runs.isRunning(SimRunKind.ReforgeOptimize)) return;
		hideToast();
		setProgressOpen(true);
		trackEvent({ action: 'settings', category: 'reforging', label: 'suggest_start' });

		// Reset per run, so an error after the first cancel is not swallowed and the gear is restored.
		isCancelling.current = false;
		try {
			performance.mark('reforge-optimization-start');
			const gear = await model.optimizeReforges();
			await player.setGearAsync(gear);
			onReforgeDone();
		} catch (error) {
			if (isCancelling.current) return;
			onReforgeError(error);
		} finally {
			onReforgeFinally();
		}
	};

	const onCancel = async () => {
		isCancelling.current = true;
		if (isDevMode()) console.log('User cancelled gear optimization');
		try {
			await model.abortReforgeOptimization();
		} catch {
			// The abort races the solve finishing on its own; either way the run is over.
		}
		if (model.previousGear) player.setGear(model.previousGear);
		setProgressOpen(false);
		trackEvent({ action: 'settings', category: 'reforging', label: 'suggest_cancel' });

		toastManager.add({ variant: 'warning', body: i18n.t('sidebar.buttons.suggest_reforges.reforge_optimization_cancelled'), delay: 3000 });
	};

	const settingsTriggerProps = { ...tooltipAnchorProps(settingsTooltipId), 'data-sidebar-action': '' };

	return (
		<>
			<Button
				variant="outline-primary"
				className="peer col-span-full row-start-1 pr-settings-button-inset"
				aria-label={i18n.t('sidebar.buttons.suggest_reforges.title')}
				data-testid="suggest-reforges-action-button"
				data-sidebar-action=""
				disabled={isRunning || needsPreset}
				focusableWhenDisabled={needsPreset}
				loading={isRunning}
				onClick={onOptimize}
				{...tooltipAnchorProps(softCapsTooltipId)}>
				{/* One line-height tall (h-lh), so the button pair is as tall as a text action such as Simulate. */}
				<span className="flex h-lh items-center">
					<Icon name="screwdriver-wrench" />
				</span>
			</Button>
			<Popover
				open={open}
				onOpenChange={nextOpen => {
					setOpen(nextOpen);
					if (!nextOpen) return;
					// Closed through the ref rather than unmounted: `hidden` leaves react-tooltip open on the remount, with no anchor hover left to close it.
					settingsTooltipRef.current?.close();
					trackPageView('Reforge Settings', 'reforge-settings');
				}}
				side="right"
				align="start"
				className="min-w-popup-min-w"
				testId="reforge-optimiser-popover"
				backdrop
				maxWidthClassName="max-w-86 max-lg:max-w-settings-panel-max"
				triggerRender={
					<Button
						variant="outline-primary"
						className="col-start-2 row-start-1 w-(--settings-button-width) border-transparent bg-transparent peer-hover:peer-not-data-disabled:text-primary-foreground hover:border-primary-hover hover:bg-primary-hover focus-visible:border-primary-hover focus-visible:bg-primary-hover active:border-primary-active active:bg-primary-active active:text-primary-foreground data-popup-open:border-primary-active data-popup-open:bg-primary-active data-popup-open:text-primary-foreground"
					/>
				}
				triggerTestId="suggest-reforges-button-settings"
				triggerProps={settingsTriggerProps}
				trigger={<Icon name="gear" />}>
				<ReforgeSettingsPanel model={model} options={options} onClose={() => setOpen(false)} />
			</Popover>
			<Tooltip ref={settingsTooltipRef} id={settingsTooltipId} place="bottom" content={i18n.t('sidebar.buttons.suggest_reforges.tooltip')} />
			<Tooltip
				id={softCapsTooltipId}
				place="bottom"
				clickable
				className="suggest-reforges-softcaps"
				maxWidth="max-w-76"
				// The limits are read per open.
				render={() => {
					const title = i18n.t('sidebar.buttons.suggest_reforges.title');
					const softCaps = model.softCapsConfigWithLimits;
					const body = needsPreset ? (
						<p className="mb-0">{i18n.t('sidebar.buttons.suggest_reforges.no_enchant_preset')}</p>
					) : softCaps?.length ? (
						<ReforgeSoftCapsTooltip
							playerClass={player.getClass()}
							softCaps={softCaps}
							additionalInformation={options?.additionalSoftCapTooltipInformation ?? {}}
						/>
					) : null;
					if (!body) return title;
					return (
						<>
							<p className="mb-1 font-semibold">{title}</p>
							{body}
						</>
					);
				}}
			/>
			{progressOpen && (
				<ProgressTrackerDialog
					open
					testId="reforge-optimizer-progress-tracker"
					title={i18n.t('sidebar.buttons.suggest_reforges.progress.title')}
					state={{ stage: 'initializing' }}
					warning={
						<>
							<p>{i18n.t('sidebar.buttons.suggest_reforges.progress.lengthy')}</p>
							<p className="mb-0">{i18n.t('sidebar.buttons.suggest_reforges.progress.cancel_hint')}</p>
						</>
					}
					onCancel={onCancel}
				/>
			)}
		</>
	);
};
