import { ConnectedNumberField } from '@features/inputs/components/ConnectedTextField';
import type { NumberFieldConfig } from '@features/inputs/configs';
import type { WeaponSkill } from '@generated/proto/common';
import { usePlayer } from '@sim/context/SimHostContext';
import { MAX_WEAPON_SKILL, type Player } from '@sim/player/player';
import { useMemo } from 'react';

// The bonus is on top of the level 60 maximum; the skill may go past it, to 350, for theorycrafting.
const LOWEST_WEAPON_SKILL = 1;
const HIGHEST_WEAPON_SKILL = 350;

export interface WeaponSkillPickerProps {
	skill: WeaponSkill;
	label: string;
	onCommit: () => void;
}

export const WeaponSkillPicker = ({ skill, label, onCommit }: WeaponSkillPickerProps) => {
	const player = usePlayer();

	const config = useMemo(
		(): NumberFieldConfig<Player<any>> => ({
			id: `character-weapon-skill-bonus-${skill}`,
			label,
			inputClassName: 'm-0 w-32 flex-1',
			storeField: 'weaponSkills',
			getValue: subject => subject.getWeaponSkill(skill) - MAX_WEAPON_SKILL,
			setValue: (subject, bonus) => {
				subject.setWeaponSkill(skill, Math.min(Math.max(MAX_WEAPON_SKILL + bonus, LOWEST_WEAPON_SKILL), HIGHEST_WEAPON_SKILL));
				onCommit();
			},
		}),
		[skill, label, onCommit],
	);

	return <ConnectedNumberField modObject={player} config={config} />;
};
