import { ResourceType } from '@generated/proto/spell';
import i18n from '@i18n/config';
import { renderWithHost } from '@testing/react';
import { mockSubscriptions } from '@testing/sim/subscriptions';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../../assets/locales/en/translation.json';
import type { AuraItem, CastItem, ResourceItem, RowItem, TickItem } from '../../../model/timeline/rotation';
import { RowItemTooltip } from './RowItemTooltip';

vi.mock('@sim/state/subscriptions', async () => mockSubscriptions());

const text = (item: RowItem) =>
	renderWithHost(<RowItemTooltip item={item} />, {
		sim: { getShowDamageMetrics: () => true, getShowThreatMetrics: () => true, getShowHealingMetrics: () => true },
	}).container.textContent!;

const damage = (amount: number) => ({
	timestamp: 1.5,
	amount,
	threat: 40.25,
	outcome: 'hit',
	effect: null,
	tick: false,
	source: { isTarget: false },
	target: { isTarget: true, index: 0 },
});

const cast = (over: Record<string, unknown> = {}) =>
	({
		kind: 'cast',
		start: 0,
		end: 1,
		log: {
			actionId: { name: 'Fireball' },
			timestamp: 2,
			castTime: 1.5,
			effectiveTime: 1.5,
			travelTime: 0,
			cancelTime: 0,
			castCancelledLog: null,
			damageDealtLogs: [],
			...over,
		},
	}) as unknown as CastItem;

// Every line is a locale string, and a damage line splices its target and amount in as elements that
// only exist once the string resolves; the test setup loads no locale, so this suite loads the English one.
beforeAll(() => {
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

describe('RowItemTooltip', () => {
	it('gives a cast its window, its cast time and its GCD time', () => {
		expect(text(cast())).toContain('Fireball from 2.00s to 3.50s (1.50s, 1.50s GCD Time)');
	});

	it('reports a cancel instead of a GCD, at the point it was cancelled', () => {
		expect(text(cast({ castCancelledLog: { timestamp: 2.75 }, cancelTime: 0.75 }))).toContain('Fireball from 2.00s to 2.75s (Cancelled after 0.75s)');
	});

	it('adds travel time when the spell had any', () => {
		expect(text(cast({ travelTime: 0.4 }))).toContain('+ 0.40s travel time');
	});

	it('totals a cast’s damage and its damage per effective time, and lists the hits with their threat', () => {
		const withDamage = text(cast({ damageDealtLogs: [damage(3000), damage(1500)] }));
		expect(withDamage).toContain('Total: 4500.00 (3000.00 DPET)');
		expect(withDamage).toContain('1.50s - Hit [Target 1] for 3000.00 damage');
		expect(withDamage).toContain('(40.3 Threat)');
	});

	it('lists the damage of every cast in a stack and totals it over their effective time', () => {
		const first = { ...cast().log, timestamp: 1.9, effectiveTime: 1 };
		const last = { ...cast().log, effectiveTime: 1, damageDealtLogs: [damage(3000)] };
		const stacked = text({ ...cast(), log: last, stack: [{ ...first, damageDealtLogs: [damage(1000)] }, last] } as CastItem);
		expect(stacked).toContain('2 casts from 1.90s to 2.00s');
		expect(stacked).toContain('Total: 4000.00 (2000.00 DPET)');
		expect(stacked).toContain('for 1000.00 damage');
		expect(stacked).toContain('for 3000.00 damage');
	});

	it('leaves the total out of a cast that dealt none', () => {
		expect(text(cast())).not.toContain('Total:');
	});

	it('gives a tick its time, its action and its result', () => {
		const tick = { kind: 'tick', start: 4, end: 4, log: { ...damage(900), actionId: { name: 'Ignite' } } } as unknown as TickItem;
		expect(text(tick)).toContain('1.50s - Ignite Hit [Target 1] for 900.00 damage');
	});

	it('gives an aura the window it was up for', () => {
		const aura = { kind: 'aura', start: 1, end: 5, log: { actionId: { name: 'Combustion' }, gainedAt: 1, fadedAt: 5.5 } } as unknown as AuraItem;
		expect(text(aura)).toBe('Combustion: 1.00s - 5.50s');
	});

	it('gives a resource block the value it started from as its maximum', () => {
		const resource = {
			kind: 'resource',
			start: 0,
			end: 1,
			startValue: 200,
			log: {
				resourceType: ResourceType.ResourceTypeMana,
				timestamp: 1,
				valueBefore: 200,
				valueAfter: 100,
				logs: [],
				activeAuras: [{ actionId: { name: 'X', iconUrl: '' } }],
			},
		} as unknown as ResourceItem;
		const rendered = text(resource);
		expect(rendered).toContain('200.0 (100%)');
		// The rotation's own blocks never list auras: the row beneath already shows them.
		expect(rendered).not.toContain('Active Auras');
	});
});
