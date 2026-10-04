import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { SimTabPane } from './SimTabPane';

describe('SimTabPane', () => {
	it('names the pane by its id and renders its children in the columns container', () => {
		render(
			<SimTabPane id="gear-tab">
				<span data-testid="body" />
			</SimTabPane>,
		);

		expect(screen.getByTestId('sim-tab').id).toBe('gear-tab');
		expect(screen.getByTestId('body').parentElement).toBe(screen.getByTestId('tab-pane-content-container'));
	});

	it('takes its testId and passes other attributes to the pane', () => {
		render(
			<SimTabPane id="rotation-tab" testId="rotation-pane" data-rotation-type="apl">
				<span />
			</SimTabPane>,
		);

		expect(screen.getByTestId('rotation-pane').getAttribute('data-rotation-type')).toBe('apl');
	});
});
