import { APLRotation } from '@generated/proto/apl';
import { RaidBuffs } from '@generated/proto/buffs';
import { ConsumesSpec, Stat } from '@generated/proto/common';
import i18n from '@i18n/config';
import { ActionId } from '@sim/proto/action_id';
import { renderWithHost } from '@testing/react';
import { filled } from '@testing/sim/hooks';
import { makeIndividualSim, serveRealDatabase } from '@testing/sim/individual_sim';
import { act, fireEvent, screen, waitFor, within } from '@testing-library/react';
import { afterAll, beforeAll, describe, expect, it, vi } from 'vitest';

import english from '../../../../../assets/locales/en/translation.json';
import { ConsumesPicker } from './ConsumesPicker';

vi.mock('@sim/workers/worker_pool', async () => (await import('@testing/sim/worker_pool')).idleWorkerPool());

const ELIXIR_OF_FEROCITY = 250350;
const ELIXIR_OF_THE_MONGOOSE = 13452;
const JUJU_POWER = 12451;
const SCROLL_OF_STRENGTH_IV = 10310;
const SCROLL_OF_INTELLECT_IV = 10308;
const STRENGTH_AND_AGILITY = { autoRotation: () => APLRotation.create(), epStats: [Stat.StatStrength, Stat.StatAgility], presets: { epWeights: [] } };

// The test setup loads no locale, so the tooltips would read as their keys without the names.
beforeAll(() => {
	serveRealDatabase();
	i18n.addResourceBundle('en', 'translation', english, true, true);
});
afterAll(() => {
	i18n.removeResourceBundle('en', 'translation');
});

const WINTERFALL_FIREWATER = 12820;

const renderPicker = async (consumes: ConsumesSpec) => {
	const { player, host } = await makeIndividualSim({ config: STRENGTH_AND_AGILITY });
	player.setConsumes(consumes);
	vi.spyOn(ActionId.prototype, 'fill').mockImplementation(function (this: ActionId) {
		return Promise.resolve(filled(this, String(this.itemId), 'icon.jpg'));
	});
	const { container } = renderWithHost(<ConsumesPicker />, host as unknown as Record<string, unknown>);
	const anchorIn = (scope: ParentNode, itemId: number) => scope.querySelector(`a[href="${ActionId.makeItemUrl(itemId)}"]`);
	const row = () => screen.getByTestId('consumes-elixirs');
	const menu = () => screen.getByTestId('consumes-elixirs-menu');
	const swatch = (scope: () => ParentNode, itemId: number) =>
		waitFor(() => {
			const anchor = anchorIn(scope(), itemId);
			expect(anchor).not.toBeNull();
			return anchor!;
		});
	const overwrittenTooltip = async (scope: () => ParentNode, itemId: number) => {
		const box = (await swatch(scope, itemId)).parentElement!;
		if (!box.hasAttribute('data-overwritten')) return undefined;
		return box.querySelector('[data-testid="icon-swatch-overwritten"]')!.getAttribute('data-tooltip-content');
	};
	const rowPicks = () => [ELIXIR_OF_FEROCITY, ELIXIR_OF_THE_MONGOOSE, JUJU_POWER, WINTERFALL_FIREWATER].filter(id => anchorIn(row(), id));
	const scrolls = () => container.querySelector('[data-testid="consumes-scrolls"]')!;
	return { player, row, menu, swatch, overwrittenTooltip, rowPicks, scrolls };
};

const FEROCITY_TOOLTIP = 'Adds nothing: every stat it grants is matched by Juju Power and Elixir of the Mongoose.';

describe('ConsumesPicker', () => {
	it('shows the picked elixirs alone in the row, and adds one from the menu, which stays open, grouped by stat', async () => {
		const { player, row, menu, swatch, overwrittenTooltip, rowPicks } = await renderPicker(
			ConsumesSpec.create({ elixirIds: [ELIXIR_OF_FEROCITY, ELIXIR_OF_THE_MONGOOSE] }),
		);
		await swatch(row, ELIXIR_OF_FEROCITY);
		expect(new Set(rowPicks())).toEqual(new Set([ELIXIR_OF_FEROCITY, ELIXIR_OF_THE_MONGOOSE]));
		expect(await overwrittenTooltip(row, ELIXIR_OF_FEROCITY)).toBeUndefined();

		fireEvent.click(screen.getByTestId('consumes-elixirs-add'));
		const labels = await waitFor(() => {
			const shown = within(menu())
				.getAllByTestId('consumes-elixir-group')
				.map(group => group.firstElementChild!.textContent);
			expect(shown.length).toBeGreaterThan(1);
			return shown;
		});
		expect(labels[0]).toBe('Critical Strike');
		expect(labels).toContain('Strength');
		expect(labels).toContain('Agility');
		const strength = within(menu())
			.getAllByTestId('consumes-elixir-group')
			.find(group => group.firstElementChild!.textContent === 'Strength')!;
		expect(anchorInGroup(strength, ELIXIR_OF_FEROCITY)).toBe(true);
		expect(anchorInGroup(strength, JUJU_POWER)).toBe(true);

		fireEvent.click(await swatch(menu, JUJU_POWER));
		await waitFor(async () => expect(await overwrittenTooltip(row, ELIXIR_OF_FEROCITY)).toBe(FEROCITY_TOOLTIP));
		expect(await overwrittenTooltip(menu, ELIXIR_OF_FEROCITY)).toBe(FEROCITY_TOOLTIP);
		expect(await overwrittenTooltip(row, ELIXIR_OF_THE_MONGOOSE)).toBeUndefined();
		expect(new Set(rowPicks())).toEqual(new Set([ELIXIR_OF_FEROCITY, ELIXIR_OF_THE_MONGOOSE, JUJU_POWER]));
		expect(player.getConsumes().extraIds).toEqual([JUJU_POWER]);
		expect(screen.queryByTestId('consumes-elixirs-menu')).not.toBeNull();
	});

	it('unpicks an elixir clicked in the row, and takes it out of the row', async () => {
		const { player, row, swatch, rowPicks } = await renderPicker(ConsumesSpec.create({ elixirIds: [ELIXIR_OF_FEROCITY, ELIXIR_OF_THE_MONGOOSE] }));
		fireEvent.click(await swatch(row, ELIXIR_OF_FEROCITY));
		await waitFor(() => expect(player.getConsumes().elixirIds).toEqual([ELIXIR_OF_THE_MONGOOSE]));
		await waitFor(() => expect(rowPicks()).toEqual([ELIXIR_OF_THE_MONGOOSE]));
	});

	it('dims the picked scroll while an elixir picked beside it grants as much of its stat', async () => {
		const { player, overwrittenTooltip, scrolls } = await renderPicker(ConsumesSpec.create({ extraIds: [SCROLL_OF_STRENGTH_IV] }));
		expect(await overwrittenTooltip(scrolls, SCROLL_OF_STRENGTH_IV)).toBeUndefined();

		player.setConsumes(ConsumesSpec.create({ extraIds: [SCROLL_OF_STRENGTH_IV, JUJU_POWER] }));
		await waitFor(async () =>
			expect(await overwrittenTooltip(scrolls, SCROLL_OF_STRENGTH_IV)).toBe('Adds nothing: every stat it grants is matched by Juju Power.'),
		);
	});

	it('dims the picked scroll once a raid buff of its stat that grants as much is turned on', async () => {
		const { player, overwrittenTooltip, scrolls } = await renderPicker(ConsumesSpec.create({ extraIds: [SCROLL_OF_INTELLECT_IV] }));
		act(() => player.getRaid()!.setBuffs(RaidBuffs.create()));
		expect(await overwrittenTooltip(scrolls, SCROLL_OF_INTELLECT_IV)).toBeUndefined();

		act(() => player.getRaid()!.setBuffs(RaidBuffs.create({ arcaneIntellect: true })));
		await waitFor(async () =>
			expect(await overwrittenTooltip(scrolls, SCROLL_OF_INTELLECT_IV)).toBe('Adds nothing: every stat it grants is matched by Arcane Intellect.'),
		);
	});
});

const anchorInGroup = (group: HTMLElement, itemId: number) => !!group.querySelector(`a[href="${ActionId.makeItemUrl(itemId)}"]`);
