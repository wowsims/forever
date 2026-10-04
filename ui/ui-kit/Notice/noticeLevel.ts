import type { IconName } from '@ui-kit/Icon';

export type NoticeLevel = 'info' | 'warning' | 'error';

export const NOTICE_ICON_NAME = {
	info: 'circle-info',
	warning: 'triangle-exclamation',
	error: 'triangle-exclamation',
} as const satisfies Record<NoticeLevel, IconName>;

export const NOTICE_COLOR: Record<NoticeLevel, string> = {
	info: 'text-off-white drop-shadow-glow-link',
	warning: 'text-link-warning drop-shadow-glow-danger',
	error: 'text-danger drop-shadow-glow-danger',
};
