import { SelectorModalTabs } from '../types';

export type TabEligibility = {
	hasEnchants: boolean;
};

export const resolveSelectedTab = (selectedTab: SelectorModalTabs, { hasEnchants }: TabEligibility) => {
	if (selectedTab === SelectorModalTabs.Enchants && !hasEnchants) {
		return SelectorModalTabs.Items;
	}
	return selectedTab;
};
