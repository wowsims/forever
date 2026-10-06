import { PlayerStats } from '@generated/proto/api';
import { Class, WeaponSkill, type WeaponSkillValue } from '@generated/proto/common';
import i18n from '@i18n/config';
import { translateWeaponSkill } from '@i18n/localization';
import { getClassWeaponSkills } from '@sim/player/classes/capabilities';
import { Player } from '@sim/player/player';
import { createSimStore, PLAYER_FIELDS, type PlayerSlice, seedKeyed, zeroVersions } from '@sim/state/sim_store';
import { renderWithHost } from '@testing/react';
import { act, fireEvent, screen } from '@testing-library/react';
import { Accordion } from '@ui-kit/Accordion';
import { describe, expect, it } from 'vitest';

import { WeaponSkills } from './WeaponSkills';

const KEY = 3;

// The Player's own setters run against a seeded slice; its constructor needs a whole sim behind it.
const renderForPaladin = (weaponSkills: Array<WeaponSkillValue> = [], finalSkills?: Array<WeaponSkillValue>) => {
	const store = createSimStore();
	const currentStats = finalSkills ? PlayerStats.create({ weaponSkills: finalSkills }) : undefined;
	seedKeyed(store, 'players', KEY, { weaponSkills, currentStats, v: zeroVersions(PLAYER_FIELDS) } as unknown as PlayerSlice);
	const player = Object.assign(Object.create(Player.prototype), { sim: { store }, storeKey: KEY, playerSpec: { classID: Class.ClassPaladin } });
	renderWithHost(
		<Accordion value={['weapon-skill']} onValueChange={() => {}}>
			<WeaponSkills mode="all" />
		</Accordion>,
		{ player, individualConfig: {} },
	);
	return { stored: () => store.getState().players[KEY].weaponSkills };
};

const shown = (skill: WeaponSkill) => {
	const value = screen.getByText(translateWeaponSkill(skill)).closest('tr')!.querySelector('[data-tone]')!;
	return { value: value.textContent, tone: value.getAttribute('data-tone') };
};

const bonusLabel = (skill: WeaponSkill) => `${i18n.t('sidebar.character_stats.bonus_prefix')} ${translateWeaponSkill(skill)}`;

const openBonus = async (skill: WeaponSkill) => {
	await act(async () => {
		fireEvent.click(screen.getByRole('button', { name: bonusLabel(skill) }));
	});
	return (await screen.findByLabelText(bonusLabel(skill), { selector: 'input' })) as HTMLInputElement;
};

const commit = async (input: HTMLInputElement, text: string) => {
	await act(async () => {
		input.value = text;
		fireEvent.change(input);
	});
};

describe('WeaponSkills', () => {
	it('shows every weapon skill of the class at the level 60 maximum, untoned', () => {
		renderForPaladin();

		for (const skill of getClassWeaponSkills(Class.ClassPaladin)) expect(shown(skill)).toEqual({ value: '300', tone: 'neutral' });
	});

	it("shows the sim's full skill once it has computed the stats, the gear's bonus on top of the stated one", () => {
		renderForPaladin(
			[{ skill: WeaponSkill.WeaponSkillMaces, value: 302 }],
			[
				{ skill: WeaponSkill.WeaponSkillMaces, value: 306 },
				{ skill: WeaponSkill.WeaponSkillSwords, value: 300 },
			],
		);

		expect(shown(WeaponSkill.WeaponSkillMaces)).toEqual({ value: '306', tone: 'positive' });
		expect(shown(WeaponSkill.WeaponSkillSwords)).toEqual({ value: '300', tone: 'neutral' });
	});

	it.each([
		['5', 305, 'positive'],
		['-15', 285, 'negative'],
		['60', 350, 'positive'],
		['-400', 1, 'negative'],
	])('stores a bonus of %s as a skill of %d, between 1 and 350', async (bonus, skill, tone) => {
		const { stored } = renderForPaladin();

		await commit(await openBonus(WeaponSkill.WeaponSkillMaces), bonus);

		expect(stored()).toEqual([{ skill: WeaponSkill.WeaponSkillMaces, value: skill }]);
		expect(shown(WeaponSkill.WeaponSkillMaces)).toEqual({ value: String(skill), tone });
	});

	it('opens on the stored bonus, and drops the skill when its bonus goes back to 0', async () => {
		const { stored } = renderForPaladin([
			{ skill: WeaponSkill.WeaponSkillMaces, value: 305 },
			{ skill: WeaponSkill.WeaponSkillSwords, value: 290 },
		]);

		const input = await openBonus(WeaponSkill.WeaponSkillMaces);
		expect(input.value).toBe('5');
		await commit(input, '0');

		expect(stored()).toEqual([{ skill: WeaponSkill.WeaponSkillSwords, value: 290 }]);
		expect(shown(WeaponSkill.WeaponSkillMaces)).toEqual({ value: '300', tone: 'neutral' });
	});
});
