// The SIML importer loads what the sim reads from the text into a host that records what it is given: which
// categories, and the settings they come from. The sim's reading itself is tested in Go (sim/siml/api_test.go).
import { Player } from '@generated/proto/api';
import { APLRotation, APLRotation_Type as APLRotationType } from '@generated/proto/apl';
import { ConsumesSpec, Cooldowns, Encounter, EquipmentSpec, ItemSpec, Spec } from '@generated/proto/common';
import { IndividualSimSettings, SimlDiagnostic, SimlParseResult } from '@generated/proto/ui';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { fakeHost } from '@testing/sim/fake_host';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { SIML_IMPORTER, SIML_ROTATION_IMPORTER } from './siml';
import { ImportProblems } from './types';

const parseSiml = vi.hoisted(() => vi.fn());
vi.mock('../siml_client', async importOriginal => ({ ...(await importOriginal<typeof import('../siml_client')>()), parseSiml }));

const loadLeftovers = vi.hoisted(() => vi.fn().mockResolvedValue(undefined));
vi.mock('@sim/proto/database', () => ({ Database: { loadLeftoversIfNecessary: loadLeftovers } }));

// The key, and what it is filled with, so a test can tell which message was given and about what.
vi.mock('@i18n/config', () => ({ default: { t: (key: string, values?: object) => (values ? `${key} ${JSON.stringify(values)}` : key) } }));

const toasts = vi.hoisted(() => [] as Array<{ variant: string; body: unknown }>);
vi.mock('@ui-kit/Toast', async importOriginal => ({
	...(await importOriginal<typeof import('@ui-kit/Toast')>()),
	toastManager: {
		add: (options: { variant: string; body: unknown }) => {
			toasts.push(options);
			return '';
		},
		close: () => {},
	},
}));

const fromProto = vi.fn<(settings: IndividualSimSettings, categories: Array<SimSettingCategories>) => void>();
// A page with nothing of its own, so what the importer loads is what the text holds. The cooldowns are read before
// a rotation loads and set after it.
const toProto = () => IndividualSimSettings.create({ player: Player.create() });
const cooldowns = { getSimpleCooldowns: () => Cooldowns.create(), setSimpleCooldowns: vi.fn() };
const hostFor = (spec: Spec) => fakeHost({ player: { getSpec: () => spec, ...cooldowns }, toProto, fromProto });
const hunter = hostFor(Spec.SpecHunter);

const loaded = () => {
	expect(fromProto).toHaveBeenCalledTimes(1);
	const [settings, categories] = fromProto.mock.calls[0];
	return { settings, categories: [...categories].sort() };
};

const problems = async (promise: Promise<void>) => {
	const error = await promise.then(
		() => undefined,
		(e: unknown) => e,
	);
	expect(error).toBeInstanceOf(ImportProblems);
	return (error as ImportProblems).problems;
};

// What the sim reads from a text: settings, the letters of the categories it holds, and its spec line.
const reads = (player: Partial<Player>, categories: string[], spec = Spec.SpecHunter, diagnostics: SimlDiagnostic[] = []) =>
	parseSiml.mockResolvedValue(
		SimlParseResult.create({ settings: IndividualSimSettings.create({ player: Player.create(player) }), categories, spec, diagnostics }),
	);

const rotation = APLRotation.fromJson({ type: 'TypeAPL', priorityList: [{ action: { castSpell: { spellId: { spellId: 2643 } } } }] });
const readsRotation = () => reads({ rotation }, ['r']);
const readsGear = () => reads({ equipment: EquipmentSpec.create({ items: [ItemSpec.create({ id: 16566 })] }) }, ['g'], Spec.SpecUnknown);

const pathRead = () => parseSiml.mock.calls[0][1];

beforeEach(() => {
	toasts.length = 0;
	fromProto.mockClear();
	loadLeftovers.mockClear();
	parseSiml.mockReset();
});

describe('a rotation', () => {
	it('loads the rotation alone, so the gear and the talents stay as they are', async () => {
		readsRotation();
		await SIML_ROTATION_IMPORTER.onImport(hunter, 'the text');
		expect(parseSiml).toHaveBeenCalledWith(hunter.sim, 'rotation.apl.siml', 'the text');
		const { settings, categories } = loaded();
		expect(categories).toEqual([SimSettingCategories.Rotation]);
		expect(settings.player?.rotation?.priorityList).toHaveLength(1);
		expect(loadLeftovers).not.toHaveBeenCalled();
	});

	it('reads as an x.apl.siml whatever the uploaded file is named', async () => {
		readsRotation();
		await SIML_ROTATION_IMPORTER.onImport(hunter, 'the text', 'wowsims.siml');
		expect(pathRead()).toBe('rotation.apl.siml');
	});

	it('refuses a type the player would load as Auto', async () => {
		reads({ rotation: APLRotation.create({ type: APLRotationType.TypeUnknown }) }, ['r']);
		await expect(SIML_ROTATION_IMPORTER.onImport(hunter, 'the text')).rejects.toThrow('import.siml.error_unknown_type');
		expect(fromProto).not.toHaveBeenCalled();
	});
});

describe('pasted text', () => {
	it('loads an encounter the page holds as no entry of My encounters, when the text leaves it as it is', async () => {
		const own = fakeHost({
			player: { getSpec: () => Spec.SpecHunter, ...cooldowns },
			toProto: () => IndividualSimSettings.create({ player: Player.create(), encounter: Encounter.create({ name: 'Mine', myEncounter: true }) }),
			fromProto,
		});
		parseSiml.mockResolvedValue(SimlParseResult.create({ settings: IndividualSimSettings.create(), categories: ['e'], spec: Spec.SpecHunter }));
		await SIML_IMPORTER.onImport(own, 'the text');
		expect(loaded().settings.encounter).toMatchObject({ name: 'Mine', myEncounter: false });
	});

	it('reads as a full setup, and loads each category it holds', async () => {
		reads({ talentsString: '-0053', rotation, consumables: ConsumesSpec.create() }, ['t', 'r', 'c']);
		await SIML_IMPORTER.onImport(hunter, 'the text');
		expect(pathRead()).toBe('wowsims.siml');
		const { settings, categories } = loaded();
		expect(categories).toEqual([SimSettingCategories.Talents, SimSettingCategories.Rotation, SimSettingCategories.Consumes]);
		expect(settings.player?.talentsString).toBe('-0053');
	});

	it('refuses text that holds no category, which would load every one', async () => {
		reads({}, []);
		await expect(SIML_IMPORTER.onImport(hunter, 'the text')).rejects.toThrow('import.siml.error_empty');
		expect(fromProto).not.toHaveBeenCalled();
	});

	it('refuses with what the sim could not take at all', async () => {
		parseSiml.mockRejectedValue(new Error('no SIML name table is read'));
		await expect(SIML_IMPORTER.onImport(hunter, 'the text')).rejects.toThrow('no SIML name table is read');
		expect(fromProto).not.toHaveBeenCalled();
	});
});

describe('an uploaded file', () => {
	it('reads as the kind its name gives, and loads the leftover items its gear needs', async () => {
		readsGear();
		await SIML_IMPORTER.onImport(hunter, 'the text', 'p1.gear.siml');
		expect(pathRead()).toBe('p1.gear.siml');
		expect(loaded().categories).toEqual([SimSettingCategories.Gear]);
		expect(loadLeftovers).toHaveBeenCalledTimes(1);
		expect(loadLeftovers.mock.calls[0][0].items[0].id).toBe(16566);
	});
});

describe('a file a browser renamed on download', () => {
	it.each([
		['p1.gear (1).siml', 'p1.gear.siml'],
		['p1.gear(1).siml', 'p1.gear.siml'],
		['P1.GEAR.siml', 'p1.gear.siml'],
		['P1.Gear (12).siml', 'p1.gear.siml'],
		['rotation.apl (1).siml', 'rotation.apl.siml'],
		['rotation.apl(1).siml', 'rotation.apl.siml'],
		['Rotation.APL.siml', 'rotation.apl.siml'],
	])('reads %s as %s', async (name, path) => {
		readsGear();
		await SIML_IMPORTER.onImport(hunter, 'the text', name);
		expect(pathRead()).toBe(path);
	});
});

// Pasted, the same text reads as a full setup, and so does the file.
describe('a file whose name has a dot that names no section', () => {
	it.each([
		['hunter.sv.siml', 'wowsims.siml'],
		['bm.siml', 'bm.siml'],
	])('reads %s as %s', async (name, path) => {
		readsRotation();
		await SIML_IMPORTER.onImport(hunter, 'the text', name);
		expect(pathRead()).toBe(path);
	});
});

describe('an import the dialog dropped', () => {
	it('loads nothing when the read ends after the dialog closed', async () => {
		let answer: (result: SimlParseResult) => void = () => {};
		parseSiml.mockReturnValue(new Promise(resolve => (answer = resolve)));
		const dialog = new AbortController();
		const imported = SIML_ROTATION_IMPORTER.onImport(hunter, 'the text', undefined, dialog.signal);

		dialog.abort();
		answer(SimlParseResult.create({ settings: IndividualSimSettings.create({ player: Player.create({ rotation }) }), categories: ['r'] }));
		await imported;

		expect(fromProto).not.toHaveBeenCalled();
	});
});

describe('a gear planner', () => {
	const readsEncounter = (...more: string[]) => reads({}, [...more, 'e'], Spec.SpecUnknown);
	const planner = fakeHost({ player: { getSpec: () => Spec.SpecHolyPaladin, ...cooldowns }, simDisabled: true, toProto, fromProto });

	it('refuses a file holding only the encounter, which it does not import', async () => {
		readsEncounter();
		await expect(SIML_IMPORTER.onImport(planner, 'the text', 'raid.encounter.siml')).rejects.toThrow('import.siml.error_nothing_here');
		expect(fromProto).not.toHaveBeenCalled();
	});

	it('imports a file holding the encounter and more', async () => {
		readsEncounter('t');
		await SIML_IMPORTER.onImport(planner, 'the text');
		expect(loaded().categories).toEqual([SimSettingCategories.Talents, SimSettingCategories.Encounter].sort());
	});

	it('leaves a page that sims to import the encounter', async () => {
		readsEncounter();
		await SIML_IMPORTER.onImport(hunter, 'the text', 'raid.encounter.siml');
		expect(loaded().categories).toEqual([SimSettingCategories.Encounter]);
	});
});

describe('what the importer refuses or warns of', () => {
	const problem = (line: number, col: number, message: string, warning = false) =>
		SimlDiagnostic.create({ startLine: line, startCol: col, endLine: line, endCol: col + 1, warning, code: 'x', message });

	it('lists the errors at their lines and columns', async () => {
		parseSiml.mockResolvedValue(
			SimlParseResult.create({ spec: Spec.SpecHunter, diagnostics: [problem(5, 8, 'no spell is named hunters_mrak; did you mean hunters_mark?')] }),
		);
		expect(await problems(SIML_IMPORTER.onImport(hunter, 'the text'))).toEqual(['5:8 no spell is named hunters_mrak; did you mean hunters_mark?']);
		expect(fromProto).not.toHaveBeenCalled();
	});

	it("warns that the text is for another spec than the page's, and imports it", async () => {
		readsRotation();
		await SIML_IMPORTER.onImport(hostFor(Spec.SpecDpsWarrior), 'the text');
		expect(toasts).toEqual([{ variant: 'warning', body: 'import.siml.warning_spec {"file":"hunter","page":"dps_warrior"}', delay: 8000 }]);
		expect(loaded().categories).toEqual([SimSettingCategories.Rotation]);
	});

	it('lists the warnings in a toast, and imports', async () => {
		reads({ rotation }, ['r'], Spec.SpecHunter, [problem(6, 3, 'no group is named Nowhere', true)]);
		await SIML_ROTATION_IMPORTER.onImport(hunter, 'the text');
		expect(toasts).toHaveLength(1);
		expect(toasts[0].variant).toBe('warning');
		expect(loaded().settings.player?.rotation?.priorityList).toHaveLength(1);
	});
});
