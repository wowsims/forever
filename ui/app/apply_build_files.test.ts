import { makePresetBuildFromJSON } from '@app/preset_utils';
import { applyBuild } from '@features/settings/model/apply_build';
import { APLRotation } from '@generated/proto/apl';
import { IndividualSimSettings } from '@generated/proto/ui';
import { specTypeFunctions, withSpec } from '@sim/proto/spec_functions';
import { getPlayerSpecFromPlayer } from '@sim/proto/utils';
import { describe, expect, it } from 'vitest';

// Each committed build applied over its spec's defaults by the site's applyBuild, against what siml.LoadBuild gives
// for the same files (sim/siml/testdata/builds, written by TestBuildsOverDefaults). The two must load the same character.
const builds = import.meta.glob<any>('../specs/*/*/builds/*.siml', { eager: true, import: 'default' });
const defaults = import.meta.glob<any>('../specs/*/*/defaults.siml', { eager: true, import: 'default' });
const merged = import.meta.glob<any>('../../sim/siml/testdata/builds/*.json', { eager: true, import: 'default' });

// A host whose setters write into s, the settings the site would export.
const hostOver = (s: IndividualSimSettings, spec: number) => {
	const p = s.player!;
	let options = specTypeFunctions[spec as keyof typeof specTypeFunctions].optionsFromPlayer(p);
	const player = {
		setGear: (e: any) => (p.equipment = e),
		itemSwapSettings: {
			setItemSwapSettings: (enable: boolean, swap: any) => ((p.enableItemSwap = enable), (p.itemSwap = swap)),
			setEnableItemSwap: (enable: boolean) => (p.enableItemSwap = enable),
		},
		setTalentsString: (t: string) => (p.talentsString = t),
		modifyAplRotation: (modify: (r: APLRotation) => void) => {
			p.rotation = APLRotation.clone(p.rotation ?? APLRotation.create());
			modify(p.rotation);
		},
		setAplRotation: (r: APLRotation) => (p.rotation = APLRotation.clone(r)),
		setEpWeights: () => {},
		setRace: (r: any) => (p.race = r),
		setConsumes: (c: any) => (p.consumables = c),
		setProfession1: (x: any) => (p.profession1 = x),
		setProfession2: (x: any) => (p.profession2 = x),
		setDistanceFromTarget: (x: number) => (p.distanceFromTarget = x),
		setReactionTime: (x: number) => (p.reactionTimeMs = x),
		setChannelClipDelay: (x: number) => (p.channelClipDelayMs = x),
		setInFrontOfTarget: (x: boolean) => (p.inFrontOfTarget = x),
		getSpecOptions: () => options,
		setSpecOptions: (o: any) => (options = o),
		setBuffs: (b: any) => (p.buffs = b),
		setHealingModel: (h: any) => (p.healingModel = h),
		getParty: () => ({ setBuffs: (b: any) => (s.partyBuffs = b) }),
	};
	const sim = {
		db: { lookupEquipmentSpec: (e: any) => e, lookupItemSwap: (i: any) => i },
		raid: {
			setBuffs: (b: any) => (s.raidBuffs = b),
			setDebuffs: (d: any) => (s.debuffs = d),
			setTanks: (t: any) => (s.tanks = t),
			setTargetDummies: (n: number) => (s.targetDummies = n),
		},
		encounter: { fromProto: (e: any) => (s.encounter = e) },
	};
	return { host: { player, sim } as any, options: () => options };
};

describe('applyBuild over defaults.siml', () => {
	const cases = Object.entries(builds).map(([path, json]) => {
		const [, cls, folder, file] = path.match(/specs\/([^/]+)\/([^/]+)\/builds\/([^/]+)\.siml$/)!;
		return [
			`${cls}/${folder}/${file}`,
			json,
			defaults[`../specs/${cls}/${folder}/defaults.siml`],
			merged[`../../sim/siml/testdata/builds/${cls}_${folder}_${file}.json`],
		];
	});

	it('covers every build', () => {
		expect(cases.length).toBeGreaterThan(0);
		for (const [name, , , want] of cases) expect(want, `${name} has no testdata/builds file`).toBeDefined();
	});

	it.each(cases)('%s loads as siml.LoadBuild loads it', (_name, buildJson, defaultsJson, wantJson) => {
		const s = IndividualSimSettings.fromJson(defaultsJson);
		const specID = getPlayerSpecFromPlayer(s.player!).specID;
		const { host, options } = hostOver(s, specID);
		applyBuild(makePresetBuildFromJSON('build', specID, buildJson), host);
		s.player = withSpec(specID, s.player!, options());

		expect(IndividualSimSettings.toJson(s)).toEqual(IndividualSimSettings.toJson(IndividualSimSettings.fromJson(wantJson)));
	});
});
