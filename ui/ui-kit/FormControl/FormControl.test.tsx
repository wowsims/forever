import { render, screen } from '@testing-library/react';
import { createRef } from 'react';
import { describe, expect, it } from 'vitest';

import { FieldLabel, HelpText, Input, TextArea } from './index';

describe('FieldLabel', () => {
	it('renders a label by default with htmlFor', () => {
		render(<FieldLabel htmlFor="thing">Name</FieldLabel>);
		const label = screen.getByText('Name');
		expect(label.tagName).toBe('LABEL');
		expect(label.getAttribute('for')).toBe('thing');
	});

	it('renders as a span or div without htmlFor', () => {
		render(<FieldLabel as="span">Group</FieldLabel>);
		const span = screen.getByText('Group');
		expect(span.tagName).toBe('SPAN');
		expect(span.hasAttribute('for')).toBe(false);
	});

	it('sets data-testid from testId and forwards other native props', () => {
		render(
			<FieldLabel testId="cooldown-picker-label" title="Field title">
				Name
			</FieldLabel>,
		);
		const label = screen.getByTestId('cooldown-picker-label');
		expect(label.getAttribute('title')).toBe('Field title');
	});
});

describe('HelpText', () => {
	it('renders as a p and forwards hidden', () => {
		render(
			<HelpText as="p" hidden>
				Help
			</HelpText>,
		);
		const el = screen.getByText('Help');
		expect(el.tagName).toBe('P');
		expect(el.hasAttribute('hidden')).toBe(true);
	});
});

describe('TextArea', () => {
	it('forwards ref and attributes', () => {
		const ref = createRef<HTMLTextAreaElement>();
		render(<TextArea ref={ref} spellCheck={false} data-testid="textarea" />);
		const textarea = screen.getByTestId('textarea');
		expect(ref.current).toBe(textarea);
		expect(textarea.getAttribute('spellcheck')).toBe('false');
	});
});

describe('Input', () => {
	it('renders a native input and forwards ref and attributes', () => {
		const ref = createRef<HTMLElement>();
		render(<Input ref={ref} type="text" data-testid="input" />);
		const input = screen.getByTestId('input');
		expect(input.tagName).toBe('INPUT');
		expect(ref.current).toBe(input);
	});
});
