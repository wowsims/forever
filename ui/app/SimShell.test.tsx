import { render } from '@testing-library/react';
import { createRef, type ReactNode } from 'react';
import { describe, expect, it, vi } from 'vitest';

import type { SimHostObject } from './individual_sim_ui';
import { SimShell } from './SimShell';
import type { ShellDom } from './types/shell_dom';

vi.mock('./SimRunCard', () => ({ SimRunCard: () => <section data-testid="sim-run-card" /> }));
vi.mock('./SimTabsSection', () => ({ SimTabsSection: ({ children }: { children: ReactNode }) => <div data-testid="tabs-root">{children}</div> }));
vi.mock('./SimImportExport', () => ({ SimImportExport: () => <div /> }));
vi.mock('@features/character-stats', () => ({ CharacterStats: () => <div data-testid="character-stats" /> }));
vi.mock('@features/results/components/SimResultsPanel', () => ({ SimWarnings: () => <div data-testid="warning-zone" /> }));
vi.mock('@sim/hooks/useSimStatus', () => ({ useSimStatus: () => ({ status: 'ready', error: null }) }));
vi.mock('./header/SimTitleDropdown', () => ({ SimTitleDropdown: () => <div /> }));
vi.mock('./header/SimToolbar', () => ({ SimToolbar: () => <div /> }));
vi.mock('@ui-kit/SocialLink', () => ({ SocialLink: () => <a /> }));
vi.mock('@ui-kit/Toast', () => ({ ToastArea: () => <div data-testid="sim-toast-portal" />, toastManager: {} }));
vi.mock('@sim/hooks/useDisplayMetrics', () => ({ useDisplayMetrics: () => ({}) }));
vi.mock('@sim/hooks/useShowExperimental', () => ({ useShowExperimental: () => false }));

const renderShell = (host: SimHostObject<any> | null) =>
	render(
		<SimShell
			domRef={createRef<ShellDom | null>()}
			host={host}
			sim={{} as never}
			className="warrior"
			spec={{ launch: 0 } as never}
			knownIssues={[]}
			onOpenSettings={() => {}}
		/>,
	);

describe('SimShell sidebar', () => {
	it('stacks the run card, then its warnings, then the character stats', () => {
		const { getByTestId } = renderShell({ warnings: {} } as SimHostObject<any>);

		expect([...getByTestId('sim-sidebar-content').children].map(child => child.getAttribute('data-testid'))).toEqual([
			'sim-run-card',
			'warning-zone',
			'character-stats',
		]);
	});

	it('renders the sidebar inside the tab root, so the run card can open a tab', () => {
		const { getByTestId } = renderShell(null);

		expect(getByTestId('tabs-root').contains(getByTestId('sim-sidebar'))).toBe(true);
		expect(getByTestId('tabs-root').contains(getByTestId('sim-main'))).toBe(true);
	});
});
