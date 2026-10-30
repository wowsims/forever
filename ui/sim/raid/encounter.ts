import {
	AreaType,
	Encounter as EncounterProto,
	EncounterDesign,
	MobType,
	PresetEncounter,
	PresetTarget,
	SpellSchool,
	Stat,
	Target as TargetProto,
	TargetInput,
} from '@generated/proto/common';

import { ENCOUNTER_DEFAULTS } from '../constants/encounter';
import * as Mechanics from '../constants/mechanics';
import { CURRENT_API_VERSION } from '../constants/other';
import { UnitMetadataList } from '../player/player';
import { Stats } from '../proto/stats';
import { Sim } from '../sim';
import { batch } from '../state/batch';
import { EncounterSlice, SimStore } from '../state/sim_store';

// The design points at targets by index. Following each old target object to its new position keeps a reorder
// intact, and a target replaced by a new object at the same index (a field edit, a preset) keeps its place; a
// target that is gone becomes -1 so validation flags it. Returns `design` itself when no index changes.
export const remapDesignTargets = (
	design: EncounterDesign,
	oldTargets: ReadonlyArray<TargetProto>,
	newTargets: ReadonlyArray<TargetProto>,
): EncounterDesign => {
	const before = new Set(oldTargets);
	const after = new Set(newTargets);
	const remap = (index: number) => {
		const target = oldTargets[index];
		if (!target) return -1;
		if (after.has(target)) return newTargets.indexOf(target);
		const replacement = newTargets[index];
		return replacement && !before.has(replacement) ? index : -1;
	};
	const moved = (index: number) => remap(index) !== index;
	const touched =
		design.downtimes.some(d => moved(d.targetIndex)) ||
		design.bossAbilities.some(a => moved(a.targetIndex)) ||
		design.damageAmps.some(a => a.targets?.targetIndices.some(moved));
	// An empty DesignTargets means target 0, a position rather than an identity, so it stays as is.
	if (!touched) return design;

	const next = EncounterDesign.clone(design);
	next.downtimes.forEach(d => (d.targetIndex = remap(d.targetIndex)));
	next.bossAbilities.forEach(a => (a.targetIndex = remap(a.targetIndex)));
	next.damageAmps.forEach(a => {
		if (a.targets) a.targets.targetIndices = a.targets.targetIndices.map(remap);
	});
	return next;
};

// Manages all the settings for an Encounter. State lives in the Zustand
// store's `encounter` slice.
export class Encounter {
	readonly sim: Sim;
	private readonly store: SimStore;

	targetsMetadata: UnitMetadataList;

	constructor(sim: Sim) {
		this.sim = sim;
		this.store = sim.store;
		// Seed the default target (initialization, not a change).
		this.store.setState(s => ({ encounter: { ...s.encounter, targets: [Encounter.defaultTargetProto()] } }));
		this.targetsMetadata = new UnitMetadataList();
	}

	private get enc(): EncounterSlice {
		return this.store.getState().encounter;
	}

	private set(patch: Partial<EncounterSlice>) {
		this.store.setState(s => ({ encounter: { ...s.encounter, ...patch } }));
	}

	// Read-only access. The returned protos are live objects — do NOT mutate
	// them; route all writes through setTargets/modifyTarget so change events fire.
	get primaryTarget(): TargetProto {
		return this.enc.targets[0];
	}

	getTargets(): Array<TargetProto> {
		return this.enc.targets;
	}

	getTarget(index: number): TargetProto | undefined {
		return this.enc.targets[index];
	}

	setTargets(newTargets: Array<TargetProto>) {
		// Always notifies: a same-reference write is stored as a fresh array so
		// subscribers still see a change.
		const targets = newTargets === this.enc.targets ? newTargets.slice() : newTargets;
		this.set({ targets, design: remapDesignTargets(this.enc.design, this.enc.targets, targets) });
	}

	getDesign(): EncounterDesign {
		return this.enc.design;
	}
	setDesign(design: EncounterDesign) {
		if (EncounterDesign.equals(design, this.enc.design)) return;
		this.set({ design: EncounterDesign.clone(design) });
	}
	// Replace-on-write: the callback gets a clone, which becomes the stored design.
	modifyDesign(modify: (design: EncounterDesign) => void) {
		const clone = EncounterDesign.clone(this.enc.design);
		modify(clone);
		this.set({ design: clone });
	}
	hasDesign(): boolean {
		const d = this.enc.design;
		return d.downtimes.length + d.bossHealth.length + d.addWaves.length + d.moves.length + d.damageAmps.length + d.bossAbilities.length > 0;
	}

	// Target mutation with a single change event — the write path for the
	// encounter picker's per-field edits. Replace-on-write: the callback gets a
	// clone, which then replaces the original in a new targets array.
	modifyTarget(index: number, modify: (target: TargetProto) => void) {
		const targets = this.enc.targets.slice();
		if (!targets[index]) {
			// Old behavior for a missing target: the write went to a throwaway
			// object and the change event still fired.
			this.set({ targets });
			return;
		}
		const clone = TargetProto.clone(targets[index]);
		modify(clone);
		targets[index] = clone;
		this.set({ targets });
	}

	getName(): string {
		return this.enc.name;
	}
	setName(name: string) {
		if (name == this.enc.name) return;
		this.set({ name });
	}

	// Whether the encounter is the My encounters entry of its name, which an import of the same name is not.
	isMyEncounter(): boolean {
		return this.enc.myEncounter;
	}
	setMyEncounter(myEncounter: boolean) {
		if (myEncounter == this.enc.myEncounter) return;
		this.set({ myEncounter });
	}

	getDurationVariation(): number {
		return this.enc.durationVariation;
	}
	setDurationVariation(newDuration: number) {
		if (newDuration == this.enc.durationVariation) return;
		this.set({ durationVariation: newDuration });
	}

	getDuration(): number {
		return this.enc.duration;
	}
	setDuration(newDuration: number) {
		if (newDuration == this.enc.duration) return;
		this.set({ duration: newDuration });
	}

	getExecuteProportion20(): number {
		return this.enc.executeProportion20;
	}
	setExecuteProportion20(newExecuteProportion20: number) {
		if (newExecuteProportion20 == this.enc.executeProportion20) return;
		this.set({ executeProportion20: newExecuteProportion20 });
	}
	getExecuteProportion35(): number {
		return this.enc.executeProportion35;
	}
	setExecuteProportion35(newExecuteProportion35: number) {
		if (newExecuteProportion35 == this.enc.executeProportion35) return;
		this.set({ executeProportion35: newExecuteProportion35 });
	}

	getAreaTypes(): Array<AreaType> {
		return this.enc.areaTypes;
	}
	setAreaTypes(newAreaTypes: Array<AreaType>, zoneId = 0) {
		const areaTypes = [...new Set(newAreaTypes)].filter(areaType => areaType != AreaType.AreaTypeUnknown).sort((a, b) => a - b);
		const current = this.enc.areaTypes;
		const sameSet = areaTypes.length == current.length && areaTypes.every((areaType, i) => areaType == current[i]);
		if (sameSet && zoneId == this.enc.zoneId) return;
		this.set({ areaTypes, zoneId });
	}

	getZoneId(): number {
		return this.enc.zoneId;
	}
	// An unknown zone (0 is Custom) keeps the set and clears the zone.
	setZone(zoneId: number) {
		const zone = zoneId ? this.sim.db.getZone(zoneId) : null;
		this.setAreaTypes(zone ? zone.areaTypes : this.enc.areaTypes, zone ? zoneId : 0);
	}

	matchesPreset(preset: PresetEncounter): boolean {
		const targets = this.enc.targets;
		// A preset has no name, and one without a design matches an empty one.
		return (
			!this.enc.name &&
			preset.targets.length == targets.length &&
			targets.every((t, i) => TargetProto.equals(t, preset.targets[i].target)) &&
			EncounterDesign.equals(this.enc.design, preset.design ?? EncounterDesign.create())
		);
	}

	// A preset's target proto is owned by `sim.db` and shared by every encounter naming that preset,
	// so the store gets a clone: what it holds, it owns.
	applyPreset(preset: PresetEncounter) {
		// The preset's design replaces the current one together with the targets.
		this.set({
			name: '',
			myEncounter: false,
			targets: preset.targets.map(presetTarget => (presetTarget.target ? TargetProto.clone(presetTarget.target) : TargetProto.create())),
			design: preset.design ? EncounterDesign.clone(preset.design) : EncounterDesign.create(),
		});

		const zoneId = this.sim.db?.getNpc(preset.targets[0]?.target?.id ?? 0)?.zoneId ?? 0;
		if (zoneId && this.sim.db.getZone(zoneId)?.areaTypes.length) this.setZone(zoneId);
	}

	applyPresetTarget(preset: PresetTarget, index: number) {
		const targets = this.enc.targets.slice();
		targets[index] = preset.target ? TargetProto.clone(preset.target) : TargetProto.create();
		this.set({ targets });
	}

	toProto(): EncounterProto {
		const enc = this.enc;
		return EncounterProto.create({
			name: enc.name,
			myEncounter: enc.myEncounter,
			duration: enc.duration,
			durationVariation: enc.durationVariation,
			executeProportion20: enc.executeProportion20,
			executeProportion35: enc.executeProportion35,
			areaTypes: enc.areaTypes,
			zoneId: enc.zoneId,
			targets: enc.targets,
			design: this.hasDesign() ? enc.design : undefined,
			apiVersion: CURRENT_API_VERSION,
		});
	}

	fromProto(proto: EncounterProto) {
		batch(() => {
			this.setName(proto.name);
			this.setMyEncounter(proto.myEncounter);
			this.setDuration(proto.duration);
			this.setDurationVariation(proto.durationVariation);
			this.setExecuteProportion20(proto.executeProportion20);
			this.setExecuteProportion35(proto.executeProportion35);
			this.setAreaTypes(proto.areaTypes, proto.zoneId);
			// Clone so edits in the Advanced Encounter picker cannot mutate the
			// saved entry's (or the config default's) own target protos.
			this.setTargets(proto.targets.map(t => TargetProto.clone(t)));
			// After the targets, whose reindexing would otherwise rewrite the loaded design.
			this.setDesign(proto.design ? EncounterDesign.clone(proto.design) : EncounterDesign.create());
		});
	}

	applyDefaults() {
		this.fromProto(Encounter.defaultEncounterProto());
	}

	static defaultEncounterProto(numTargets = 1): EncounterProto {
		return EncounterProto.create({
			...ENCOUNTER_DEFAULTS,
			targets: Array.from({ length: numTargets }, () => Encounter.defaultTargetProto()),
			apiVersion: CURRENT_API_VERSION,
		});
	}

	static defaultTargetProto(): TargetProto {
		// Copy default raid target used as fallback for missing DB.
		// https://github.com/wowsims/forever/blob/3570c4fcf1a4e2cd81926019d4a1b3182f613de1/sim/encounters/register_all.go#L24
		return TargetProto.create({
			id: 31146,
			name: 'Raid Target',
			level: Mechanics.BOSS_LEVEL,
			mobType: MobType.MobTypeMechanical,
			stats: Stats.fromMap({
				[Stat.StatArmor]: 3731,
				[Stat.StatAttackPower]: 320,
				[Stat.StatBlockValue]: 54,
				[Stat.StatHealth]: 1428900,
			}).asProtoArray(),
			minBaseDamage: 2601,
			damageSpread: 0.5,
			tankIndex: 0,
			swingSpeed: 2,
			suppressDodge: false,
			parryHaste: true,
			canCrush: true,
			dualWield: false,
			dualWieldPenalty: false,
			spellSchool: SpellSchool.SpellSchoolPhysical,
			targetInputs: new Array<TargetInput>(0),
		});
	}
}
