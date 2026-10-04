import { useSimHost } from '@sim/context/SimHostContext';
import { screen } from '@testing-library/react';
import { usePortalContainer } from '@ui-kit/hooks/usePortalContainer';
import { describe, expect, it } from 'vitest';

import { renderWithHost } from './react';

const HostName = () => <span data-testid="class-name">{useSimHost().config.className}</span>;
const PortalOwner = () => <span data-testid="portal-owner">{usePortalContainer()?.dataset.testid ?? 'none'}</span>;

describe('renderWithHost', () => {
	it('provides a fake host built from the parts given', () => {
		renderWithHost(<HostName />, { config: { className: 'warrior-sim-ui' } });

		expect(screen.getByTestId('class-name').textContent).toBe('warrior-sim-ui');
	});

	it('provides a portal container inside the rendered tree', () => {
		const { container } = renderWithHost(<PortalOwner />);

		expect(screen.getByTestId('portal-owner').textContent).toBe('test-portal');
		expect(container.contains(screen.getByTestId('test-portal'))).toBe(true);
	});
});
