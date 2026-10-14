import i18n from '@i18n/config';
import { WOWHEAD_GEAR_PLANNER_URL } from '@sim/proto/wowhead';
import { screen } from '@testing-library/react';
import type { ComponentType } from 'react';
import { describe, expect, it, vi } from 'vitest';

// The addon importer asks GitHub for its version when its module loads. Answered offline here, in
// `vi.hoisted` so it runs before the imports, no request is left open when the window closes.
vi.hoisted(() => {
	(globalThis as unknown as { fetch: unknown }).fetch = () => Promise.reject(new Error('offline'));
});

import { renderWithHost } from '@testing/react';

import { ADDON_IMPORTER, BULK_GEAR_IMPORTER, JSON_IMPORTER, SIXTY_UPGRADES_IMPORTER, WOWHEAD_GEAR_PLANNER_IMPORTER } from '../../importers';
import { AddonImporterDialog } from './AddonImporterDialog';
import { BulkGearImporterDialog } from './BulkGearImporterDialog';
import { JsonImporterDialog } from './JsonImporterDialog';
import { SixtyUpgradesImporterDialog } from './SixtyUpgradesImporterDialog';
import type { ImporterDialogProps } from './types';
import { WowheadImporterDialog } from './WowheadImporterDialog';

const CASES: Array<[string, ComponentType<ImporterDialogProps>, string, string]> = [
	['AddonImporterDialog', AddonImporterDialog, ADDON_IMPORTER.title, i18n.t('import.addon.instructions')],
	['BulkGearImporterDialog', BulkGearImporterDialog, BULK_GEAR_IMPORTER.title, i18n.t('bulk_tab.import_modal.description_line1')],
	['JsonImporterDialog', JsonImporterDialog, JSON_IMPORTER.title, i18n.t('import.json.instructions')],
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
