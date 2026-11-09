import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import { NoticeIcon, type NoticeLevel } from '@ui-kit/Notice';
import type { ReactNode } from 'react';

import { ENCHANT_NOTICES, ITEM_INFO_NOTICES, ITEM_NOTICES } from '../../item_notices';

export interface ItemNoticeIconViewProps {
	level: NoticeLevel;
	/** The notices, shown in a clickable tooltip on the icon. */
	children: ReactNode;
}

export const ItemNoticeIconView = ({ level, children }: ItemNoticeIconViewProps) => (
	<div className="relative z-1 flex items-center">
		<NoticeIcon
			level={level}
			size="1x"
			label={i18n.t(level === 'info' ? 'common.list_picker.additional_information' : 'common.list_picker.warnings')}
			className="mr-2">
			{children}
		</NoticeIcon>
	</div>
);

export interface ItemNoticeIconProps {
	itemId?: number;
	enchantId?: number;
	additionalNotice?: ReactNode;
}

export const ItemNoticeIcon = ({ itemId, enchantId, additionalNotice }: ItemNoticeIconProps) => {
	const spec = usePlayer().getSpec();

	const itemNotice = itemId ? ITEM_NOTICES.get(itemId) : undefined;
	const ownNotice = itemNotice?.[spec] || itemNotice?.[Spec.SpecUnknown];
	const enchantNotice = enchantId ? ENCHANT_NOTICES.get(enchantId) : undefined;
	const ownEnchantNotice = enchantNotice?.[spec] || enchantNotice?.[Spec.SpecUnknown];
	const infoNotice = itemId ? ITEM_INFO_NOTICES.get(itemId) : undefined;

	if (!ownNotice && !ownEnchantNotice && !additionalNotice && !infoNotice) return null;

	const warns = !!ownNotice || !!ownEnchantNotice || !!additionalNotice;

	return (
		<ItemNoticeIconView level={warns ? 'warning' : 'info'}>
			{ownNotice}
			{ownEnchantNotice}
			{additionalNotice}
			{infoNotice}
		</ItemNoticeIconView>
	);
};
