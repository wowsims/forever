import { Player } from '@generated/proto/api';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { IndividualSimSettings, type SimlPrintRequest } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SIM_CATEGORY_KEYS, type SimSettingCategories } from '@sim/constants/sim_settings';
import type { IndividualSimHost } from '@sim/sim_host';
import { withoutStatFilters } from '@sim/state/stat_filter_sharing';

import { located, printSiml } from '../siml_client';
import { selectedCategories } from './categories';
import type { ExporterDefinition } from './types';

/** The file a rotation export downloads as, whose double extension says it holds a rotation. */
export const SIML_ROTATION_FILE = 'rotation.apl.siml';

/**
 * The file a setup export downloads as, which is also what pasted text with no file name reads as: a full setup, or
 * any part of one (sim/siml/SPEC.md §1).
 */
export const SIML_SETUP_FILE = 'wowsims.siml';

/** The print `request` asks for. Text that would not import is refused with its problems, rather than handed out. */
const printReadable = async (host: IndividualSimHost<any>, request: Partial<SimlPrintRequest>): Promise<string> => {
	const { text, diagnostics } = await printSiml(host.sim, { spec: host.player.getSpec(), ...request });
	if (diagnostics.length) {
		throw new Error(i18n.t('export.siml.error_unreadable', { problems: diagnostics.map(located).join('; ') }));
	}
	return text;
};

/**
 * The setup as a full SIML file, holding the sections of `categories`. An import loads it over the page
 * (sim/siml/SPEC.md §8): an empty consumables: or encounter: section empties that message, but a setting at its zero
 * value, such as no talents or 0 target dummies, keeps the importer's. No category, as in the JSON export, is every
 * category. The rotation keeps the type the player set, so an Auto or Simple one reads back as Auto or Simple.
 */
const printSimlSettings = (host: IndividualSimHost<any>, categories: Array<SimSettingCategories>): Promise<string> =>
	printReadable(host, {
		settings: withoutStatFilters(host.toProto(categories)),
		categories: categories.map(category => SIM_CATEGORY_KEYS.get(category)!),
	});

/**
 * The rotation the sim runs, as an x.apl.siml to edit and import back. An Auto or Simple rotation is written as the
 * APL list the sim generates for it: kept as Auto or Simple, the import would have the sim generate the list again
 * and drop the edits.
 */
const printSimlRotation = (host: IndividualSimHost<any>): Promise<string> =>
	printReadable(host, {
		section: 'apl',
		settings: IndividualSimSettings.create({
			player: Player.create({ rotation: { ...host.player.getResolvedAplRotation(), type: APLRotationType.TypeAPL, simple: undefined } }),
		}),
	});

export const SIML_EXPORTER: ExporterDefinition = {
	title: i18n.t('export.siml.title'),
	allowDownload: true,
	downloadFileName: SIML_SETUP_FILE,
	downloadMimeType: 'text/plain',
	selectCategories: true,
	getData: (host, categories) => printSimlSettings(host, selectedCategories(categories)),
};

export const SIML_ROTATION_EXPORTER: ExporterDefinition = {
	title: i18n.t('export.siml_rotation.title'),
	allowDownload: true,
	downloadFileName: SIML_ROTATION_FILE,
	downloadMimeType: 'text/plain',
	getData: printSimlRotation,
};
