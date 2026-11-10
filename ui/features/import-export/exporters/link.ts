import { IndividualSimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SIM_CATEGORY_KEYS, SimSettingCategories } from '@sim/constants/sim_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { encodeLinkHash, LINK_CATEGORY_PARAM, LINK_DEFAULT_CATEGORIES } from '@sim/state/sim_links';
import { withoutStatFilters } from '@sim/state/stat_filter_sharing';
import { arrayEquals } from '@sim/utils/collections';

import { selectedCategories } from './categories';
import type { ExporterDefinition } from './types';

export const createLink = async (host: IndividualSimHost<any>, exportCategories?: Array<SimSettingCategories>): Promise<string> => {
	if (!exportCategories) {
		exportCategories = LINK_DEFAULT_CATEGORIES;
	}

	const proto = withoutStatFilters(host.toProto(exportCategories));
	const encoded = await encodeLinkHash(IndividualSimSettings.toBinary(proto));

	const linkUrl = new URL(window.location.href);
	linkUrl.hash = encoded;
	if (arrayEquals(exportCategories, LINK_DEFAULT_CATEGORIES)) {
		linkUrl.searchParams.delete(LINK_CATEGORY_PARAM);
	} else {
		const categoryCharString = exportCategories.map(c => SIM_CATEGORY_KEYS.get(c)).join('');
		linkUrl.searchParams.set(LINK_CATEGORY_PARAM, categoryCharString);
	}
	return linkUrl.toString();
};

export const LINK_EXPORTER: ExporterDefinition = {
	title: i18n.t('export.link.title'),
	selectCategories: true,
	holdsLink: true,
	getData: (host, categories) => createLink(host, selectedCategories(categories)),
};
