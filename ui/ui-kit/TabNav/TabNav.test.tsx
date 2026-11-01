import { Tabs } from '@base-ui/react/tabs';
import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TabNav } from './TabNav';

describe('TabNav', () => {
	it('gives each tab of the sim variant its id as its testid', () => {
		const { getByTestId } = render(
			<Tabs.Root defaultValue="a">
				<TabNav variant="sim" testId="sim-tabs" tabs={[{ id: 'a', label: 'A' }]} />
			</Tabs.Root>,
		);
		expect(getByTestId('a').textContent).toBe('A');
		expect(getByTestId('sim-tabs').contains(getByTestId('a'))).toBe(true);
	});

	it('places a tab through its renderTab, still as a tab', () => {
		const { getByTestId } = render(
			<Tabs.Root defaultValue="a">
				<TabNav variant="sim" tabs={[{ id: 'a', label: 'A', renderTab: tab => <span data-testid="wrapper">{tab}</span> }]} />
			</Tabs.Root>,
		);
		expect(getByTestId('wrapper').contains(getByTestId('a'))).toBe(true);
		expect(getByTestId('a').getAttribute('role')).toBe('tab');
	});

	it('carries per-tab id, aria-controls and data-label', () => {
		const { getByText } = render(
			<Tabs.Root defaultValue="a">
				<TabNav tabs={[{ id: 'a', label: 'A', tabId: 'a-nav', ariaControls: 'a-pane', dataLabel: 'A' }]} />
			</Tabs.Root>,
		);
		const tab = getByText('A');
		expect(tab.getAttribute('id')).toBe('a-nav');
		expect(tab.getAttribute('aria-controls')).toBe('a-pane');
		expect(tab.getAttribute('data-label')).toBe('A');
	});
});
