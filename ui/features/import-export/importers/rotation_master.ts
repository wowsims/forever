import { APLAction, APLRotation, APLRotation_Type, APLValue } from '@generated/proto/apl';
import { ActionID, Class, Spec } from '@generated/proto/common';
import type { IndividualSimHost } from '@sim/sim_host';
import { batch } from '@sim/state/batch';

import {
	classKey,
	type ExportOptions,
	type ExportResult,
	exportRotationMaster,
	type ImportOptions,
	type ImportResult,
	importRotationMaster,
	ladderIndex,
	type Ladders,
	laddersOf,
	type SimContext,
} from '../rotation_master';

// What this simulator is, in Rotation Master's words. `ranks` is whether its spells have ranks (Rotation Master then
// names a spell by its first rank and the simulator needs the rank its character casts).
export const ROTATION_MASTER_SIM = { repo: 'forever', edition: 'forever', ranks: true } as const;

const fieldNames = (type: { fields: ReadonlyArray<{ localName: string }> }): Set<string> => new Set(type.fields.map(field => field.localName));

// The ladders of ranks come from the spell list the database already carries (a spell's name and rank), once per database.
const laddersByDatabase = new WeakMap<object, Ladders>();
const laddersOfDatabase = (host: IndividualSimHost<any>): Ladders => {
	const db = host.sim.db;
	let ladders = laddersByDatabase.get(db);
	if (!ladders) {
		ladders = ladderIndex(laddersOf(db.getSpellIcons()));
		laddersByDatabase.set(db, ladders);
	}
	return ladders;
};

/** What this simulator and this player are, for the converters. */
export const rotationMasterContext = (host: IndividualSimHost<any>): SimContext => {
	const cls = classKey(Class[host.player.getClass()]);
	if (!cls) throw new Error(`Rotation Master has no class ${Class[host.player.getClass()]}`);
	return {
		repo: ROTATION_MASTER_SIM.repo,
		edition: ROTATION_MASTER_SIM.edition,
		cls,
		spec: Spec[host.player.getSpec()],
		kinds: { values: fieldNames(APLValue), actions: fieldNames(APLAction) },
		actionIdHasRank: fieldNames(ActionID).has('rank'),
		ladders: ROTATION_MASTER_SIM.ranks ? laddersOfDatabase(host) : undefined,
	};
};

/** Reads and converts a Rotation Master export; changes nothing. */
export const checkRotationMasterImport = (host: IndividualSimHost<any>, text: string, options: ImportOptions = {}): ImportResult =>
	importRotationMaster(text, rotationMasterContext(host), options);

/**
 * Puts an imported APL in as the player's rotation. This replaces the APL and only the APL: the rotation type becomes APL,
 * the pre-pull actions, the priority list, the groups and the variables are the import's, and everything else of the
 * rotation (the Simple rotation's settings) and of the player (gear, talents, glyphs, consumes, buffs, race, professions,
 * the encounter, the sim settings) is left exactly as it was. Unlike the JSON importer, which loads every category and
 * resets the ones the file does not carry.
 */
export const applyRotationMasterImport = (host: IndividualSimHost<any>, apl: unknown): void => {
	const imported = APLRotation.fromJson(apl as never);
	const next = APLRotation.clone(host.player.aplRotation);
	next.type = APLRotation_Type.TypeAPL;
	next.prepullActions = imported.prepullActions;
	next.priorityList = imported.priorityList;
	next.groups = [];
	next.valueVariables = [];
	batch(() => host.player.setAplRotation(next));
};

/** What the export is made from: the player's APL, or the one the simulator generates for an Auto or Simple rotation. */
export interface RotationMasterExport {
	result: ExportResult;
	/** Set when the player's rotation is Auto or Simple and the simulator's generated priority list was exported. */
	generatedFrom: 'TypeAuto' | 'TypeSimple' | null;
}

/** The player's rotation as a Rotation Master export string. Reads only. */
export const exportRotationMasterText = (host: IndividualSimHost<any>, options: ExportOptions = {}): RotationMasterExport => {
	const type = host.player.getRotationType();
	const rotation = host.player.getResolvedAplRotation(false);
	const json = APLRotation.toJson(rotation) as Record<string, unknown>;
	return {
		result: exportRotationMaster(json, rotationMasterContext(host), { ...options, anyType: true }),
		generatedFrom: type === APLRotation_Type.TypeAuto ? 'TypeAuto' : type === APLRotation_Type.TypeSimple ? 'TypeSimple' : null,
	};
};
