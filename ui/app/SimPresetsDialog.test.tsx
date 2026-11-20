import i18n from '@i18n/config';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { SimPresetsDialog } from './SimPresetsDialog';

const host = vi.hoisted(() => ({ simDisabled: false }));
const open = vi.hoisted(() => ({ tab: null as string | null }));
vi.mock('@sim/context/SimHostContext', () => ({ useSimHost: () => host }));
vi.mock('./SimTabs', () => ({ useActiveSimTab: () => open.tab }));
vi.mock('./GearPresets', () => ({ GearPresets: () => <p>gear presets</p> }));
vi.mock('./PresetConfigurationPicker', () => ({ PresetConfigurationPicker: () => <p>build presets</p> }));
vi.mock('./tabs/RotationTabBody', () => ({ RotationSidebar: () => <p>rotation presets</p> }));
vi.mock('./tabs/SettingsTabBody', () => ({ SettingsPresets: () => <p>settings presets</p> }));
vi.mock('./tabs/TalentsTabBody', () => ({ TalentPresets: () => <p>talent presets</p> }));

const selected = () => screen.getByRole('tab', { selected: true }).textContent;

describe('SimPresetsDialog', () => {
	it("opens on the open tab's presets", () => {
		open.tab = 'talents-tab';
		render(<SimPresetsDialog open onOpenChange={() => {}} />);

		expect(selected()).toBe(i18n.t('talents_tab.title'));
		expect(screen.getByRole('tabpanel').textContent).toBe('talent presets');
	});

	it('opens on the builds from a tab with no presets column', () => {
		open.tab = 'detailed-results-tab-tab';
		render(<SimPresetsDialog open onOpenChange={() => {}} />);

		expect(selected()).toBe(i18n.t('sidebar.buttons.presets.builds'));
	});

	it('leaves the rotation out of a gear planner', () => {
		host.simDisabled = true;
		render(<SimPresetsDialog open onOpenChange={() => {}} />);

		expect(screen.getAllByRole('tab').map(tab => tab.textContent)).toEqual(
			['sidebar.buttons.presets.builds', 'gear_tab.title', 'settings_tab.title', 'talents_tab.title'].map(key => i18n.t(key)),
		);
		host.simDisabled = false;
	});
});
