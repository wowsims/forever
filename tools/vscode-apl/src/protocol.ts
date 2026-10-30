// What the extension and the panel's page send each other, and the shapes of the SIML server's requests they carry
// (sim/siml/lsp/panel.go, patterns.go and sim.go).

import type { Range, TextEdit } from 'vscode-languageserver-types';

// Zero-based, as the protocol counts.
export type { Range, TextEdit };

export type JsonPath = Array<string | number>;
export type JsonValue = null | boolean | number | string | JsonValue[] | JsonObject;
export interface JsonObject {
	[key: string]: JsonValue;
}

// An edit of the rotation's JSON, which siml/splice makes in the text. A path ['whens', id, ...] leads into a when
// block of the structure: its 'condition', its 'hide', or the block itself, which set to no value takes apart.
export type AplEdit =
	| { op: 'set'; path: JsonPath; value?: unknown; call?: string } // No value removes the property or element; `call` writes a def call there.
	| { op: 'insert'; path: JsonPath; index: number; value: unknown } // `path` names the array.
	| { op: 'move'; path: JsonPath; from: number; to: number } // Within one array; `to` is the index the element ends at.
	| { op: 'wrap'; path: JsonPath; from: number; to: number; value: unknown } // Puts rows `from` to `to` in a new block with condition `value`.
	| { op: 'inline'; path: JsonPath } // Writes the def call at `path` as the value it reads as.
	| { op: 'extract'; path: JsonPath; name: string } // Names the value at `path` with a new def.
	| { op: 'defInsert'; name: string; params?: string[]; text: string }
	| { op: 'defDelete'; name: string } // Its calls are written as the values they read as.
	| { op: 'defSet'; name: string; params: string[]; text: string } // Goes alone.
	| { op: 'defRename'; name: string; text: string }; // Renames its calls too; goes alone.

// A when block: the rows `first` to `last` of the list at `list` are in it, nested blocks' included.
export interface WhenBlock {
	id: number;
	list: JsonPath;
	parent: number; // -1 for none.
	first: number;
	last: number; // first - 1 for a block with no rows.
	hide?: boolean;
	condition: JsonValue;
	conjuncts: JsonValue[]; // What it joins to the condition of each row in it, in order.
}

// What siml/rotation says of the text beyond the rotation's JSON (sim/siml/structure.go).
export interface AplStructure {
	whens: WhenBlock[];
	// Each row of a block: how many conjuncts at the start of its condition its blocks join to it, and whether it has
	// conjuncts of its own after them.
	rows: { path: JsonPath; inherited: number; own: boolean }[];
	// The values written as def calls, at their rotation paths or ['whens', id, 'condition', ...].
	calls: { path: JsonPath; name: string; args: string[] }[];
	defs: { name: string; params: string[]; value: string; used: boolean }[];
	whenRanges: Range[];
	defRanges: Range[];
}

// A diagnostic of the file, at the rotation path of the text it is about.
export interface Finding {
	path: JsonPath;
	severity: 'error' | 'warning' | 'info';
	code: string;
	message: string;
}

// What the panel shows for an id: the name and the icon file name from the database.
export type Named = [name: string, icon: string];

// siml/database: the spell and item names and icons of the repository's database.
export interface AplDatabase {
	spells: Record<number, Named>;
	items: Record<number, Named>;
}

// An ActionID as JSON: a spell's id, tag and rank, an item's id, or another action.
export interface ActionIdJson {
	spellId?: number;
	itemId?: number;
	otherId?: string;
	tag?: number;
	rank?: number;
}

// A name of sim/siml/names.json and the ActionID it writes.
export interface RegisteredName {
	name: string;
	id: ActionIdJson;
	icdOnly?: boolean; // Known only as an internal cooldown trigger, which only the ICD kinds find.
	uncastable?: boolean; // A spell no build of the spec lets a rotation cast (the sim's APL flag).
}

// The names a spec's builds register, by the namespace each position reads (SPEC.md §7.1), and the items with a use.
export interface RegisteredSpec {
	spells: RegisteredName[];
	auras: RegisteredName[];
	targetAuras: RegisteredName[];
	petAuras: RegisteredName[];
	targetSpells: RegisteredName[];
	items: RegisteredName[];
}

export type Registered =
	| { status: 'loading' }
	| { status: 'ready'; spec: string; registered: RegisteredSpec } // `spec` is the field of Player's spec oneof.
	| { status: 'unavailable'; reason: string };

// siml/rotation: the rotation an .apl.siml reads as at a document version. `rotation` is left out while the text has
// errors; `spans` are the text of each JSON path; `used` the files its `use` lines read.
export interface RotationResult {
	version: number;
	rotation?: JsonObject;
	readOnly?: string;
	className?: string;
	findings: Finding[];
	spans: { path: JsonPath; range: Range }[];
	used: string[];
	spec?: string; // The field of Player's spec oneof, which siml/registered answers for.
	names: string; // Which name table the file is read with; siml/registered answers again when it changes.
	database: string; // Which database read the server holds; siml/database answers again when it changes.
	structure?: AplStructure; // Left out while the text has errors.
}

// siml/patterns: a common shape Insert Pattern offers, and the values it asks for.
export interface AplPlaceholder {
	name: string;
	label: string;
	kind: 'spell' | 'aura' | 'target_aura' | 'duration' | 'number' | 'group' | 'variable' | 'choice';
	default?: unknown;
	optional?: boolean;
	existing?: boolean;
	options?: string[];
}

export interface AplPattern {
	id: string;
	label: string;
	description: string;
	target: 'row' | 'condition' | 'group' | 'prepull';
	placeholders: AplPlaceholder[];
}

// siml/sim: the result `wowsimcli siml sim` writes, rows keyed by their JSON path joined with '/'.
export interface SimRun {
	dps: { mean: number; stdev: number };
	iterations: number;
	duration: number; // Average iteration length in seconds.
	rows: Record<string, { executions: number; hidden?: boolean }>;
	spells: { id: ActionIdJson; name?: string; dps: number; casts: number; pet?: string }[];
	timeline?: { timeMs: number; path: string; label?: string; via?: string }[];
}

export interface SimResult extends SimRun {
	character: { source: string; race: string; talents: string; gear?: string; warnings?: string[] };
	compare?: SimRun;
	difference?: { delta: number; percent: number }; // Set with compare.
	baseline?: 'saved' | 'HEAD'; // What a compare's base was.
}

// What the extension sends the panel.
export type ToPanel =
	| { type: 'init'; db: AplDatabase; fileName: string; uri: string }
	// The rotation the text reads as, at a document version. `handled` is the last of the panel's edits the extension
	// has made or refused before reading it. With `readOnly` the panel shows why it cannot edit; `rotation` is left out
	// while the text has errors, and the panel keeps showing the last one.
	| {
			type: 'rotation';
			version: number;
			handled: number;
			rotation?: JsonObject;
			readOnly?: string;
			className?: string;
			findings: Finding[];
			structure?: AplStructure;
	  }
	| { type: 'cursor'; path: JsonPath }
	| { type: 'registered'; registered: Registered }
	// The file's last sim; `stale` once the text has changed since.
	| { type: 'sim'; sim?: { result: SimResult; stale: boolean }; running?: boolean };

// What the panel sends the extension. An edit names the document version of the rotation it was made against and its
// place in the panel's sequence of edits; one against a version that someone else's change has moved past is dropped.
export type FromPanel =
	| { type: 'ready' }
	| { type: 'edit'; version: number; seq: number; edits: AplEdit[] }
	| { type: 'reveal'; path: JsonPath }
	| { type: 'revealRange'; range: Range }
	| { type: 'sim'; compare: boolean }
	// Insert a pattern after the row at `path` (an index of -1 is the start of an empty list).
	| { type: 'pattern'; path: JsonPath };

// What the panel keeps through a window reload, so the restored panel finds its file.
export interface PanelState {
	uri: string;
}
