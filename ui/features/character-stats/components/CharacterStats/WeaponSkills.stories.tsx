import { WeaponSkill } from '@generated/proto/common';
import { MAX_WEAPON_SKILL } from '@sim/player/player';
import type { Meta, StoryObj } from '@storybook/react-vite';
import { NumberField } from '@ui-kit/TextField';
import { useState } from 'react';

import { BonusLink } from './BonusLink';
import { type WeaponSkillRow, WeaponSkillsView, type WeaponSkillsViewProps } from './WeaponSkills';

const row = (skill: WeaponSkill, label: string, value: number): WeaponSkillRow => ({ skill, label, value });

const PALADIN_ROWS: Array<WeaponSkillRow> = [
	row(WeaponSkill.WeaponSkillAxes, 'Axes', 300),
	row(WeaponSkill.WeaponSkillTwoHandedAxes, 'Two-Handed Axes', 300),
	row(WeaponSkill.WeaponSkillMaces, 'Maces', 305),
	row(WeaponSkill.WeaponSkillTwoHandedMaces, 'Two-Handed Maces', 300),
	row(WeaponSkill.WeaponSkillSwords, 'Swords', 300),
	row(WeaponSkill.WeaponSkillTwoHandedSwords, 'Two-Handed Swords', 310),
	row(WeaponSkill.WeaponSkillPolearms, 'Polearms', 285),
];

const StatefulView = ({ rows: initialRows }: WeaponSkillsViewProps) => {
	const [rows, setRows] = useState(initialRows);
	const setBonus = (skill: WeaponSkill, bonus: number) =>
		setRows(current => current.map(entry => (entry.skill === skill ? row(skill, entry.label, MAX_WEAPON_SKILL + bonus) : entry)));
	return (
		<WeaponSkillsView
			rows={rows.map(entry => ({
				...entry,
				bonusLink: (
					<BonusLink
						label={`Bonus ${entry.label}`}
						picker={onCommit => (
							<NumberField
								label={`Bonus ${entry.label}`}
								value={entry.value - MAX_WEAPON_SKILL}
								onChange={bonus => {
									setBonus(entry.skill, bonus);
									onCommit();
								}}
							/>
						)}
					/>
				),
			}))}
		/>
	);
};

const meta = {
	title: 'Features/CharacterStats/Components/WeaponSkills',
	component: WeaponSkillsView,
	tags: ['autodocs'],
	args: { rows: PALADIN_ROWS },
	argTypes: { rows: { control: false } },
	render: args => <StatefulView {...args} />,
	decorators: [
		Story => (
			<div className="flex w-62 flex-col">
				<Story />
			</div>
		),
	],
} satisfies Meta<typeof WeaponSkillsView>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {};
