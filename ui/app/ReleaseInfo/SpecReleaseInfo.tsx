import i18n from '@i18n/config';
import { translatePlayerClass, translatePlayerSpec } from '@i18n/localization';
import { BUILD_INFO, type Change, specRelease } from '@sim/constants/build_info';
import type { PlayerSpec } from '@sim/player/player_spec';
import { PlayerSpecs } from '@sim/player/specs/index';
import { Button } from '@ui-kit/Button';
import { Icon } from '@ui-kit/Icon';
import { Tooltip, tooltipAnchorProps } from '@ui-kit/Tooltip';
import clsx from 'clsx';
import { type ReactNode, useId } from 'react';

import { CommitLink, joinParts } from './CommitLink';
import { ReleaseInfo } from './ReleaseInfo';

export interface SpecReleaseInfoProps {
	/** The spec the page sims. */
	spec: PlayerSpec<any>;
	className?: string;
}

// The spec's name, or for a spec with none, such as Mage's, the class's, as the page title takes them.
const specLabel = (spec: PlayerSpec<any>): string => translatePlayerSpec(spec) || translatePlayerClass(PlayerSpecs.getPlayerClass(spec));

const changed = (change: Change) =>
	joinParts([i18n.t('common.release_info.changed', { date: change.date }), <CommitLink key="commit" commit={change.commit} />]);

// The page's release, under the sidebar's stats: the game build of the spell data the spec reads and the
// build date. The details open on click, tap or Enter rather than on hover, so a touch screen can open
// them too.
export const SpecReleaseInfo = ({ spec, className }: SpecReleaseInfoProps) => {
	const id = useId();
	const release = specRelease(spec.simLink);
	const rows: { label: string; value: ReactNode[] }[] = [
		{
			label: i18n.t('common.release_info.sim_build'),
			value: joinParts([
				i18n.t('common.release_info.built', { date: BUILD_INFO.date }),
				BUILD_INFO.commit && <CommitLink key="commit" commit={BUILD_INFO.commit} />,
			]),
		},
	];
	if (release.presets) {
		rows.push({ label: i18n.t('common.release_info.presets', { spec: specLabel(spec) }), value: changed(release.presets) });
	}
	if (release.engine) {
		rows.push({ label: i18n.t('common.release_info.engine', { spec: specLabel(spec) }), value: changed(release.engine) });
	}
	return (
		<div className={clsx('flex items-center gap-2', className)}>
			<ReleaseInfo gameVersion={release.gameVersion ?? ''} withCommit={false} />
			<Button variant="unstyled" aria-label={i18n.t('common.release_info.details')} {...tooltipAnchorProps(id)}>
				<Icon name="circle-info" />
			</Button>
			<Tooltip
				id={id}
				place="top"
				align="start"
				openOnClick
				clickable
				content={
					<dl className="m-0 grid grid-cols-[auto_1fr] gap-x-4 gap-y-1">
						{rows.map(row => (
							<div key={row.label} className="contents">
								<dt className="font-bold">{row.label}</dt>
								<dd className="m-0">{row.value}</dd>
							</div>
						))}
					</dl>
				}
			/>
		</div>
	);
};
