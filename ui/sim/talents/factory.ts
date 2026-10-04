import { Class, Spec } from '@generated/proto/common';

import { PlayerSpec } from '../player/player_spec';
import { specTypeFunctions } from '../proto/spec_functions';
import type { SpecTalents } from '../proto/spec_types';
import { TalentsConfig } from './config';
import { druidTalentsConfig } from './druid';
import { hunterTalentsConfig } from './hunter';
import { mageTalentsConfig } from './mage';
import { paladinTalentsConfig } from './paladin';
import { priestTalentsConfig } from './priest';
import { rogueTalentsConfig } from './rogue';
import { shamanTalentsConfig } from './shaman';
import { parseTalentsString } from './talents_string';
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
