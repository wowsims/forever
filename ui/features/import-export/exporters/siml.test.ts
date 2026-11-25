// The SIML exporters ask the sim to print what the host gives them. The sim's printing itself is tested in Go
// (sim/siml/api_test.go).
import { Player } from '@generated/proto/api';
import { APLListItem, APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { Encounter, Spec } from '@generated/proto/common';
import { IndividualSimSettings, SimlDiagnostic, SimlPrintResult } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { getEnumValues } from '@sim/utils/collections';
import { fakeHost } from '@testing/sim/fake_host';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SIML_EXPORTER, SIML_ROTATION_EXPORTER } from './siml';
import type { ExportCategories } from './types';

const printSiml = vi.hoisted(() => vi.fn());
vi.mock('../siml_client', async importOriginal => ({ ...(await importOriginal<typeof import('../siml_client')>()), printSiml }));

// The key, and what it is filled with, so a test can tell which message was given and about what.
vi.mock('@i18n/config', () => ({ default: { t: (key: string, values?: object) => (values ? `${key} ${JSON.stringify(values)}` : key) } }));

const TEXT = 'siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n';

const rotation = APLRotation.fromJson({ type: 'TypeAPL', priorityList: [{ action: { castSpell: { spellId: { spellId: 2643 } } } }] });
// What an Auto or Simple rotation resolves to: the list the sim generates for it, beside the type and the simple settings.
const generated = (type: 'TypeAuto' | 'TypeSimple') =>
	APLRotation.fromJson({
		type,
		simple: { cooldowns: { hpPercentForDefensives: 0.4 } },
		priorityList: rotation.priorityList.map(item => APLListItem.toJson(item)),
	});
const settings = IndividualSimSettings.create({ player: Player.create({ talentsString: '-0053', rotation }) });

const toProto = vi.fn((_categories: Array<SimSettingCategories>) => settings);
const host = fakeHost({ player: { getSpec: () => Spec.SpecHunter, getResolvedAplRotation: () => rotation }, toProto });

const ticked = (...categories: Array<SimSettingCategories>) =>
	Object.fromEntries(
		(getEnumValues(SimSettingCategories) as Array<SimSettingCategories>).map(category => [category, categories.includes(category)]),
	) as ExportCategories;

// The request the sim was asked to print.
const asked = () => printSiml.mock.lastCall![1];

// What the sim answers a print that does not read back with: no text, and the problems of the read.
const unreadable = SimlPrintResult.create({
	diagnostics: [SimlDiagnostic.create({ startLine: 4, startCol: 1, endLine: 4, endCol: 10, code: 'no-spec', message: 'a rotation needs a spec line' })],
});

beforeEach(() => {
	printSiml.mockReset();
	printSiml.mockResolvedValue(SimlPrintResult.create({ text: TEXT }));
});

describe('SIML_EXPORTER', () => {
	it('gives a promise of the text the sim prints for the ticked categories', async () => {
		const data = SIML_EXPORTER.getData(host, ticked(SimSettingCategories.Rotation));
		expect(data).toBeInstanceOf(Promise);
		expect(await data).toBe(TEXT);
		expect(toProto).toHaveBeenLastCalledWith([SimSettingCategories.Rotation]);
		expect(printSiml).toHaveBeenCalledWith(host.sim, { spec: Spec.SpecHunter, settings, categories: ['r'] });
	});

	// So that importing the file resets that category, as a JSON import of the same categories does.
	it('asks for each ticked category by its letter', async () => {
		await SIML_EXPORTER.getData(host, ticked(SimSettingCategories.Talents, SimSettingCategories.Rotation, SimSettingCategories.Consumes));
		expect(asked().categories).toEqual(['t', 'r', 'c']);
	});

	it('asks for every section the settings hold when no category is ticked, as the JSON export does', async () => {
		await SIML_EXPORTER.getData(host, ticked());
		expect(toProto).toHaveBeenLastCalledWith([]);
		expect(asked().categories).toEqual([]);
	});

	it('says nothing of the encounter being an entry of My encounters', async () => {
		const own = fakeHost({
			player: { getSpec: () => Spec.SpecHunter },
			toProto: () => IndividualSimSettings.create({ encounter: Encounter.create({ name: 'Mine', myEncounter: true }) }),
		});
		await SIML_EXPORTER.getData(own, ticked(SimSettingCategories.Encounter));
		expect(asked().settings.encounter).toMatchObject({ name: 'Mine', myEncounter: false });
	});

	it('refuses text that would not import, with its problems', async () => {
		printSiml.mockResolvedValue(unreadable);
		await expect(SIML_EXPORTER.getData(host, ticked(SimSettingCategories.Rotation))).rejects.toThrow(
			'export.siml.error_unreadable {"problems":"4:1 a rotation needs a spec line"}',
		);
	});

	it('downloads as a full setup', () => {
		expect(SIML_EXPORTER.downloadFileName).toBe('wowsims.siml');
	});
});

describe('SIML_ROTATION_EXPORTER', () => {
	// Kept as Auto or Simple, the import would have the sim generate the list again, and the edits would be lost.
	it.each(['TypeAuto', 'TypeSimple'] as const)('writes the list a %s rotation generates as an APL rotation, to edit', async type => {
		const resolved = fakeHost({ player: { getSpec: () => Spec.SpecHunter, getResolvedAplRotation: () => generated(type) } });
		expect(await SIML_ROTATION_EXPORTER.getData(resolved, ticked())).toBe(TEXT);
		const printed = asked().settings.player.rotation;
		expect(printed.type).toBe(APLRotationType.TypeAPL);
		expect(printed.simple).toBeUndefined();
		expect(printed.priorityList).toEqual(rotation.priorityList);
	});

	it('refuses text that would not import, with its problems', async () => {
		printSiml.mockResolvedValue(unreadable);
		await expect(SIML_ROTATION_EXPORTER.getData(host, ticked())).rejects.toThrow(/^export\.siml\.error_unreadable .*spec line/);
	});

	it('prints the rotation the sim runs as an x.apl.siml', async () => {
		expect(await SIML_ROTATION_EXPORTER.getData(host, ticked())).toBe(TEXT);
		expect(asked()).toEqual(expect.objectContaining({ section: 'apl', spec: Spec.SpecHunter }));
		expect(SIML_ROTATION_EXPORTER.downloadFileName).toBe('rotation.apl.siml');
	});
});
