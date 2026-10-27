import { Popover as BasePopover } from '@base-ui/react/popover';
import i18n from '@i18n/config';
import { Button } from '@ui-kit/Button';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import { type KeyboardEventHandler, type ReactNode, useRef } from 'react';

import { RING_PADDING } from '../../model/placement';
import type { TourCardKind } from '../../model/tour_steps';

/** Between the ring and the card. */
const CARD_GAP = 16;
/** Between the card and the viewport's edge. */
const VIEWPORT_MARGIN = 12;

export interface TourCardProps {
	kind: TourCardKind;
	title: ReactNode;
	body: ReactNode;
	/** Above the title, such as the welcome's logo. */
	media?: ReactNode;
	/** 1-based, of `total`: the counted steps, the welcome and the last card left out. */
	position: number;
	total: number;
	/** The part of the page the card sits beside; null centres it. */
	anchor: Element | null;
	onNext: () => void;
	/** Absent on the first card. */
	onPrevious?: () => void;
	/** Leaves the tour, from the Skip button or Escape. */
	onSkip: () => void;
	/** The card waits for the player to act on the page: the page keeps the focus and the clicks, and the primary button passes the step over. */
	waiting?: boolean;
	onKeyDown?: KeyboardEventHandler<HTMLDivElement>;
}

const viewportAnchor = {
	getBoundingClientRect: () => new DOMRect(0, 0, window.innerWidth, window.innerHeight),
};

// From the bottom of a viewport-sized anchor back up to the middle of the page.
const centred = ({ anchor, positioner }: { anchor: { height: number }; positioner: { height: number } }) => -(anchor.height + positioner.height) / 2;

/**
 * One card of a tour, beside the ringed part of the page or in the middle of it: what that part is, how
 * far along the tour is, and the way back, on or out. While it explains, the focus stays in the card.
 */
export const TourCard = ({ kind, title, body, media, position, total, anchor, onNext, onPrevious, onSkip, waiting = false, onKeyDown }: TourCardProps) => {
	const portalContainer = usePortalContainer();
	const card = useRef<HTMLDivElement>(null);
	const label =
		kind === 'welcome'
			? i18n.t('onboarding.start')
			: kind === 'finish'
				? i18n.t('onboarding.finish')
				: waiting
					? i18n.t('onboarding.skip_step')
					: i18n.t('onboarding.next');
	return (
		<BasePopover.Root
			open
			modal={waiting ? false : 'trap-focus'}
			onOpenChange={(open, details) => {
				// A press outside or the focus leaving is the page being used on a waiting step, not a way out.
				if (!open && (details.reason === 'escape-key' || details.reason === 'close-press')) onSkip();
			}}>
			<BasePopover.Portal className="contents" container={portalContainer ?? undefined}>
				<BasePopover.Positioner
					className="ui-tour-positioner"
					anchor={anchor ?? viewportAnchor}
					side={anchor ? 'right' : 'bottom'}
					align={anchor ? 'start' : 'center'}
					sideOffset={anchor ? CARD_GAP + RING_PADDING : centred}
					collisionPadding={VIEWPORT_MARGIN}
					collisionAvoidance={anchor ? { side: 'flip', align: 'shift', fallbackAxisSide: 'end' } : { side: 'none', align: 'none' }}
					data-testid="tour-card-positioner">
					{/* The card takes the focus, not its primary button: a script focus on a button shows its keyboard ring under the mouse, and the arrow keys move the tour from the card. */}
					<BasePopover.Popup ref={card} className="ui-tour-card" initialFocus={waiting ? false : card} onKeyDown={onKeyDown} data-testid="tour-card">
						{media}
						{kind === 'step' && (
							<div className="flex items-center justify-between text-xs text-muted" data-testid="tour-card-progress">
								<span className="font-bold tracking-wide uppercase">{i18n.t('onboarding.label')}</span>
								<span className="tabular-nums">{i18n.t('onboarding.progress', { position, total })}</span>
							</div>
						)}
						<BasePopover.Title className="mb-0 text-lg leading-snug" data-testid="tour-card-title">
							{title}
						</BasePopover.Title>
						<BasePopover.Description className="mb-0 text-sm leading-relaxed text-muted" data-testid="tour-card-body">
							{body}
						</BasePopover.Description>
						{kind === 'step' && total > 0 && (
							<div className="h-1 overflow-hidden bg-surface" aria-hidden="true">
								<div className="h-full bg-primary" style={{ width: `${(position / total) * 100}%` }} />
							</div>
						)}
						<div className="flex items-center gap-2" data-testid="tour-card-actions">
							{kind !== 'finish' && (
								<BasePopover.Close render={<Button variant="link" size="sm" className="mr-auto" />} data-testid="tour-card-skip">
									{i18n.t('onboarding.skip')}
								</BasePopover.Close>
							)}
							{onPrevious && (
								<Button
									variant="outline-secondary"
									size="sm"
									className={kind === 'finish' ? 'ml-auto' : undefined}
									onClick={onPrevious}
									data-testid="tour-card-previous">
									{i18n.t('onboarding.previous')}
								</Button>
							)}
							<Button variant="primary" size="sm" onClick={kind === 'finish' ? onSkip : onNext} data-testid="tour-card-next">
								{label}
							</Button>
						</div>
					</BasePopover.Popup>
				</BasePopover.Positioner>
			</BasePopover.Portal>
		</BasePopover.Root>
	);
};
