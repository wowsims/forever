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

	it('keeps a start control apart from the trigger, so pressing it leaves the item closed', () => {
		const { container } = render(
			<Accordion value={[]} onValueChange={() => {}}>
				<AccordionItem value="first" title="First" testId="group-first" start={<button type="button" data-testid="start" />}>
					<span data-testid="first-body" />
				</AccordionItem>
			</Accordion>,
		);

		const start = container.querySelector('[data-testid="start"]')!;
		expect(start.closest('[data-testid="group-first-trigger"]')).toBeNull();
		fireEvent.click(start);
		expect(container.querySelector('[data-testid="first-body"]')).toBeNull();
	});
});
