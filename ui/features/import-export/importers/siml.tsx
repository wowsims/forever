import { Player as PlayerProto } from '@generated/proto/api';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { EquipmentSpec, Spec } from '@generated/proto/common';
import { IndividualSimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SIM_CATEGORY_OF_KEY, SimSettingCategories } from '@sim/constants/sim_settings';
import { Database } from '@sim/proto/database';
import type { IndividualSimHost } from '@sim/sim_host';
import { withViewerStatFilters } from '@sim/state/stat_filter_sharing';
import { toastManager } from '@ui-kit/Toast';

import { SIML_ROTATION_FILE, SIML_SETUP_FILE } from '../exporters/siml';
import { located, parseSiml, specField } from '../siml_client';
import { type ImporterDefinition, ImportProblems } from './types';

/**
 * Reads `text` as the file `path` names and loads exactly the categories it holds, so a file holding a rotation
 * leaves the gear and the talents as they are, and an empty section resets its category. A file that does not read
 * is refused with each problem at its line and column; `use` is among them, as the site reads no other files.
 */
const importSiml = async (host: IndividualSimHost<any>, text: string, path: string): Promise<void> => {
	const parsed = await parseSiml(host.sim, path, text);
	const errors = parsed.diagnostics.filter(d => !d.warning);
	if (errors.length) throw new ImportProblems(i18n.t('import.siml.error_problems'), errors.map(located));

	const categories = parsed.categories.map(key => SIM_CATEGORY_OF_KEY.get(key)!);
	// `fromProto` reads no category as every category, which a file holding none must not do.
	if (!categories.length) throw new Error(i18n.t('import.siml.error_empty'));
	// A gear planner loads no encounter, so a file holding only that has nothing for the page.
	if (host.simDisabled && categories.every(category => category === SimSettingCategories.Encounter)) {
		throw new Error(i18n.t('import.siml.error_nothing_here'));
	}
	const settings = parsed.settings ?? IndividualSimSettings.create();
	// A section file with no spec line, such as an x.encounter.siml, sets no player, and `fromProto` loads nothing
	// without one.
	const player = (settings.player ??= PlayerProto.create());
	// The player loads TypeUnknown as TypeAuto, and a file only writes it with an explicit `type unknown`.
	if (categories.includes(SimSettingCategories.Rotation) && (player.rotation?.type ?? APLRotationType.TypeUnknown) === APLRotationType.TypeUnknown) {
		throw new Error(i18n.t('import.siml.error_unknown_type'));
	}

	const page = host.player.getSpec();
	if (parsed.spec !== Spec.SpecUnknown && parsed.spec !== page) {
		toastManager.add({
			variant: 'warning',
			body: i18n.t('import.siml.warning_spec', { file: specField(parsed.spec), page: specField(page) }),
			delay: 8000,
		});
	}
	const warnings = parsed.diagnostics.filter(d => d.warning);
	if (warnings.length) {
		toastManager.add({
			variant: 'warning',
			body: (
				<>
					<p>{i18n.t('import.siml.warnings')}</p>
					<ul>
						{warnings.map(located).map((line, i) => (
							<li key={i}>{line}</li>
						))}
					</ul>
				</>
			),
			delay: 8000,
		});
	}

	if (categories.includes(SimSettingCategories.Gear)) {
		await Database.loadLeftoversIfNecessary(EquipmentSpec.create({ items: [...(player.equipment?.items ?? []), ...(player.itemSwap?.items ?? [])] }));
	}
	host.fromProto(withViewerStatFilters(settings, host.sim), categories);
};

/**
 * The name the file was given before a browser renamed it on download, so its double extension reads: a copy
 * such as `rotation.apl (1).siml` or `rotation.apl(1).siml` drops the count, and `Rotation.APL.siml` reads lowercase.
 */
const uploadedPath = (fileName: string) => fileName.toLowerCase().replace(/\s*\(\d+\)(?=\.siml$)/, '');

export const SIML_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.siml.title'),
	allowFileUpload: true,
	onImport: (host, data, fileName) => importSiml(host, data, fileName ? uploadedPath(fileName) : SIML_SETUP_FILE),
};

export const SIML_ROTATION_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.siml_rotation.title'),
	allowFileUpload: true,
	onImport: (host, data) => importSiml(host, data, SIML_ROTATION_FILE),
};
