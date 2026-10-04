import { Stat } from '@generated/proto/common';
import { Stats } from '@sim/proto/stats';
import { describe, expect, it } from 'vitest';

import { SIXTY_UPGRADES_EP_EXPORTER } from './sixty_upgrades_ep';

const exportFor = (weights: Stats) =>
	SIXTY_UPGRADES_EP_EXPORTER.getData({ player: { getEpWeights: () => weights, getPlayerSpec: () => ({ friendlyName: 'Arms' }) } } as any, {} as any);

describe('SIXTY_UPGRADES_EP_EXPORTER', () => {
	it('links to the Forever EP import', () => {
		expect(exportFor(new Stats().withStat(Stat.StatStrength, 1))).toMatch(
			/^https:\/\/sixtyupgrades\.com\/forever\/ep\/import\?name=Arms%20WoWSims%20Weights&strength=1\.000$/,
		);
	});

	it('scales rating weights to per-percent weights', () => {
		const url = exportFor(
			new Stats()
				.withStat(Stat.StatHitRating, 2.5)
				.withStat(Stat.StatCritRating, 1)
				.withStat(Stat.StatHasteRating, 0.4)
				.withStat(Stat.StatBlockRating, 0.2),
		);
		expect(url).toContain('&hit=25.000');
		expect(url).toContain('&crit=14.000');
		expect(url).toContain('&haste=4.000');
		expect(url).toContain('&block=1.000');
	});

	it('scales the avoidance and expertise ratings, and sends defense per point', () => {
		const url = exportFor(
			new Stats()
				.withStat(Stat.StatExpertiseRating, 0.5)
				.withStat(Stat.StatDodgeRating, 0.5)
				.withStat(Stat.StatParryRating, 0.2)
				.withStat(Stat.StatDefenseRating, 0.3),
		);
		expect(url).toContain('&expertise=5.000');
		expect(url).toContain('&dodge=6.000');
		expect(url).toContain('&parry=3.000');
		expect(url).toContain('&defense=0.300');
	});

	it('sends the one hit and crit rating as its hit and crit', () => {
		const url = exportFor(new Stats().withStat(Stat.StatHitRating, 1.3).withStat(Stat.StatCritRating, 0.5).withStat(Stat.StatSpellDamage, 0.5));
		expect(url).toContain('&hit=13.000');
		expect(url).toContain('&crit=7.000');
		expect(url).toContain('&spellDamage=0.500');
		expect(url).not.toContain('spellHit');
		expect(url).not.toContain('spellCrit');
		expect(url).not.toContain('rangedHit');
	});

	it('names healing, spell penetration and resistances the way the import reads them', () => {
		const url = exportFor(new Stats().withStat(Stat.StatHealingPower, 1).withStat(Stat.StatSpellPiercing, 1).withStat(Stat.StatFireResistance, 1));
		expect(url).toContain('&healing=1.000');
		expect(url).toContain('&spellPen=1.000');
		expect(url).toContain('&fireResist=1.000');
	});
});
