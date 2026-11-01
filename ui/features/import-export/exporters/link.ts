import { IndividualSimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SIM_CATEGORY_KEYS, SimSettingCategories } from '@sim/constants/sim_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { LINK_CATEGORY_PARAM, LINK_DEFAULT_CATEGORIES } from '@sim/state/sim_links';
import { withoutStatFilters } from '@sim/state/stat_filter_sharing';
import { arrayEquals } from '@sim/utils/collections';

import { selectedCategories } from './categories';
import type { ExporterDefinition } from './types';

// Bytes per `String.fromCharCode` call, far below the engines' argument limits.
const CHAR_CODE_CHUNK = 0x8000;

export const createLink = async (host: IndividualSimHost<any>, exportCategories?: Array<SimSettingCategories>): Promise<string> => {
	if (!exportCategories) {
		exportCategories = LINK_DEFAULT_CATEGORIES;
	}

	const proto = withoutStatFilters(host.toProto(exportCategories));

	const protoBytes = IndividualSimSettings.toBinary(proto);
	// `slice()` types the bytes as ArrayBuffer-backed, which a BlobPart needs.
	const deflatedStream = new Blob([protoBytes.slice()]).stream().pipeThrough(new CompressionStream('deflate'));
	const deflated = new Uint8Array(await new Response(deflatedStream).arrayBuffer());
	let binary = '';
	for (let i = 0; i < deflated.length; i += CHAR_CODE_CHUNK) {
		binary += String.fromCharCode(...deflated.subarray(i, i + CHAR_CODE_CHUNK));
	}
	const encoded = btoa(binary);

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
