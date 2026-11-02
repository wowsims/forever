import { Class, Spec } from '@generated/proto/common';

import { PlayerSpec } from '../player/player_spec';
import { specTypeFunctions } from '../proto/spec_functions';
import type { SpecTalents } from '../proto/spec_types';
import { brokenTalentNames } from './broken_talents';
import { TalentsConfig } from './config';
import { druidTalentsConfig } from './druid';
import { hunterTalentsConfig } from './hunter';
import { mageTalentsConfig } from './mage';
import { paladinTalentsConfig } from './paladin';
import { priestTalentsConfig } from './priest';
import { rogueTalentsConfig } from './rogue';
import { shamanTalentsConfig } from './shaman';
import { parseTalentsString, serializeTalentsString } from './talents_string';
import { warlockTalentsConfig } from './warlock';
import { warriorTalentsConfig } from './warrior';

export const classTalentsConfig: Record<Class, TalentsConfig<any>> = {
	[Class.ClassUnknown]: [],
	[Class.ClassExtra1]: [],
	[Class.ClassExtra2]: [],
	[Class.ClassExtra3]: [],
	[Class.ClassExtra4]: [],
	[Class.ClassExtra5]: [],
	[Class.ClassExtra6]: [],
	[Class.ClassDruid]: druidTalentsConfig,
	[Class.ClassShaman]: shamanTalentsConfig,
	[Class.ClassHunter]: hunterTalentsConfig,
	[Class.ClassMage]: mageTalentsConfig,
	[Class.ClassRogue]: rogueTalentsConfig,
	[Class.ClassPaladin]: paladinTalentsConfig,
	[Class.ClassPriest]: priestTalentsConfig,
	[Class.ClassWarlock]: warlockTalentsConfig,
	[Class.ClassWarrior]: warriorTalentsConfig,
} as const;

export function playerTalentStringToProto<SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>, talentString: string): SpecTalents<SpecType> {
	const specFunctions = specTypeFunctions[playerSpec.specID];
	const proto = specFunctions.talentsCreate() as SpecTalents<SpecType>;
	const talentsConfig = classTalentsConfig[playerSpec.classID] as TalentsConfig<SpecTalents<SpecType>>;

	return talentStringToProto(proto, talentString, talentsConfig);
}

export function playerBrokenTalentNames<SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>, talentString: string): Array<string> {
	const talentsConfig = classTalentsConfig[playerSpec.classID];
	return brokenTalentNames(talentsConfig, parseTalentsString(talentsConfig, talentString));
}

/**
 * The string as the class's trees read it. A string saved before a tree was reworked can put five points where a talent
 * now takes three, or name a talent the tree no longer has; the sim indexes the talent's ranks by that digit, so it has
 * to be in range before it is sent. A string already within its trees comes back as it is, trailing zeros and all.
 */
export function normalizedTalentsString(classID: Class, talentString: string): string {
	const talentsConfig = classTalentsConfig[classID];
	// A class with no trees (a test player) has nothing to cap against.
	if (!talentsConfig.length) return talentString;
	const inRange = talentString.split('-').every((treeString, treeIdx) => {
		const tree = talentsConfig[treeIdx];
		return (
			tree !== undefined && treeString.length <= tree.talents.length && [...treeString].every((digit, i) => Number(digit) <= tree.talents[i].maxPoints)
		);
	});
	return inRange ? talentString : serializeTalentsString(parseTalentsString(talentsConfig, talentString));
}

export function playerNormalizedTalentsString<SpecType extends Spec>(playerSpec: PlayerSpec<SpecType>, talentString: string): string {
	return normalizedTalentsString(playerSpec.classID, talentString);
}

function talentStringToProto<TalentsProto>(proto: TalentsProto, talentString: string, talentsConfig: TalentsConfig<TalentsProto>): TalentsProto {
	parseTalentsString(talentsConfig, talentString).forEach((treePoints, treeIdx) => {
		talentsConfig[treeIdx].talents.forEach((talentConfig, i) => {
			if (!talentConfig.fieldName) return;
			if (talentConfig.maxPoints == 1) {
				(proto[talentConfig.fieldName as keyof TalentsProto] as unknown as boolean) = treePoints[i] == 1;
			} else {
				(proto[talentConfig.fieldName as keyof TalentsProto] as unknown as number) = treePoints[i];
			}
		});
	});

	return proto;
}
