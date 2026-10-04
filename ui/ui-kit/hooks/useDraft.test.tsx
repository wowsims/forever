import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { useDraft } from './useDraft';

const Probe = ({ display }: { display: string }) => {
	const { text, setDraft, clear } = useDraft(display);
	return (
		<div>
			<span data-testid="text">{text}</span>
			<button onClick={() => setDraft('typed')}>type</button>
			<button onClick={clear}>clear</button>
		</div>
	);
};

describe('useDraft', () => {
	it('survives a re-render with the same display', () => {
		const { rerender } = render(<Probe display="500" />);
		fireEvent.click(screen.getByText('type'));
		expect(screen.getByTestId('text').textContent).toBe('typed');

		rerender(<Probe display="500" />);
		expect(screen.getByTestId('text').textContent).toBe('typed');
	});

	it('drops the draft when display changes', () => {
		const { rerender } = render(<Probe display="500" />);
		fireEvent.click(screen.getByText('type'));
		expect(screen.getByTestId('text').textContent).toBe('typed');

		rerender(<Probe display="800" />);
		expect(screen.getByTestId('text').textContent).toBe('800');
	});

	it('drops the draft on clear()', () => {
		render(<Probe display="500" />);
		fireEvent.click(screen.getByText('type'));
		expect(screen.getByTestId('text').textContent).toBe('typed');

		fireEvent.click(screen.getByText('clear'));
		expect(screen.getByTestId('text').textContent).toBe('500');
	});
});
