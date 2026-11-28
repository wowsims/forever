import i18n from '@i18n/config';
import { BUILD_INFO } from '@sim/constants/build_info';
import clsx from 'clsx';

import { CommitLink, joinParts } from './CommitLink';

export interface ReleaseInfoProps {
	/** The game build to name, where the page reads only some parts of the spell data; the newest of all of them by default. */
	gameVersion?: string;
	/** Whether the line names the commit; a spec page leaves it to its details. */
	withCommit?: boolean;
	className?: string;
}

// The game build, the build date and the commit of the running page.
export const ReleaseInfo = ({ gameVersion = BUILD_INFO.gameVersion, withCommit = true, className }: ReleaseInfoProps) => (
	<p className={clsx('m-0', className)}>
		{joinParts([
			gameVersion && i18n.t('common.release_info.game_version', { version: gameVersion }),
			i18n.t('common.release_info.built', { date: BUILD_INFO.date }),
			withCommit && BUILD_INFO.commit && <CommitLink key="commit" commit={BUILD_INFO.commit} />,
		])}
	</p>
);
