import { Icon } from '@ui-kit/Icon';
import clsx from 'clsx';
import type { ReactNode } from 'react';

import { NOTICE_COLOR, NOTICE_ICON_NAME, type NoticeLevel } from './noticeLevel';

export interface NoticeProps {
	level: NoticeLevel;
	className?: string;
	testId?: string;
	children: ReactNode;
}

export const Notice = ({ level, className, testId, children }: NoticeProps) => (
	<div className={clsx('ui-notice', className)} data-level={level} data-testid={testId}>
		<Icon name={NOTICE_ICON_NAME[level]} className={NOTICE_COLOR[level]} />
		<div>{children}</div>
	</div>
);
