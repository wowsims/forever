import { Encounter, Target } from '@generated/proto/common';
import { BulkEncounterRef, SavedEncounter } from '@generated/proto/ui';
import { SimHostProvider } from '@sim/context/SimHostContext';
import { bulkState, loadStoredBulkEncounters, patchBulkState } from '@sim/settings/bulk_settings';
import { createSimStore, patchSlice } from '@sim/state/sim_store';
import { makeBulkPlayer } from '@testing/sim/bulk_player';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';

import { Encounters } from './Encounters';

const SAVED_KEY = 'test-saved-encounters';

const CURRENT = { duration: 180, durationVariation: 5, targets: [Target.create()] };

const mount = ({
	presets = [],
	saved = {},
	checked = [],
}: {
	presets?: Array<{ name: string; encounter: Encounter }>;
	saved?: Record<string, Encounter>;
	checked?: BulkEncounterRef[];
} = {}) => {
	const record = Object.entries(saved).map(([name, encounter]) => [name, SavedEncounter.toJson(SavedEncounter.create({ encounter }))]);
	window.localStorage.setItem(SAVED_KEY, JSON.stringify(Object.fromEntries(record)));
	const player = makeBulkPlayer();
	patchSlice(player.sim.store, 'encounter', CURRENT);
	patchBulkState(player, { encounters: checked });
	const host = { player, sim: player.sim, individualConfig: { presets: { encounters: presets } }, getSavedEncounterStorageKey: () => SAVED_KEY } as never;
	render(
		<SimHostProvider host={host}>
			<Encounters />
		</SimHostProvider>,
	);
	return player;
};

const rows = () => screen.queryAllByTestId('bulk-encounter');
const rowNamed = (name: string) => rows().find(row => row.textContent!.startsWith(name))!;

const PATCHWERK = Encounter.create({ duration: 300, durationVariation: 30, targets: [Target.create()] });
const COUNCIL = Encounter.create({ duration: 240, durationVariation: 10, targets: [Target.create(), Target.create(), Target.create()] });
const SAME_AS_CURRENT = Encounter.create({ ...createSimStore().getState().encounter, ...CURRENT });

beforeEach(() => window.localStorage.clear());

describe('Encounters', () => {
	it('lists the presets, then the saved encounters, under their headings', () => {
		mount({ presets: [{ name: 'Patchwerk', encounter: PATCHWERK }], saved: { Council: COUNCIL } });

		expect(rows()).toHaveLength(2);
		expect(rows()[0].textContent).toMatch(/^Patchwerk.*5:00 ±30s · bulk_tab\.settings\.encounters\.targets$/);
		expect(rows()[1].textContent).toMatch(/^Council.*4:00 ±10s · bulk_tab\.settings\.encounters\.targets$/);
		expect(screen.getByText('bulk_tab.settings.encounters.presets')).toBeTruthy();
		expect(screen.getByText('bulk_tab.settings.encounters.saved')).toBeTruthy();
		expect(screen.queryByText('bulk_tab.settings.encounters.empty')).toBeNull();
	});

	it('omits the heading of an empty list', () => {
		mount({ saved: { Council: COUNCIL } });

		expect(screen.queryByText('bulk_tab.settings.encounters.presets')).toBeNull();
		expect(screen.getByText('bulk_tab.settings.encounters.saved')).toBeTruthy();
	});

	it('the locked current row follows the live encounter', () => {
		const player = mount();
		const summary = () => screen.getByTestId('bulk-encounter-current-summary').textContent;
		const current = screen.getByRole('checkbox', { name: 'bulk_tab.settings.encounters.current' });

		expect(current.getAttribute('aria-checked')).toBe('true');
		expect(current.getAttribute('aria-disabled')).toBe('true');
		expect(summary()).toBe('3:00 ±5s · bulk_tab.settings.encounters.targets');

		act(() => patchSlice(player.sim.store, 'encounter', { duration: 200, useHealth: true }));

		expect(summary()).toBe('bulk_tab.settings.encounters.health_summary · bulk_tab.settings.encounters.targets');
	});

	it('stores the checked ref in list order', () => {
		const player = mount({ presets: [{ name: 'Patchwerk', encounter: PATCHWERK }], saved: { Council: COUNCIL } });

		fireEvent.click(screen.getByRole('checkbox', { name: 'Council' }));
		fireEvent.click(screen.getByRole('checkbox', { name: 'Patchwerk' }));

		const expected = [BulkEncounterRef.create({ name: 'Patchwerk', preset: true }), BulkEncounterRef.create({ name: 'Council', preset: false })];
		expect(bulkState(player).encounters).toEqual(expected);
		expect(loadStoredBulkEncounters(player)?.encounters).toEqual(expected);
		expect(screen.getByRole('checkbox', { name: 'Council' }).getAttribute('aria-checked')).toBe('true');
	});

	it('outlines a checked encounter equal to the current one or to an earlier checked one', () => {
		mount({
			presets: [{ name: 'Patchwerk', encounter: PATCHWERK }],
			saved: { 'Same as current': SAME_AS_CURRENT, 'Patchwerk again': PATCHWERK, Council: COUNCIL },
			checked: [
				BulkEncounterRef.create({ name: 'Patchwerk', preset: true }),
				BulkEncounterRef.create({ name: 'Same as current', preset: false }),
				BulkEncounterRef.create({ name: 'Patchwerk again', preset: false }),
			],
		});

		expect(rows().map(row => row.hasAttribute('data-duplicate'))).toEqual([false, true, true, false]);
		expect(rowNamed('Same as current').textContent).toContain('bulk_tab.settings.encounters.duplicate');
		expect(rowNamed('Council').textContent).not.toContain('bulk_tab.settings.encounters.duplicate');

		const describedBy = screen.getByRole('checkbox', { name: 'Same as current' }).getAttribute('aria-describedby')!;
		expect(document.getElementById(describedBy)?.textContent).toBe('bulk_tab.settings.encounters.duplicate');
		expect(screen.getByRole('checkbox', { name: 'Council' }).hasAttribute('aria-describedby')).toBe(false);
		expect(screen.getByRole('checkbox', { name: 'Patchwerk' }).hasAttribute('aria-describedby')).toBe(false);
	});

	it('leaves an unchecked encounter equal to the current one unmarked', () => {
		mount({ saved: { 'Same as current': SAME_AS_CURRENT } });

		expect(rowNamed('Same as current').hasAttribute('data-duplicate')).toBe(false);
	});

	it('points to the saved encounters when there is nothing to pick', () => {
		mount();

		expect(rows()).toHaveLength(0);
		expect(screen.getByText('bulk_tab.settings.encounters.empty')).toBeTruthy();
		expect(screen.getByTestId('bulk-encounter-current')).toBeTruthy();
	});
});
