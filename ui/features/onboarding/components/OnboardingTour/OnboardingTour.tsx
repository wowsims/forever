import { type KeyboardEvent, type ReactNode, useEffect, useEffectEvent, useState } from 'react';

import { useTrackedBox } from '../../hooks/useTrackedBox';
import { dimPath, ringAround } from '../../model/placement';
import { cardKindAt, type SimulationState, stepCount, type TourOutcome, type TourStep } from '../../model/tour_steps';
import { TourCard } from '../TourCard';

/** How long a step looks for its part of the page (a tab's pane, a result) before it is passed over. */
const TARGET_WAIT_MS = 1500;
/** The same, for a step whose part the page may not have at all. */
const OPTIONAL_WAIT_MS = 300;
const TARGET_POLL_MS = 100;
/** The room a card needs beside a tall part of the page to sit next to it. */
const CROWDED_SIDE_ROOM = 400;

export interface TourEnd<S> {
	outcome: TourOutcome;
	/** The card the tour ended on, and where it was: for analytics. */
	step: S;
	index: number;
	total: number;
}

export interface OnboardingTourProps<S extends TourStep<never>> {
	/** The steps this page has, in order (`applicableSteps`). */
	steps: ReadonlyArray<S>;
	/** A step's title and body, and a picture above them (the welcome's logo). */
	text: (step: S) => { title: ReactNode; body: ReactNode; media?: ReactNode };
	/** Opens what a step needs before its target is looked for, such as its tab. */
	prepare?: (step: S) => void;
	/** For a step that waits for a simulation: it moves on once one it saw running is done. */
	simulation?: SimulationState;
	onEnd: (end: TourEnd<S>) => void;
}

// A hidden element still matches its selector, with an empty box.
const findTarget = (selector: string): Element | null => {
	const element = document.querySelector(selector);
	const rect = element?.getBoundingClientRect();
	return rect && rect.width > 0 && rect.height > 0 ? element : null;
};

const missingHere = (candidate: TourStep<never>, from: TourStep<never>): boolean =>
	!!candidate.optional && !!candidate.target && candidate.tab === from.tab && !document.querySelector(candidate.press ?? candidate.target);

const viewport = () => ({ width: window.innerWidth, height: window.innerHeight });

// A control that moves its own value or selection with the arrow keys: a field, a select, a slider, a
// tab list, a menu. An arrow pressed in one of them is the page's, not the tour's.
const ARROW_KEY_CONTROLS =
	'input, textarea, select, [contenteditable=""], [contenteditable="true"], [role="tablist"], [role="listbox"], [role="menu"], [role="menubar"], [role="radiogroup"], [role="slider"], [role="grid"], [role="tree"]';
const takesArrowKeys = (target: EventTarget | null): boolean => target instanceof Element && !!target.closest(ARROW_KEY_CONTROLS);

const scroller = (element: Element): Element | null => {
	for (let parent = element.parentElement; parent; parent = parent.parentElement) {
		const { overflowY } = getComputedStyle(parent);
		if ((overflowY === 'auto' || overflowY === 'scroll') && parent.scrollHeight > parent.clientHeight) return parent;
	}
	return document.scrollingElement;
};

// At once: a smooth scroll is cut short when a tab's pane re-renders. A part taller than most of the page
// goes near the top rather than the middle, where its start would be lost, with room left for a sticky header.
const bringIntoView = (element: Element) => {
	const tall = element.getBoundingClientRect().height > window.innerHeight * 0.7;
	element.scrollIntoView({ block: tall ? 'start' : 'center', inline: 'nearest' });
	if (tall) scroller(element)?.scrollBy(0, -window.innerHeight * 0.15);
};

/**
 * A guided tour over the page: the page dims except for a ring around the part a card explains, and
 * the card sits beside it. The page cannot be clicked while a card explains it, except on a step that
 * waits for the player to run a simulation. A step whose part the page does not show is passed over in
 * the direction of travel. In the card, the right and left arrows move on and back; Escape leaves. On a
 * step that waits, the page keeps the focus, so the left arrow goes back from the page too, unless a
 * control the page uses the arrows in has the focus.
 */
export const OnboardingTour = <S extends TourStep<never>>({ steps, text, prepare, simulation, onEnd }: OnboardingTourProps<S>) => {
	// By id: the steps change under the tour (a result appears once the player has run the sim). When the
	// current one drops out, the tour stays where it was in the list.
	const [current, setCurrent] = useState({ id: steps[0].id, index: 0, direction: 1 as 1 | -1 });
	const found = steps.findIndex(entry => entry.id === current.id);
	const index = found >= 0 ? found : Math.min(current.index, steps.length - 1);
	const step = steps[index];
	// The element a step rings, kept with the step it was found for.
	const [located, setLocated] = useState<{ id: string; element: Element } | null>(null);
	const target = located?.id === step.id ? located.element : null;
	const box = useTrackedBox(target);
	const ring = box && ringAround(box, viewport());
	// A part that fills most of a narrow page leaves no side for the card, which then sits in the middle, over it.
	const crowded = !!box && box.height > window.innerHeight * 0.6 && box.width > window.innerWidth - CROWDED_SIDE_ROOM;
	const kind = cardKindAt(steps, index);
	const waiting = step.waitFor === 'simulation';

	// The steps passed over for want of their part, left out of the count.
	const [passed, setPassed] = useState<ReadonlyArray<string>>([]);
	// The step a simulation was seen running on: a run counts only once the waiting step has seen it start.
	const [sawRun, setSawRun] = useState<string | null>(null);

	const finish = (outcome: TourOutcome) => onEnd({ outcome, step, index, total: steps.length });
	const leave = () => finish(kind === 'finish' ? 'done' : 'skipped');
	const go = (delta: 1 | -1) => {
		// An optional step on the tab already open is passed over at once when the page lacks its part: a run of
		// them (the results tabs a run has nothing for) would otherwise each wait out its short look.
		let to = index + delta;
		const skipped: string[] = [];
		while (steps[to] && missingHere(steps[to], step)) {
			skipped.push(steps[to].id);
			to += delta;
		}
		const next = steps[to];
		if (skipped.length) setPassed(ids => [...ids, ...skipped]);
		if (!next) {
			if (delta === 1) finish('done');
			return;
		}
		setCurrent({ id: next.id, index: to, direction: delta });
		setSawRun(null);
	};

	const open = useEffectEvent(() => prepare?.(step));
	const passOver = useEffectEvent(() => {
		setPassed(ids => [...ids, step.id]);
		go(current.direction);
	});
	const advance = useEffectEvent(() => go(1));

	// Opens what the step needs and finds the part of the page it rings.
	const { id: stepId, target: selector, press, optional } = step;
	useEffect(() => {
		open();
		if (!selector) return;
		const giveUpAt = Math.ceil((optional ? OPTIONAL_WAIT_MS : TARGET_WAIT_MS) / TARGET_POLL_MS);
		let tries = 0;
		let timer = 0;
		let pressed = !press;
		const look = () => {
			if (!pressed && press) {
				const control = document.querySelector<HTMLElement>(press);
				control?.click();
				pressed = !!control;
			}
			const element = pressed ? findTarget(selector) : null;
			if (element) {
				bringIntoView(element);
				setLocated({ id: stepId, element });
			} else if (++tries > giveUpAt) {
				passOver();
			} else {
				timer = window.setTimeout(look, TARGET_POLL_MS);
			}
		};
		// After the tab's pane has rendered.
		timer = window.setTimeout(look, TARGET_POLL_MS / 2);
		return () => clearTimeout(timer);
	}, [stepId, selector, press, optional]);

	useEffect(() => {
		if (!waiting || simulation === undefined) return;
		if (simulation === 'running') setSawRun(stepId);
		else if (simulation === 'done' && sawRun === stepId) advance();
	}, [waiting, simulation, sawRun, stepId]);

	// The card is not modal on a waiting step and does not take the focus, so its own handler hears
	// nothing until the player clicks into it.
	const back = useEffectEvent(() => go(-1));
	const canGoBack = index > 0;
	useEffect(() => {
		if (!waiting || !canGoBack) return;
		const onPageKeyDown = (event: globalThis.KeyboardEvent) => {
			if (event.key !== 'ArrowLeft' || event.defaultPrevented || event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
			if (event.target instanceof Element && event.target.closest('.ui-tour-card')) return;
			if (takesArrowKeys(event.target)) return;
			event.preventDefault();
			back();
		};
		document.addEventListener('keydown', onPageKeyDown);
		return () => document.removeEventListener('keydown', onPageKeyDown);
	}, [waiting, canGoBack]);

	const onKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
		if (event.key === 'ArrowRight' && !waiting) go(1);
		else if (event.key === 'ArrowLeft' && index > 0) go(-1);
		else return;
		event.preventDefault();
	};

	const { title, body, media } = text(step);
	const counted = steps.filter(entry => entry === step || !passed.includes(entry.id));
	const { position, total } = stepCount(counted, counted.indexOf(step));

	return (
		<>
			<div className="ui-tour" data-waiting={waiting ? '' : undefined} data-testid="tour" data-step={step.id}>
				<div className="ui-tour-shield" aria-hidden="true" data-testid="tour-shield" />
				<svg className="ui-tour-dim" aria-hidden="true" width="100%" height="100%">
					<path fillRule="evenodd" d={dimPath(ring, viewport())} />
				</svg>
				{ring && (
					<div
						className="ui-tour-ring"
						aria-hidden="true"
						style={{ transform: `translate(${ring.left}px, ${ring.top}px)`, width: ring.width, height: ring.height }}
						data-testid="tour-ring"
					/>
				)}
			</div>
			{(!step.target || target) && (
				<TourCard
					// A card per step, so each one takes the focus as it opens.
					key={step.id}
					kind={kind}
					title={title}
					body={body}
					media={media}
					position={position}
					total={total}
					anchor={crowded ? null : target}
					waiting={waiting}
					onNext={() => go(1)}
					onPrevious={index > 0 ? () => go(-1) : undefined}
					onSkip={leave}
					onKeyDown={onKeyDown}
				/>
			)}
		</>
	);
};
