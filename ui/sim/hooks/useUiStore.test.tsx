import { mountWithSimStore as mount } from '@testing/sim/hooks';
import { act } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { patchSlice } from '../state/sim_store';
import { useUiStore } from './useUiStore';

describe('useUiStore', () => {
	it('renders the field value and re-renders when it changes', () => {
		const renders = { n: 0 };
		const Probe = () => {
			const showEPValues = useUiStore('showEPValues');
			renders.n++;
			return <span>{String(showEPValues)}</span>;
		};

		const { store, container } = mount(<Probe />);
		expect(container.textContent).toBe('false');
		const afterMount = renders.n;

		act(() => patchSlice(store, 'ui', { showEPValues: true }));

		expect(container.textContent).toBe('true');
		expect(renders.n).toBe(afterMount + 1);
	});

	it('does not re-render when another ui field changes', () => {
		const renders = { n: 0 };
		const Probe = () => {
			useUiStore('showEPValues');
			renders.n++;
			return null;
		};

		const { store } = mount(<Probe />);
		const afterMount = renders.n;

		act(() => patchSlice(store, 'ui', { showQuickSwap: false }));

		expect(renders.n).toBe(afterMount);
	});
});
