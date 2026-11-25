import { IndividualSimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { Database } from '@sim/proto/database';
import { fromJsonStringLenient } from '@sim/proto/lenient_json';
import { withoutMyEncounter } from '@sim/state/my_encounter_sharing';
import { individualSimSettingsCategories } from '@sim/state/serialization';
import { withViewerStatFilters } from '@sim/state/stat_filter_sharing';

import type { ImporterDefinition } from './types';

export const JSON_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.json.title'),
	allowFileUpload: true,
	onImport: async (host, data) => {
		let proto: ReturnType<typeof IndividualSimSettings.fromJson>;
		try {
			proto = fromJsonStringLenient(IndividualSimSettings, data);
		} catch {
			throw new Error(i18n.t('import.json.error_invalid_json'));
		}
		if (proto.player?.equipment) {
			await Database.loadLeftoversIfNecessary(proto.player.equipment);
		}
		host.fromProto(withoutMyEncounter(withViewerStatFilters(proto, host.sim)), individualSimSettingsCategories(proto));
	},
};
