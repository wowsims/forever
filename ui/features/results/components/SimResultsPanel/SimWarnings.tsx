import type { WarningsRegistry } from '@features/results/model/warnings';
import i18n from '@i18n/config';
import { useReadyStoreSubscribe } from '@sim/hooks/useReadyStoreSubscribe';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { NOTICE_COLOR, NOTICE_ICON_NAME } from '@ui-kit/Notice';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx, { ClassValue } from 'clsx';
import { useId } from 'react';

export interface SimWarningsProps {
	warnings: WarningsRegistry;
	/** Every warning judges the player's gear and talents, which are empty until the sim has loaded. */
	ready: boolean;
	className?: ClassValue;
}

/**
 * `getContents()` builds a fresh array per call, so it is read through the snapshot cache;
 * `useSyncExternalStore` would read the new identity as a change and loop.
 */
export const SimWarnings = ({ warnings, ready, className }: SimWarningsProps) => {
	const id = useId();
	const contents = useReadyStoreSubscribe(warnings.subscribe, warnings.getContents, ready) ?? [];
	return (
		<div className={clsx('text-center empty:hidden', className)} data-testid="warning-zone">
			{contents.length > 0 && (
				<>
					<Button
						variant="unstyled"
						className="focus-visible:focus-ring"
						data-testid="warning-trigger"
						aria-label={i18n.t('sidebar.warnings.label')}
						{...tooltipAnchorProps(id)}>
						<Icon name={NOTICE_ICON_NAME.warning} size="3x" className={NOTICE_COLOR.warning} />
					</Button>
					<Tooltip
						id={id}
						widthClassName="w-full"
						place="bottom"
						content={
							<ul className="mb-0 ps-2 text-start">
								{contents.map((warning, index) => (
									<li key={`${index}:${warning}`}>{warning}</li>
								))}
							</ul>
						}
					/>
				</>
			)}
		</div>
	);
};
