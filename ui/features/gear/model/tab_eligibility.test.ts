import { describe, expect, it } from 'vitest';

import { SelectorModalTabs } from '../types';
import { resolveSelectedTab, TabEligibility } from './tab_eligibility';

const all: TabEligibility = { hasEnchants: true };

describe('resolveSelectedTab', () => {
	it('keeps a tab that has something to show', () => {
		expect(resolveSelectedTab(SelectorModalTabs.Enchants, all)).toBe(SelectorModalTabs.Enchants);
	});

	it('falls back to Items when the requested tab is empty', () => {
		expect(resolveSelectedTab(SelectorModalTabs.Enchants, { ...all, hasEnchants: false })).toBe(SelectorModalTabs.Items);
	});

	it('ignores eligibility that belongs to another tab', () => {
		expect(resolveSelectedTab(SelectorModalTabs.Items, { hasEnchants: false })).toBe(SelectorModalTabs.Items);
		expect(resolveSelectedTab(SelectorModalTabs.RandomSuffixes, { ...all, hasEnchants: false })).toBe(SelectorModalTabs.RandomSuffixes);
	});
});
