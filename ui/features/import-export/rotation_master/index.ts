// Rotation Master <-> WowSims. The two entry points a simulator calls:
//   importRotationMaster(text, context)  an export string from the addon or rotationmaster.neuromancy.net -> an APL
//   exportRotationMaster(apl, context)   the simulator's APL -> an export string
// Both are pure (no DOM, no network, no state) and both report, item by item, how faithful the conversion is. Neither
// touches anything but the rotation: applying the result is the caller's job (and replaces the APL only).
import { decodeRmExport, encodeRmExport, type RmValue } from './codec';
import { aplToRotation, type FromAplItem, type FromAplSummary } from './from_apl';
import rawMapping from './mapping.json';
import { type ClassKey, type Edition, type Ladders, type SimKinds, type SimRepo } from './report';
import rawSpecs from './rm_specs.json';
import { rotationToApl, type ToAplItem, type ToAplSummary } from './to_apl';

export * from './report';
export { decodeRmExport, encodeRmExport, EXPORT_COLUMNS, stripExportWhitespace, wrapExport } from './codec';
export type { RmValue } from './codec';
export type { FromAplItem, FromAplSummary, RmStep } from './from_apl';
export type { ToAplItem, ToAplSummary } from './to_apl';
export { exportRows, importRows, type ReportRow } from './rows';
export { rotationToApl } from './to_apl';
export { aplToRotation } from './from_apl';

type Json = any;
const mapping = rawMapping as Json;
const specs = (rawSpecs as Json).classes as Record<string, Record<string, { ids: number[]; tree: number | null; editions?: string[] }>>;
const own = (o: unknown, k: string): boolean => o !== null && typeof o === 'object' && Object.prototype.hasOwnProperty.call(o, k);
const isObj = (v: unknown): v is Record<string, Json> => v !== null && typeof v === 'object' && !Array.isArray(v);

/** What the simulator knows about itself. The glue builds this once per simulator page. */
export interface SimContext {
	repo: SimRepo;
	edition: Edition;
	cls: ClassKey;
	/** The simulator's spec as its proto enum names it (SpecElementalShaman), or any word naming the spec. The export is scoped to it and an import for another spec warns. */
	spec?: string;
	/** The APL kinds the simulator reads (field names of its APLValue and APLAction messages). Absent: assume all. */
	kinds?: SimKinds;
	/** Spell rank ladders, on editions with ranks. */
	ladders?: Ladders;
	/** Whether the simulator's ActionID message has a `rank` field. */
	actionIdHasRank?: boolean;
}

export interface Warning {
	code: string;
	params: Record<string, string | number | boolean>;
}

export type ImportResult =
	| { ok: true; apl: Json; items: ToAplItem[]; summary: ToAplSummary; meta: { name: string; description: string; spec: string | null }; warnings: Warning[] }
	| { ok: false; error: { code: string; params: Record<string, string | number | boolean> }; items?: ToAplItem[]; summary?: ToAplSummary };

export type ExportResult =
	| { ok: true; text: string; body: Json; items: FromAplItem[]; summary: FromAplSummary; warnings: Warning[] }
	| { ok: false; error: { code: string; params: Record<string, string | number | boolean> }; items?: FromAplItem[]; summary?: FromAplSummary };

export const EXPORT_WARNING =
	'Rotation Master is a priority list that suggests your next action in the game; the simulator is a model of the fight. Conditions on fight length, cast times and simulator state have no counterpart, so the export may differ from your APL. Steps that are not an exact copy are exported disabled. Check the report before relying on it.';

const fail = (
	code: string,
	params: Record<string, string | number | boolean> = {},
): { ok: false; error: { code: string; params: Record<string, string | number | boolean> } } => ({
	ok: false,
	error: { code, params },
});

/** The class a proto enum name ("ClassWarrior", "ClassDeathKnight") or a plain word names, or null. */
export function classKey(name: string | undefined | null): ClassKey | null {
	if (typeof name !== 'string') return null;
	const direct = mapping.classes[name];
	if (typeof direct === 'string') return direct as ClassKey;
	const low = name.toLowerCase().replace(/[^a-z]/g, '');
	return typeof mapping.classes[low] === 'string' ? (mapping.classes[low] as ClassKey) : null;
}

// ---- specializations -------------------------------------------------------------------------------------------------------
/**
 * The specialization a simulator spec names, in Rotation Master's words. `named` is false when the spec names none (a
 * class-only spec such as SpecMage, or a role such as SpecDpsWarrior): the rotation then applies to any specialization.
 * `slug` is null when it names one Rotation Master does not know.
 */
export function specInfo(cls: ClassKey, word: string | undefined | null): { named: boolean; slug: string | null } {
	if (typeof word !== 'string') return { named: false, slug: null };
	const slugs: string[] = mapping.classSpecs[cls] || [];
	let low = word
		.replace(/^Spec/, '')
		.replace(/(Druid|Hunter|Mage|Paladin|Priest|Rogue|Shaman|Warlock|Warrior|Deathknight|DeathKnight|Monk)$/, '')
		.toLowerCase()
		.replace(/[^a-z-]/g, '');
	if (low === '') return { named: false, slug: null };
	if (own(SIM_SPEC_WORDS, low)) low = SIM_SPEC_WORDS[low];
	else if (own(mapping.specAliases, low)) {
		const alias = mapping.specAliases[low];
		if (alias === null) return { named: false, slug: null };
		low = alias;
	}
	if (slugs.includes(low)) return { named: true, slug: low };
	const squeezed = low.replace(/-/g, '');
	return { named: true, slug: slugs.find(s => s.replace(/-/g, '') === squeezed) || null };
}

/** The slug Rotation Master uses for a specialization word of a class ("ElementalShaman", "elemental", "feraltank"), or null. */
export function specSlug(cls: ClassKey, word: string | undefined | null): string | null {
	return specInfo(cls, word).slug;
}

/** The specialization (slug) a rotation body is scoped to, from `specs` ids or a PRIMARY_TALENT_TREE switch; null when it applies to any. */
function scopeOf(body: Json, cls: ClassKey, edition: Edition): { slug: string | null; any: boolean } {
	const table = specs[cls] || {};
	const ids: number[] = Array.isArray(body.specs) ? body.specs.filter((x: unknown): x is number => typeof x === 'number') : [];
	if (ids.length) {
		const hit = Object.entries(table).find(([, s]) => s.ids.some(id => ids.includes(id)));
		return { slug: hit ? hit[0] : null, any: false };
	}
	const sw = body.switch;
	if (isObj(sw) && sw.type === 'PRIMARY_TALENT_TREE' && typeof sw.tree === 'number') {
		const hit = Object.entries(table).find(([slug, s]) => s.tree === sw.tree && (!s.editions || s.editions.includes(edition)) && slug);
		return { slug: hit ? hit[0] : null, any: false };
	}
	return { slug: null, any: true };
}

// Words only the simulators use for a spec, where Rotation Master's table has none: the TBC simulator's two Feral spells are
// both the Feral Combat tree.
const SIM_SPEC_WORDS: Record<string, string> = { feralcat: 'feral', feralbear: 'feral' };

const TREE_EDITIONS: ReadonlyArray<string> = ['era', 'tbc', 'forever'];

// ---- import ------------------------------------------------------------------------------------------------------------------
export interface ImportOptions {
	/** Keep an item that is disabled in Rotation Master as a hidden APL item (default true). */
	keepDisabled?: boolean;
	/** Hide the items that lost a condition (default false). */
	hideLossy?: boolean;
}

/** An export string (addon or website) to an APL. Never touches anything else of the player's settings. */
export function importRotationMaster(text: string, ctx: SimContext, options: ImportOptions = {}): ImportResult {
	const decoded = decodeRmExport(text);
	if (!decoded.ok) return fail('decode', { reason: decoded.reason });
	const root: Json = decoded.value;
	if (!isObj(root)) return fail('not-a-rotation');
	let body: Json = root;
	const warnings: Warning[] = [];
	if (decoded.envelope) {
		if (!isObj(root.body)) return fail('not-a-rotation');
		body = root.body;
		const scope = root.applicability;
		if (isObj(scope)) {
			if (typeof scope.class === 'string' && scope.class !== ctx.cls) return fail('class-mismatch', { declared: scope.class, simulator: ctx.cls });
			if (Array.isArray(scope.families) && scope.families.length && !scope.families.includes(ctx.edition)) {
				return fail('edition-mismatch', { declared: scope.families.filter((x: unknown) => typeof x === 'string').join(', '), simulator: ctx.edition });
			}
		}
	} else if (!(Array.isArray(body.rotation) || Array.isArray(body.cooldowns))) {
		return fail('not-a-rotation');
	}
	for (const key of ['rotation', 'cooldowns']) {
		if (body[key] !== undefined && !Array.isArray(body[key])) return fail('invalid-structure', { field: key });
	}
	const simSlug = specSlug(ctx.cls, ctx.spec);
	const scope = scopeOf(body, ctx.cls, ctx.edition);
	if (!scope.any && simSlug && scope.slug && scope.slug !== simSlug)
		warnings.push({ code: 'spec-mismatch', params: { rotation: scope.slug, simulator: simSlug } });
	if (!scope.any && !scope.slug) warnings.push({ code: 'spec-unreadable', params: {} });
	const result = rotationToApl(body, {
		edition: ctx.edition,
		cls: ctx.cls,
		kinds: ctx.kinds,
		ladders: ctx.ladders,
		actionIdHasRank: ctx.actionIdHasRank,
		keepDisabled: options.keepDisabled,
		hideLossy: options.hideLossy,
	});
	if (!result.ok) return fail(result.error.code, result.error);
	if (!result.summary.actions && !result.summary.prepull)
		return { ok: false, error: { code: 'nothing-imported', params: {} }, items: result.items, summary: result.summary };
	const str = (v: unknown): string => (typeof v === 'string' ? v : '');
	return {
		ok: true,
		apl: result.apl,
		items: result.items,
		summary: result.summary,
		meta: { name: str(body.name), description: str(body.description), spec: scope.slug },
		warnings,
	};
}

// ---- export ------------------------------------------------------------------------------------------------------------------
export interface ExportOptions {
	/** Export approximate steps enabled instead of disabled (default false). */
	enableApproximate?: boolean;
	/** Item ids of this edition's mana potions (the simulator's resource potion becomes an item step with them). */
	potions?: ReadonlyArray<number>;
	/** Accept an APL whatever its `type` says: the sim resolves an Auto or Simple rotation into a priority list, and that list is what to export. */
	anyType?: boolean;
	/** A name for the rotation (default "WowSims <spec> <class>"). */
	name?: string;
}

const titleCase = (s: string): string => s.toLowerCase().replace(/(^|[ -])([a-z])/g, (_m, a: string, b: string) => a + b.toUpperCase());
const CLASS_NAME: Record<string, string> = { DEATHKNIGHT: 'Death Knight' };

/** The simulator's APL (proto-JSON, `APLRotation.toJson`) to an export string, with the report of what changed. */
export function exportRotationMaster(apl: Json, ctx: SimContext, options: ExportOptions = {}): ExportResult {
	if (!isObj(apl)) return fail('not-an-apl', { type: 'none' });
	const type = apl.type;
	if (!options.anyType && type !== undefined && type !== 'TypeAPL' && type !== 3 && type !== 'TypeUnknown' && type !== 0) {
		return fail('not-an-apl', { type: String(type) });
	}
	if (!Array.isArray(apl.priorityList) && !Array.isArray(apl.prepullActions)) return fail('empty-apl');
	if (!apl.priorityList?.length && !apl.prepullActions?.length) return fail('empty-apl');
	const converted = aplToRotation(apl, { edition: ctx.edition, cls: ctx.cls, enableApproximate: options.enableApproximate, potions: options.potions });
	if (!converted.steps.length) return { ok: false, error: { code: 'nothing-exported', params: {} }, items: converted.items, summary: converted.summary };
	const warnings: Warning[] = [];
	const slug = specSlug(ctx.cls, ctx.spec);
	const clsName = CLASS_NAME[ctx.cls] ?? titleCase(ctx.cls);
	const specName = slug ? titleCase(slug) : '';
	const body: Json = {
		name: options.name ?? `WowSims ${specName ? specName + ' ' : ''}${clsName}`,
		description: `Exported from the WowSims ${ctx.repo} simulator: ${converted.summary.steps} steps, ${converted.summary.disabled} disabled, ${converted.summary.skipped} left out.`,
		tags: ['wowsims'],
	};
	const table = specs[ctx.cls] || {};
	if (slug && table[slug] && (!table[slug].editions || table[slug].editions!.includes(ctx.edition))) {
		if (TREE_EDITIONS.includes(ctx.edition)) {
			if (table[slug].tree !== null) body.switch = { type: 'PRIMARY_TALENT_TREE', tree: table[slug].tree };
		} else body.specs = [...table[slug].ids];
	} else if (specInfo(ctx.cls, ctx.spec).named) warnings.push({ code: 'spec-unreadable', params: {} });
	body.rotation = converted.steps.map(step => ({ ...step, conditions: step.conditions ?? [] }));
	body.cooldowns = [];
	const encoded = encodeRmExport(body as RmValue, { compress: true });
	if (!encoded.ok) return { ok: false, error: { code: 'encode', params: { reason: encoded.reason } }, items: converted.items, summary: converted.summary };
	return { ok: true, text: encoded.text, body, items: converted.items, summary: converted.summary, warnings };
}

// ---- words -------------------------------------------------------------------------------------------------------------------------
const ERRORS: Record<string, (p: Record<string, string | number | boolean>) => string> = {
	decode: p => `This is not a Rotation Master export (${String(p.reason)}). Copy the whole export string from the addon or the website.`,
	'not-a-rotation': () => 'This export is not a rotation (it may be a profile or an item set).',
	'invalid-structure': p => `The rotation is malformed (${String(p.field)} is not a list).`,
	'class-mismatch': p => `This rotation is for ${String(p.declared)}, but this simulator is on ${String(p.simulator)}.`,
	'edition-mismatch': p => `This rotation is for the ${String(p.declared)} edition, but this simulator is for ${String(p.simulator)}.`,
	'nothing-imported': () => 'Nothing in this rotation could be converted.',
	'nothing-exported': () => 'Nothing in this APL could be converted.',
	'empty-apl': () => 'This APL has no actions.',
	'not-an-apl': p =>
		`The simulator's rotation is not a custom APL (it is ${String(p.type)}). Pick a custom APL rotation, or one of the preset APLs, then export again.`,
	encode: p => `The rotation could not be written (${String(p.reason)}).`,
	'edition-unsupported': p => `Unsupported edition ${String(p.edition)}.`,
	'not-rotation': () => 'This is not a rotation.',
};
export function describeError(error: { code: string; params: Record<string, string | number | boolean> }): string {
	return ERRORS[error.code] ? ERRORS[error.code](error.params) : error.code;
}
export function describeWarning(warning: Warning): string {
	if (warning.code === 'spec-mismatch') {
		const rotation = titleCase(String(warning.params.rotation));
		const simulator = titleCase(String(warning.params.simulator));
		return `This rotation is for ${rotation}, but the simulator is on ${simulator}. It was imported anyway.`;
	}
	if (warning.code === 'spec-unreadable') return 'The specialization this rotation is for could not be read; it was imported without one.';
	return warning.code;
}

/** A one-line count of a report summary ("12 exact, 3 approximate, 1 lossy, 2 unsupported"). */
export function summaryLine(s: { exact: number; approximate: number; lossy: number; unsupported: number }): string {
	return `${s.exact} exact, ${s.approximate} approximate, ${s.lossy} lossy, ${s.unsupported} unsupported`;
}
