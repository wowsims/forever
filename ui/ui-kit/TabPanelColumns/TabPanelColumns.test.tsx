import { render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { TabPanelColumns } from './TabPanelColumns';

describe('TabPanelColumns', () => {
	it('renders Root as a custom element via `as`', () => {
		const { container } = render(<TabPanelColumns.Root as="section" />);
		expect(container.firstElementChild!.tagName).toBe('SECTION');
	});

	it('keeps each part addressable by its default testId and takes another', () => {
		const { getByTestId } = render(
			<TabPanelColumns.Root testId="columns">
				<TabPanelColumns.Left>
					<TabPanelColumns.Col />
					<TabPanelColumns.Col testId="second-col" />
				</TabPanelColumns.Left>
				<TabPanelColumns.Right testId="presets" />
			</TabPanelColumns.Root>,
		);

		expect(getByTestId('tab-panel-left').parentElement).toBe(getByTestId('columns'));
		expect(getByTestId('tab-panel-col').parentElement).toBe(getByTestId('tab-panel-left'));
		expect(getByTestId('second-col')).toBeTruthy();
		expect(getByTestId('presets')).toBeTruthy();
	});
});
