import i18n from '@i18n/config';
import { ContentBlock } from '@ui-kit/ContentBlock';
import type { ReactNode } from 'react';

export type BuffSectionKind = 'raid_buffs' | 'party_buffs' | 'debuffs';

export interface BuffSectionProps {
	section: BuffSectionKind;
	empty: boolean;
	children: ReactNode;
}

const BLOCK: Record<BuffSectionKind, string> = {
	raid_buffs: 'buffs-settings',
	party_buffs: 'party-buffs-settings',
	debuffs: 'debuffs-settings',
};

const sectionText = (section: BuffSectionKind): { title: string; tooltip: string; description?: string } => {
	switch (section) {
		case 'raid_buffs':
			return {
				title: i18n.t('settings_tab.raid_buffs.title'),
				tooltip: i18n.t('settings_tab.raid_buffs.tooltip'),
				description: i18n.t('settings_tab.raid_buffs.description'),
			};
		case 'party_buffs':
			return {
				title: i18n.t('settings_tab.party_buffs.title'),
				tooltip: i18n.t('settings_tab.party_buffs.tooltip'),
				description: i18n.t('settings_tab.party_buffs.description'),
			};
		case 'debuffs':
			return { title: i18n.t('settings_tab.debuffs.title'), tooltip: i18n.t('settings_tab.debuffs.tooltip') };
	}
};

export const BuffSection = ({ section, empty, children }: BuffSectionProps) => {
	const { title, tooltip, description } = sectionText(section);
	return (
		<ContentBlock
			rootDataAttributes={{ 'data-block': BLOCK[section] }}
			config={{
				header: { title, tooltip, className: 'flex-col' },
				withoutBody: empty,
				bodyClassName: 'grid grid-cols-1 xl:grid-cols-2 gap-2 fhd:grid-cols-3',
			}}
			headerChildren={description ? <p className="text-sm">{description}</p> : undefined}>
			{children}
		</ContentBlock>
	);
};
