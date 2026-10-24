import { Player } from '@generated/proto/api';
import { APLRotation } from '@generated/proto/apl';
import { Spec } from '@generated/proto/common';
import { IndividualSimSettings, SimlDiagnostic, SimlParseResult } from '@generated/proto/ui';
import i18n from '@i18n/config';
import { SimSettingCategories } from '@sim/constants/sim_settings';
import { WOWHEAD_GEAR_PLANNER_URL } from '@sim/proto/wowhead';
import { fireEvent, screen, waitFor } from '@testing-library/react';
import type { ComponentType } from 'react';
import { describe, expect, it, vi } from 'vitest';

// The addon importer asks GitHub for its version when its module loads. Answered offline here, in
// `vi.hoisted` so it runs before the imports, no request is left open when the window closes.
vi.hoisted(() => {
	(globalThis as unknown as { fetch: unknown }).fetch = () => Promise.reject(new Error('offline'));
});

import { renderWithHost } from '@testing/react';

const parseSiml = vi.hoisted(() => vi.fn());
vi.mock('../../siml_client', async importOriginal => ({ ...(await importOriginal<typeof import('../../siml_client')>()), parseSiml }));

import {
	ADDON_IMPORTER,
	BULK_GEAR_IMPORTER,
	JSON_IMPORTER,
	SIML_IMPORTER,
	SIML_ROTATION_IMPORTER,
	SIXTY_UPGRADES_IMPORTER,
	WOWHEAD_GEAR_PLANNER_IMPORTER,
} from '../../importers';
import { AddonImporterDialog } from './AddonImporterDialog';
import { BulkGearImporterDialog } from './BulkGearImporterDialog';
import { JsonImporterDialog } from './JsonImporterDialog';
import { SimlImporterDialog, SimlRotationImporterDialog } from './SimlImporterDialog';
import { SixtyUpgradesImporterDialog } from './SixtyUpgradesImporterDialog';
import type { ImporterDialogProps } from './types';
import { WowheadImporterDialog } from './WowheadImporterDialog';

const CASES: Array<[string, ComponentType<ImporterDialogProps>, string, string]> = [
	['AddonImporterDialog', AddonImporterDialog, ADDON_IMPORTER.title, i18n.t('import.addon.instructions')],
	['BulkGearImporterDialog', BulkGearImporterDialog, BULK_GEAR_IMPORTER.title, i18n.t('bulk_tab.import_modal.description_line1')],
	['JsonImporterDialog', JsonImporterDialog, JSON_IMPORTER.title, i18n.t('import.json.instructions')],
	['SimlImporterDialog', SimlImporterDialog, SIML_IMPORTER.title, i18n.t('import.siml.instructions')],
	['SimlRotationImporterDialog', SimlRotationImporterDialog, SIML_ROTATION_IMPORTER.title, i18n.t('import.siml_rotation.instructions')],
	['SixtyUpgradesImporterDialog', SixtyUpgradesImporterDialog, SIXTY_UPGRADES_IMPORTER.title, i18n.t('import.sixty_upgrades.instructions')],
	['WowheadImporterDialog', WowheadImporterDialog, WOWHEAD_GEAR_PLANNER_IMPORTER.title, i18n.t('import.wowhead.instructions')],
];

describe.each(CASES)('%s', (_name, Dialog, title, instructions) => {
	it('opens the importer under its title with its instructions above the textarea', () => {
		renderWithHost(<Dialog open onOpenChange={() => {}} />);
		expect(screen.getByTestId('sim-dialog-title').textContent).toBe(title);
		expect(screen.getByText(instructions)).toBeTruthy();
		expect(screen.getByTestId('importer-textarea')).toBeTruthy();
	});
});

describe('importer dialog links', () => {
	it('points the wowhead importer at the gear planner', () => {
		renderWithHost(<WowheadImporterDialog open onOpenChange={() => {}} />);
		expect(screen.getByRole('link', { name: i18n.t('import.wowhead.gear_planner_link') }).getAttribute('href')).toBe(WOWHEAD_GEAR_PLANNER_URL);
	});

	it('opens the addon link in a new tab without an opener', () => {
		renderWithHost(<AddonImporterDialog open onOpenChange={() => {}} />);
		const link = screen.getByRole('link', { name: i18n.t('import.addon.addon_link') });
		expect(link.getAttribute('target')).toBe('_blank');
		expect(link.getAttribute('rel')).toBe('noopener noreferrer');
	});
});

describe('the SIML rotation importer dialog', () => {
	const importText = (text: string) => {
		const fromProto = vi.fn();
		renderWithHost(<SimlRotationImporterDialog open onOpenChange={() => {}} />, { player: { getSpec: () => Spec.SpecHunter }, fromProto });
		screen.getByTestId<HTMLTextAreaElement>('importer-textarea').value = text;
		fireEvent.click(screen.getByTestId('import-button'));
		return fromProto;
	};

	it('imports the rotation the text holds, and nothing else', async () => {
		const rotation = APLRotation.fromJson({ type: 'TypeAPL', priorityList: [{ action: { castSpell: { spellId: { spellId: 2643 } } } }] });
		parseSiml.mockResolvedValue(
			SimlParseResult.create({
				settings: IndividualSimSettings.create({ player: Player.create({ rotation }) }),
				spec: Spec.SpecHunter,
				categories: ['r'],
			}),
		);
		const fromProto = importText('siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n');
		await waitFor(() => expect(fromProto).toHaveBeenCalledWith(expect.anything(), [SimSettingCategories.Rotation]));
		expect(parseSiml).toHaveBeenLastCalledWith(expect.anything(), 'rotation.apl.siml', 'siml 1\nspec hunter\n\npriority:\n  cast multi_shot\n');
	});

	it('lists a name that does not read at its line and column, under the text', async () => {
		const message = 'no spell is named multi_shto; did you mean multi_shot?';
		parseSiml.mockResolvedValue(
			SimlParseResult.create({
				diagnostics: [SimlDiagnostic.create({ startLine: 5, startCol: 8, endLine: 5, endCol: 18, code: 'unknown-name', message })],
			}),
		);
		const fromProto = importText('siml 1\nspec hunter\n\npriority:\n  cast multi_shto\n');
		const problems = await screen.findByTestId('importer-problems');
		expect(problems.querySelector('li')!.textContent).toBe('5:8 no spell is named multi_shto; did you mean multi_shot?');
		expect(fromProto).not.toHaveBeenCalled();
	});
});
