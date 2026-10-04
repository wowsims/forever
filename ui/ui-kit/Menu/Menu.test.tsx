import { fireEvent, render } from '@testing-library/react';
import { Button } from '@ui-kit/Button';
import { describe, expect, it } from 'vitest';

import { Menu } from './Menu';
import { MenuItem } from './MenuItem';
import { MenuLinkItem } from './MenuLinkItem';

describe('Menu', () => {
	it('forwards a testid onto the positioner', () => {
		const { getByTestId } = render(
			<Menu trigger="Open" surface="menu" triggerProps={{ 'data-testid': 'trigger' }} positionerProps={{ 'data-testid': 'positioner' }} open>
				<MenuItem layout="row">Item</MenuItem>
			</Menu>,
		);
		fireEvent.click(getByTestId('trigger'));
		expect(getByTestId('positioner')).toBeTruthy();
	});
});

describe('MenuItem', () => {
	it('takes its test id from testId', () => {
		const { getByRole } = render(
			<Menu trigger="Open" surface="menu" open>
				<MenuItem layout="block" testId="export-item">
					Export
				</MenuItem>
			</Menu>,
		);
		expect(getByRole('menuitem', { name: 'Export' }).getAttribute('data-testid')).toBe('export-item');
	});
});

describe('MenuLinkItem', () => {
	it('marks a bare link item', () => {
		const { getByRole } = render(
			<Menu trigger="Open" surface="menu" open>
				<MenuLinkItem layout="block" href="/mage/dps/" bare>
					Mage
				</MenuLinkItem>
			</Menu>,
		);
		expect(getByRole('menuitem', { name: 'Mage' }).getAttribute('data-bare')).toBe('');
	});

	it('renders a menu row that is a link to its href', () => {
		const { getByRole } = render(
			<Menu trigger="Open" surface="menu" open>
				<MenuLinkItem layout="row" href="/warrior/dps/" testId="sim-link">
					Warrior
				</MenuLinkItem>
			</Menu>,
		);

		const link = getByRole('menuitem', { name: 'Warrior' });
		expect(link.tagName).toBe('A');
		expect(link.getAttribute('href')).toBe('/warrior/dps/');
		expect(link.getAttribute('data-testid')).toBe('sim-link');
		expect(link.parentElement?.tagName).toBe('LI');
		expect(link.parentElement?.getAttribute('role')).toBe('none');
	});
});

describe('Menu submenu', () => {
	it('renders the trigger element the caller asked for, with its props', () => {
		const { getByTestId } = render(
			<Menu trigger="Open" surface="menu" triggerProps={{ 'data-testid': 'trigger' }} open>
				<Menu
					submenu
					surface="menu"
					trigger={<span>Sub</span>}
					triggerRender={<button type="button" />}
					triggerProps={{ 'data-testid': 'submenu-trigger' }}>
					<MenuItem layout="row">Nested item</MenuItem>
				</Menu>
			</Menu>,
		);

		const trigger = getByTestId('submenu-trigger');
		expect(trigger.tagName).toBe('BUTTON');
		expect(trigger.getAttribute('role')).toBe('menuitem');
		expect(trigger.getAttribute('aria-haspopup')).toBe('menu');
		expect(trigger.textContent).toBe('Sub');
	});

	it('renders a kit Button submenu trigger without an href as a native button, and with an href as an anchor', () => {
		const { getByTestId, rerender } = render(
			<Menu trigger="Open" surface="menu" triggerProps={{ 'data-testid': 'trigger' }} open>
				<Menu submenu surface="menu" trigger="Sub" triggerRender={<Button />} triggerProps={{ 'data-testid': 'submenu-trigger' }}>
					<MenuItem layout="row">Nested item</MenuItem>
				</Menu>
			</Menu>,
		);
		expect(getByTestId('submenu-trigger').tagName).toBe('BUTTON');

		rerender(
			<Menu trigger="Open" surface="menu" triggerProps={{ 'data-testid': 'trigger' }} open>
				<Menu submenu surface="menu" trigger="Sub" triggerRender={<Button href="https://x" />} triggerProps={{ 'data-testid': 'submenu-trigger' }}>
					<MenuItem layout="row">Nested item</MenuItem>
				</Menu>
			</Menu>,
		);
		expect(getByTestId('submenu-trigger').tagName).toBe('A');
	});
});

describe('Menu root trigger', () => {
	it('gives a kit Button trigger with no href its native role, and one with an href an explicit role="button"', () => {
		const { getByTestId, rerender } = render(
			<Menu trigger="Open" surface="menu" triggerRender={<Button />} triggerProps={{ 'data-testid': 'trigger' }}>
				<MenuItem layout="row">Item</MenuItem>
			</Menu>,
		);
		expect(getByTestId('trigger').tagName).toBe('BUTTON');
		expect(getByTestId('trigger').hasAttribute('role')).toBe(false);

		rerender(
			<Menu trigger="Open" surface="menu" triggerRender={<Button href="https://x" />} triggerProps={{ 'data-testid': 'trigger' }}>
				<MenuItem layout="row">Item</MenuItem>
			</Menu>,
		);
		expect(getByTestId('trigger').tagName).toBe('A');
		expect(getByTestId('trigger').getAttribute('role')).toBe('button');
	});
});
