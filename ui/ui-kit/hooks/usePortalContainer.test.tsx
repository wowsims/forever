import { readdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { Dialog } from '../Dialog/Dialog';
import { Menu } from '../Menu/Menu';
import { MenuItem } from '../Menu/MenuItem';
import { PortalContainerContext } from './usePortalContainer';

describe('usePortalContainer', () => {
	it('portals into the context container when no explicit container is given', () => {
		const host = document.createElement('div');
		document.body.appendChild(host);

		render(
			<PortalContainerContext value={host}>
				<Dialog open onOpenChange={() => {}} title="Options">
					body
				</Dialog>
			</PortalContainerContext>,
		);
		expect(host.contains(screen.getByRole('dialog'))).toBe(true);

		host.remove();
	});

	it('lets an explicit container win over the context', () => {
		const contextHost = document.createElement('div');
		const explicitHost = document.createElement('div');
		document.body.appendChild(contextHost);
		document.body.appendChild(explicitHost);

		render(
			<PortalContainerContext value={contextHost}>
				<Menu trigger="Open" surface="menu" container={explicitHost} open>
					<MenuItem layout="block">Export</MenuItem>
				</Menu>
			</PortalContainerContext>,
		);
		expect(explicitHost.contains(screen.getByRole('menu'))).toBe(true);
		expect(contextHost.contains(screen.getByRole('menu'))).toBe(false);

		contextHost.remove();
		explicitHost.remove();
	});

	it('falls back to document.body with no context and no explicit container', () => {
		render(
			<Dialog open onOpenChange={() => {}} title="Options">
				body
			</Dialog>,
		);
		expect(document.body.contains(screen.getByRole('dialog'))).toBe(true);
	});
});

const here = path.dirname(fileURLToPath(import.meta.url));
const uiKitRoot = path.resolve(here, '..');
const appRoot = path.resolve(here, '../../app');
const featuresRoot = path.resolve(here, '../../features');

const walk = (dir: string): Array<string> =>
	readdirSync(dir, { withFileTypes: true }).flatMap(entry => {
		const full = path.join(dir, entry.name);
		if (entry.isDirectory()) return walk(full);
		return entry.name.endsWith('.tsx') && !entry.name.endsWith('.test.tsx') ? [full] : [];
	});

const relative = (file: string) => path.relative(path.resolve(here, '../../..'), file).replaceAll('\\', '/');

const IMPORTS_HOOK = /usePortalContainer/;

const PORTAL_EXCEPTIONS: Record<string, string> = {};

const OVERLAY_PORTAL_EXCEPTIONS: Record<string, string> = {
	'app/SimTabs.tsx': 'a layout slot for the tab panes, not an overlay',
	'features/results/components/LogRunner/LogRunner.tsx':
		"portals its sticky header into DrStickySlotContext's dr-sticky-slot, a layout slot in DetailedResults.tsx, not an overlay",
	'features/results/components/Timeline/rotation/RotationTrack.tsx':
		"portals its sticky header into DrStickySlotContext's dr-sticky-slot, a layout slot in DetailedResults.tsx, not an overlay",
};

describe('portal root guard', () => {
	it.each([
		['Base UI .Portal', (contents: string) => /\.Portal\b/.test(contents), PORTAL_EXCEPTIONS],
		['createPortal for an overlay', (contents: string) => contents.includes('createPortal('), OVERLAY_PORTAL_EXCEPTIONS],
	])('every %s in ui-kit, app and features goes through usePortalContainer, or is a documented exception', (_, portals, exceptions) => {
		const offenders: Array<string> = [];
		for (const file of [...walk(uiKitRoot), ...walk(appRoot), ...walk(featuresRoot)]) {
			const contents = readFileSync(file, 'utf-8');
			if (!portals(contents)) continue;
			const key = relative(file).replace(/^ui\//, '');
			if (IMPORTS_HOOK.test(contents)) continue;
			if (key in exceptions) continue;
			offenders.push(key);
		}
		expect(offenders).toEqual([]);
	});
});
