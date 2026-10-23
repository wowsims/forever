import type { Env } from '@sim/state/env';
import { createContext, type ReactNode, use, useCallback, useEffect, useMemo, useState } from 'react';

import { readTourOutcome, type TourOutcome, writeTourOutcome } from '../../model/tour_steps';

/** How long the page settles, once ready, before the tour opens by itself. */
const FIRST_OPEN_DELAY_MS = 900;

/** How the tour opened: by itself on a first visit, or because the player asked for it. */
export type TourTrigger = 'auto' | 'restart';

export interface TourRun {
	trigger: TourTrigger;
	/** Ends the tour and records how, so it does not open by itself again. */
	end: (outcome: TourOutcome) => void;
}

interface OnboardingState {
	/** The run's number: 0 for the one that opened by itself, each restart one more; null while closed. */
	run: number | null;
	openFirst: () => void;
	restart: () => void;
	end: (outcome: TourOutcome) => void;
	seen: () => boolean;
}

const OnboardingContext = createContext<OnboardingState | null>(null);

/** Opens the tour again from its first card; undefined outside an `OnboardingProvider`. */
export const useOnboardingRestart = (): (() => void) | undefined => use(OnboardingContext)?.restart;

export interface OnboardingProviderProps {
	storage: Env['storage'];
	/** Where the outcome is kept: once a tour has ended it does not open by itself again. */
	storageKey: string;
	children: ReactNode;
}

/**
 * Whether the tour is running, for an `OnboardingOutlet` below to draw and for `useOnboardingRestart` to open
 * from anywhere under it, such as a dialog outside the part of the page the tour draws in.
 */
export const OnboardingProvider = ({ storage, storageKey, children }: OnboardingProviderProps) => {
	const [run, setRun] = useState<number | null>(null);

	const openFirst = useCallback(() => setRun(current => current ?? 0), []);
	const restart = useCallback(() => setRun(current => (current ?? 0) + 1), []);
	const end = useCallback(
		(outcome: TourOutcome) => {
			writeTourOutcome(storage, storageKey, outcome);
			setRun(null);
		},
		[storage, storageKey],
	);
	const seen = useCallback(() => readTourOutcome(storage, storageKey) !== null, [storage, storageKey]);
	const value = useMemo(() => ({ run, openFirst, restart, end, seen }), [run, openFirst, restart, end, seen]);

	return <OnboardingContext value={value}>{children}</OnboardingContext>;
};

export interface OnboardingOutletProps {
	/** The page is ready to be toured; the tour waits for it. */
	ready: boolean;
	/** Draws the running tour. A new run (a restart) remounts it at its first card. */
	renderTour: (run: TourRun) => ReactNode;
}

/** Where the running tour is drawn; nothing outside an `OnboardingProvider`. Opens it by itself once the page is first ready, unless a tour has ended before. */
export const OnboardingOutlet = ({ ready, renderTour }: OnboardingOutletProps) => {
	const state = use(OnboardingContext);
	const run = state?.run ?? null;

	useEffect(() => {
		if (!state || !ready || state.seen()) return;
		const { openFirst } = state;
		const timer = setTimeout(openFirst, FIRST_OPEN_DELAY_MS);
		return () => clearTimeout(timer);
	}, [state, ready]);

	return state && run !== null && ready ? <TourMount key={run} trigger={run === 0 ? 'auto' : 'restart'} end={state.end} renderTour={renderTour} /> : null;
};

const TourMount = ({ trigger, end, renderTour }: TourRun & { renderTour: (run: TourRun) => ReactNode }) => {
	// The focus goes back where it was when the tour closes, if that is still on the page.
	useEffect(() => {
		const previous = document.activeElement;
		return () => {
			if (previous instanceof HTMLElement && previous.isConnected) previous.focus({ preventScroll: true });
		};
	}, []);
	return renderTour({ trigger, end });
};
