import { Player as PlayerProto } from '@generated/proto/api';
import { APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Cooldowns, EquipmentSpec, Spec } from '@generated/proto/common';
import { IndividualSimSettings, SimSettings } from '@generated/proto/ui';
import i18n from '@i18n/config';
import type { IMessageType, JsonObject } from '@protobuf-ts/runtime';
import { SIM_CATEGORY_OF_KEY, SimSettingCategories } from '@sim/constants/sim_settings';
import { Database } from '@sim/proto/database';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';
import { withoutMyEncounter } from '@sim/state/my_encounter_sharing';
import { withViewerStatFilters } from '@sim/state/stat_filter_sharing';
import { toastManager } from '@ui-kit/Toast';

import { sections } from '../../../../sim/siml/forms.json';
import { SIML_ROTATION_FILE, SIML_SETUP_FILE } from '../exporters/siml';
import { located, parseSiml, specField } from '../siml_client';
import { type ImporterDefinition, ImportProblems } from './types';

// The messages each field of which is a setting of its own, which a section writes one or a few of: the setup, its
// player and its sim settings. Any other message is one setting, such as the consumables or the encounter.
const BY_FIELD: ReadonlySet<IMessageType<any>> = new Set([IndividualSimSettings, PlayerProto, SimSettings]);

/**
 * `page` with each field of `file` that its JSON writes: a message of BY_FIELD field by field, any other whole. The
 * JSON leaves out a field the file does not set, and one it sets to the default, so either keeps the page's value, as
 * it does when the text loads as a preset build. An empty section still empties a message such as the consumables.
 */
const overPage = <T extends object>(type: IMessageType<T>, page: T, file: T): T => {
	const written = type.toJson(file) as JsonObject;
	const [from, to] = [file as Record<string, any>, page as Record<string, any>];
	for (const field of type.fields) {
		if (!(field.jsonName in written)) continue;
		const name = field.oneof ?? field.localName;
		const by = field.kind === 'message' && !field.repeat && !field.oneof && BY_FIELD.has(field.T()) ? field.T() : undefined;
		to[name] = by ? overPage(by, to[name] ?? by.create(), from[name]) : from[name];
	}
	return page;
};

/**
 * Reads `text` as the file `path` names and loads the categories it holds: what the text writes replaces the page's
 * value, and what it leaves out keeps it (`overPage`). So a file holding a rotation leaves the gear and the talents as
 * they are, and a settings: section with no race leaves the race. A file that does not read is refused with each
 * problem at its line and column; `use` is among them, as the site reads no other files. Once `signal` is aborted,
 * nothing loads.
 */
const importSiml = async (host: IndividualSimHost<any>, text: string, path: string, signal?: AbortSignal): Promise<void> => {
	const parsed = await parseSiml(host.sim, path, text);
	if (signal?.aborted) return;
	const errors = parsed.diagnostics.filter(d => !d.warning);
	if (errors.length) throw new ImportProblems(i18n.t('import.siml.error_problems'), errors.map(located));

	const settings = parsed.settings ?? IndividualSimSettings.create();
	const player = settings.player ?? PlayerProto.create();
	// `cooldowns:` is the one rotation section that leaves the rotation alone (sim/siml/SPEC.md §6.2): a file with no
	// other loads no rotation, which the player would load as Auto.
	const categories = parsed.categories
		.map(key => SIM_CATEGORY_OF_KEY.get(key)!)
		.filter(category => category !== SimSettingCategories.Rotation || player.rotation);
	if (!categories.length && !player.cooldowns) throw new Error(i18n.t('import.siml.error_empty'));
	// A gear planner loads no encounter, so a file holding only that has nothing for the page.
	if (host.simDisabled && !player.cooldowns && categories.every(category => category === SimSettingCategories.Encounter)) {
		throw new Error(i18n.t('import.siml.error_nothing_here'));
	}
	// The player loads TypeUnknown as TypeAuto, and a file only writes it with an explicit `type unknown`.
	if (player.rotation?.type === APLRotationType.TypeUnknown) throw new Error(i18n.t('import.siml.error_unknown_type'));

	if (categories.includes(SimSettingCategories.Gear)) {
		await Database.loadLeftoversIfNecessary(EquipmentSpec.create({ items: [...(player.equipment?.items ?? []), ...(player.itemSwap?.items ?? [])] }));
	}
	if (signal?.aborted) return;

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

	// The player keeps the cooldowns in the rotation, which loading one replaces: read the page's before, and set them
	// after it, with what the file's cooldowns: section writes over them.
	const cooldowns =
		(player.cooldowns || categories.includes(SimSettingCategories.Rotation)) &&
		overPage(Cooldowns, host.player.getSimpleCooldowns(), player.cooldowns ?? Cooldowns.create());
	batch(() => {
		// `fromProto` reads no category as every category, which a file of cooldowns alone must not do.
		if (categories.length) {
			host.fromProto(
				withoutMyEncounter(withViewerStatFilters(overPage(IndividualSimSettings, host.toProto(categories), settings), host.sim)),
				categories,
			);
		}
		if (cooldowns) host.player.setSimpleCooldowns(cooldowns);
	});
};

// The sections a file name's double extension may name (sim/siml/SPEC.md §1).
const SECTIONS: ReadonlySet<string> = new Set(sections.map(section => section.name));

/**
 * The name the file was given before a browser renamed it on download, so its double extension reads: a copy
 * such as `rotation.apl (1).siml` or `rotation.apl(1).siml` drops the count, and `Rotation.APL.siml` reads lowercase.
 * A name whose double extension names no section, such as `hunter.sv.siml` or `rotation.apl - Copy.siml`, reads as
 * a full setup, as pasted text does.
 */
const uploadedPath = (fileName: string) => {
	const path = fileName.toLowerCase().replace(/\s*\(\d+\)(?=\.siml$)/, '');
	const section = /\.([^.]+)\.siml$/.exec(path)?.[1];
	return section === undefined || SECTIONS.has(section) ? path : SIML_SETUP_FILE;
};

export const SIML_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.siml.title'),
	allowFileUpload: true,
	onImport: (host, data, fileName, signal) => importSiml(host, data, fileName ? uploadedPath(fileName) : SIML_SETUP_FILE, signal),
};

export const SIML_ROTATION_IMPORTER: ImporterDefinition = {
	title: i18n.t('import.siml_rotation.title'),
	allowFileUpload: true,
	onImport: (host, data, _fileName, signal) => importSiml(host, data, SIML_ROTATION_FILE, signal),
};
