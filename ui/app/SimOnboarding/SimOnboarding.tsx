import { useActivateTab } from '@features/hooks/useActivateTab';
import { applicableSteps, OnboardingOutlet, OnboardingTour, type SimulationState, type TourEnd, type TourRun, type TourStep } from '@features/onboarding';
import { ResultsPanelStage } from '@features/results/components/SimResultsPanel/results_panel_store';
import i18n from '@i18n/config';
import { ADDON_URL, DISCORD_URL } from '@sim/constants/other';
import { Button } from '@ui-kit/Button';
import { useCallback, useEffect, useMemo, useState, useSyncExternalStore } from 'react';
import { Trans } from 'react-i18next';

import { trackEvent } from '../../tracking/utils';
import type { SimHostObject } from '../individual_sim_ui';
import { SIM_TOUR_STEPS, type SimTourContext } from './sim_tour';

export interface SimOnboardingProps {
	/** Null until the shell has built it; the tour waits for it. */
	host: SimHostObject<any> | null;
	/** The sim has loaded; the tour waits for it. */
	simReady: boolean;
}

/**
 * The spec page's guided tour, drawn inside `SimTabs`, which it opens tabs through. It opens by itself the
 * first time a player opens any spec page, once the page is ready, and again from the options dialog
 * (`useOnboardingRestart`); the `OnboardingProvider` around the page keeps whether it has ended.
 */
export const SimOnboarding = ({ host, simReady }: SimOnboardingProps) => {
	const renderTour = useCallback((run: TourRun) => host && <SimTour host={host} run={run} />, [host]);
	return <OnboardingOutlet ready={simReady && !!host} renderTour={renderTour} />;
};

const LOGO = '/forever/assets/img/WoW-Simulator-Icon.png';

// The links a card's words can carry, as `<addon>…</addon>` and `<discord>…</discord>`.
const LINKS = {
	addon: <Button href={ADDON_URL} target="_blank" variant="link" size="none" />,
	discord: <Button href={DISCORD_URL} target="_blank" variant="link" size="none" />,
};

const text = (step: TourStep<SimTourContext>) => ({
	title: i18n.t(`onboarding.steps.${step.id}.title`),
	body: <Trans i18n={i18n} i18nKey={`onboarding.steps.${step.id}.body`} components={LINKS} />,
	media: step.id === 'welcome' ? <img className="size-16 self-center" src={LOGO} alt="" data-testid="tour-logo" /> : undefined,
});

const simulationState = (stage: ResultsPanelStage): SimulationState =>
	stage === ResultsPanelStage.Pending || stage === ResultsPanelStage.Running ? 'running' : stage === ResultsPanelStage.Result ? 'done' : 'idle';

// The tab strip's open tab, whose test id is the tab's id.
const openTab = () => document.querySelector('[data-testid="sim-tabs"] [aria-selected="true"]')?.getAttribute('data-testid') ?? null;

const SimTour = ({ host, run }: { host: SimHostObject<any>; run: TourRun }) => {
	const activateTab = useActivateTab();
	const stage = useSyncExternalStore(host.resultsPanel.subscribe, host.resultsPanel.getStage);
	const context = useMemo((): SimTourContext => ({ canSimulate: !host.simDisabled, hasResult: stage === ResultsPanelStage.Result }), [host, stage]);
	const steps = useMemo(() => applicableSteps(SIM_TOUR_STEPS, context), [context]);
	// The tour opens tabs as it goes; the player gets back the one they were on.
	const [startTab] = useState(openTab);

	useEffect(() => {
		if (run.trigger === 'restart') trackEvent({ action: 'click', category: 'onboarding', label: 'start' });
	}, [run.trigger]);

	const prepare = useCallback(
		(step: TourStep<SimTourContext>) => {
			if (step.tab) activateTab(step.tab);
		},
		[activateTab],
	);

	const onEnd = useCallback(
		({ outcome, step, index, total }: TourEnd<TourStep<SimTourContext>>) => {
			trackEvent({
				action: 'click',
				category: 'onboarding',
				label: outcome,
				value: step.id,
				additionalData: { trigger: run.trigger, step: index + 1, steps: total },
			});
			if (startTab && startTab !== openTab()) activateTab(startTab);
			run.end(outcome);
		},
		[activateTab, run, startTab],
	);

	return <OnboardingTour steps={steps} text={text} prepare={prepare} simulation={simulationState(stage)} onEnd={onEnd} />;
};
