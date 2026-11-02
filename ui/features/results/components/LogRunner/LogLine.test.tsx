import { ResourceType } from '@generated/proto/spell';
import i18n from '@i18n/config';
import type { CombatLog, Entity } from '@sim/proto/combat_log';
import { render } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { LogLine } from './LogLine';

// The anchor's own behaviour is `ActionLink`'s test; here it only has to be identifiable in the line.
vi.mock('./ActionLink', () => ({ ActionLink: ({ isAura }: { isAura?: boolean }) => <a data-testid="log-action" data-aura={String(!!isAura)} /> }));

const player = { name: 'Warrior', ownerName: '', index: 0, isTarget: false, isPet: false } as Entity;
const boss = { name: 'Boss', ownerName: '', index: 0, isTarget: true, isPet: false } as Entity;

const base = { logIndex: 0, timestamp: 1, source: player, target: null, actionId: {}, actionIdAsString: null, spellSchool: null, threat: 0, activeAuras: [] };
const line = (extra: Record<string, unknown>) => ({ ...base, raw: '[1.00] raw', ...extra }) as unknown as CombatLog;

const text = (log: CombatLog) => render(<LogLine log={log} />).container.textContent;

// Each line is a locale string with its action link spliced in as an element, which only exists once
// the string resolves; the test setup loads no locale, so this suite loads the English one.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('LogLine', () => {
	it('trails a damage line with its threat, and drops it for a line the target dealt', () => {
		expect(text(line({ kind: 'damage', outcome: 'hit', effect: 'damage', amount: 10, tick: false, target: boss, threat: 25 }))).toContain('(25.00 Threat)');
		expect(text(line({ kind: 'damage', outcome: 'hit', effect: 'damage', amount: 10, tick: false, source: boss, target: boss, threat: 25 }))).not.toContain(
			'Threat',
		);
	});

	// The raw values already moved the right way, so a spend flips the difference back to positive:
	// the line reads "Spent 30.0 Rage", not "Spent -30.0 Rage".
	it('signs a resource line by whether it was spent, and names health losses differently', () => {
		const resource = { kind: 'resource', resourceType: ResourceType.ResourceTypeRage, valueBefore: 100, valueAfter: 70, total: 100 };
		expect(text(line({ ...resource, isSpend: true }))).toContain('Spent');
		expect(text(line({ ...resource, isSpend: true }))).toContain('30.0 Rage');
		expect(text(line({ ...resource, isSpend: false, valueBefore: 70, valueAfter: 100 }))).toContain('Gained');
		expect(text(line({ ...resource, resourceType: ResourceType.ResourceTypeHealth, isSpend: true }))).toContain('Lost');
		expect(text(line({ ...resource, resourceType: ResourceType.ResourceTypeHealth, isSpend: false }))).toContain('Recovered');
	});

	it('says which way an aura went', () => {
		expect(text(line({ kind: 'aura', isGained: true, isFaded: false }))).toContain('gained');
		expect(text(line({ kind: 'aura', isGained: false, isFaded: true }))).toContain('faded');
		expect(text(line({ kind: 'aura', isGained: false, isFaded: false }))).toContain('refreshed');
	});

	// The two aura kinds are the only ones whose tooltip is the buff rather than the spell.
	it('resolves an aura line against the buff aura', () => {
		const auraLink = (log: CombatLog) =>
			render(<LogLine log={log} />)
				.container.querySelector('[data-testid="log-action"]')!
				.getAttribute('data-aura');
		expect(auraLink(line({ kind: 'aura', isGained: true, isFaded: false }))).toBe('true');
		expect(auraLink(line({ kind: 'aura-stacks', oldStacks: 1, newStacks: 2 }))).toBe('true');
		expect(auraLink(line({ kind: 'major-cooldown' }))).toBe('false');
	});

	// The bracket in the raw text is dropped because EntityLabel renders a styled replacement; the
	// second space is the parser's own, kept so the line still matches the plain-text export.
	it('drops the entity bracket the parser left in the raw line, because the label replaces it', () => {
		expect(text(line({ kind: 'plain', raw: '[1.00] [Warrior 1] said something' }))).toBe('[Warrior 1]  said something');
	});

	it('keeps an unparseable raw line whole', () => {
		expect(text(line({ kind: 'plain', raw: 'no timestamp here', source: null }))).toBe('no timestamp here');
	});

	// `stats` is what the sim's Stats.FlatString() prints: one quoted name and amount per stat, each with a trailing comma.
	it.each([
		['a rating with a percent as that percent', '{"CritRating": 14.000,}', '1% Crit'],
		['part of a percent to two decimals', '{"HasteRating": 7.000,}', '0.7% Haste'],
		['defense rating as the skill it buys', '{"DefenseRating": 20.000,}', '20 Defense'],
		['a stat with no percent as its amount', '{"AttackPower": 232.000,}', '232 Attack Power'],
		['a stat the sim holds in percent', '{"PhysicalCritPercent": 3.000,}', '3% Melee Crit'],
		['a negative amount with its own sign', '{"Armor": -1000.000,}', '-1000 Armor'],
		['several stats', '{"Strength": 20.000,"CritRating": 7.000,}', '20 Strength, 0.5% Crit'],
	])('reads %s', (_case, stats, shown) => {
		expect(text(line({ kind: 'stat-change', isGain: true, stats }))).toContain(`Gained ${shown} from`);
		expect(text(line({ kind: 'stat-change', isGain: false, stats }))).toContain(`Lost ${shown} from fading`);
	});

	// Plain lines as sim/core prints them, with the stats from Stats.FlatString() and the parenthesis from StatDependency.String().
	it.each([
		['[1.00] [Warrior 1] Dynamic stat change: {"CritRating": 14.000,}', 'Dynamic stat change: 1% Crit'],
		['[1.00] [Warrior 1] Dynamic stat change: {"CritRating": -14.000,}', 'Dynamic stat change: -1% Crit'],
		['[1.00] [Warrior 1] Dynamic stat change: {"Agility": 20.000,}', 'Dynamic stat change: 20 Agility'],
		['[1.00] [Warrior 1] Pet stats: {"Strength": 20.000,"CritRating": 7.000,}', 'Pet stats: 20 Strength, 0.5% Crit'],
		['[1.00] [Warrior 1] Pet inherited stats: {"HitRating": 10.000,}', 'Pet inherited stats: 1% Hit'],
		['[1.00] [Warrior 1] {"DefenseRating": 20.000,}', '20 Defense'],
		['[1.00] Item Swap - Stats Change: {"HasteRating": -10.000,}', 'Item Swap - Stats Change: -1% Haste'],
	])('reads the stats of %s', (raw, shown) => {
		expect(text(line({ kind: 'plain', raw }))).toContain(shown);
		expect(text(line({ kind: 'plain', raw }))).not.toContain('{');
	});

	it.each([
		['enabled (CritRating += Intellect * 0.50): {"CritRating": 14.000,}', 'enabled (Crit += Intellect * 0.036%): 1% Crit'],
		['disabled (CritRating *= 1.10): {"CritRating": -14.000,}', 'disabled (Crit *= 1.10): -1% Crit'],
		['updated (Armor += DefenseRating * 2.00): {"Armor": 40.000,}', 'updated (Armor += Defense * 2): 40 Armor'],
		['enabled (PhysicalCritPercent += Agility * 0.05): {"Agility": 20.000,}', 'enabled (Melee Crit += Agility * 0.05%): 20 Agility'],
		['enabled (AttackPower += CritRating * 0.50): {"AttackPower": 7.000,}', 'enabled (Attack Power += Crit * 7): 7 Attack Power'],
		['enabled (AttackPower += Strength * 2.00): {"AttackPower": 40.000,}', 'enabled (Attack Power += Strength * 2): 40 Attack Power'],
	])('reads a stat dependency line: %s', (dep, shown) => {
		expect(text(line({ kind: 'plain', raw: `[1.00] [Warrior 1] Dynamic dep ${dep}` }))).toContain(`Dynamic dep ${shown}`);
	});

	it.each(['Pet summoned', 'Dynamic stat change: {}', 'Casting {SpellID: 11605} (x += y * 1.00)'])('leaves a line with no stat text alone: %s', body => {
		expect(text(line({ kind: 'plain', raw: `[1.00] [Warrior 1] ${body}` }))).toBe(`[Warrior 1]  ${body}`);
	});

	// Every kind the union declares has to reach a branch; a missing one renders undefined, silently.
	it('renders something for every kind that reaches the list', () => {
		const kinds: Array<Record<string, unknown>> = [
			{ kind: 'damage', outcome: 'hit', effect: 'damage', amount: 1, tick: false, target: boss },
			{ kind: 'resource', resourceType: ResourceType.ResourceTypeRage, valueBefore: 1, valueAfter: 2, isSpend: false, total: 100 },
			{ kind: 'aura', isGained: true, isFaded: false },
			{ kind: 'aura-stacks', oldStacks: 1, newStacks: 2 },
			{ kind: 'major-cooldown' },
			{ kind: 'cast-began', castTime: 1.5, manaCost: 100 },
			{ kind: 'cast-cancelled', cancelTime: 0.5 },
			{ kind: 'cast-completed', actionId: { name: 'Shield Slam' } },
			{ kind: 'stat-change', isGain: true, stats: '{"Strength": 100.000,}' },
			{ kind: 'resource-group', resourceType: ResourceType.ResourceTypeRage, valueBefore: 1, valueAfter: 2 },
			{ kind: 'cast', actionId: { name: 'Shield Slam' }, castTime: 1 },
			{ kind: 'plain' },
			{ kind: 'dps' },
			{ kind: 'threat-group' },
			{ kind: 'aura-uptime' },
		];
		for (const extra of kinds) expect(text(line(extra))!.length).toBeGreaterThan(0);
	});
});
