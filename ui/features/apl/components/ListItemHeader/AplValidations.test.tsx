import { APLValidation } from '@generated/proto/api';
import { LogLevel } from '@generated/proto/common';
import { act, fireEvent, render } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';

import { AplValidationsView } from './AplValidations';

vi.mock('@i18n/config', () => ({ default: { t: (key: string) => key } }));

const validation = (logLevel: LogLevel, text: string) => APLValidation.create({ logLevel, validation: text });

const button = (container: HTMLElement) => container.querySelector('[data-testid="apl-validations"]') as HTMLElement;

const openTooltip = async (container: HTMLElement) => {
	await act(async () => {
		fireEvent.mouseEnter(button(container));
	});
	return document.querySelector('.sim-tooltip') as HTMLElement;
};

describe('AplValidationsView', () => {
	it('keeps the button in place but hidden while there is nothing to say', () => {
		const { container } = render(<AplValidationsView validations={[]} />);

		expect(button(container).style.display).toBe('none');
	});

	it.each([
		['information alone', [validation(LogLevel.Information, 'fyi')], 'info'],
		['a warning above information', [validation(LogLevel.Information, 'fyi'), validation(LogLevel.Warning, 'careful')], 'warning'],
		[
			'an error above the rest',
			[validation(LogLevel.Information, 'fyi'), validation(LogLevel.Error, 'broken'), validation(LogLevel.Warning, 'careful')],
			'error',
		],
	])('shows itself at the highest level present, for %s', (_case, validations, level) => {
		const { container } = render(<AplValidationsView validations={validations} />);

		expect(button(container).style.display).not.toBe('none');
		expect(button(container).getAttribute('data-level')).toBe(level);
	});

	it('groups the messages under one header per level, in the order the levels first appear', async () => {
		const { container } = render(
			<AplValidationsView
				validations={[validation(LogLevel.Warning, 'careful'), validation(LogLevel.Error, 'broken'), validation(LogLevel.Warning, 'again')]}
			/>,
		);

		const tooltip = await openTooltip(container);
		expect([...tooltip.querySelectorAll('p')].map(header => header.textContent)).toEqual([
			'common.list_picker.action_has_warnings',
			'common.list_picker.action_has_errors',
		]);
		expect([...tooltip.querySelectorAll('ul')].map(list => [...list.querySelectorAll('li')].map(item => item.textContent))).toEqual([
			['careful', 'again'],
			['broken'],
		]);
	});

	it('renders a message as text, so markup inside it cannot execute', async () => {
		const hostile = `Group reference '<img src=x onerror=alert(1)>' not found`;
		const { container } = render(<AplValidationsView validations={[validation(LogLevel.Error, hostile)]} />);

		const tooltip = await openTooltip(container);
		expect(tooltip.querySelector('img')).toBeNull();
		expect(tooltip.querySelector('li')!.textContent).toBe(hostile);
	});
});
