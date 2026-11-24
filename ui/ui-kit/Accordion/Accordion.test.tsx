import { fireEvent, render } from '@testing-library/react';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';

import { Accordion } from './Accordion';
import { AccordionItem } from './AccordionItem';

const Harness = ({ initial = ['first'] }: { initial?: string[] }) => {
	const [open, setOpen] = useState(initial);
	return (
		<Accordion value={open} onValueChange={setOpen}>
			<AccordionItem value="first" title="First" testId="group-first">
				<span data-testid="first-body" />
			</AccordionItem>
			<AccordionItem value="second" title="Second" testId="group-second">
				<span data-testid="second-body" />
			</AccordionItem>
		</Accordion>
	);
};

describe('Accordion', () => {
	it('keeps the first item open when the second one is opened', () => {
		const { container } = render(<Harness />);

		fireEvent.click(container.querySelector('[data-testid="group-second-trigger"]')!);

		expect(container.querySelector('[data-testid="first-body"]')).not.toBeNull();
		expect(container.querySelector('[data-testid="second-body"]')).not.toBeNull();
	});

	it('marks its root with the default testId', () => {
		const { container } = render(<Harness />);

		expect(container.querySelector('[data-testid="accordion"]')).not.toBeNull();
	});
});
