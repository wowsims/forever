import { UIDatabase } from '@generated/proto/ui';
import { composeStories, type Meta, setProjectAnnotations, type StoryObj } from '@storybook/react-vite';
import { beforeAll, describe, expect, it, vi } from 'vitest';

import * as preview from '../../../.storybook/preview';

const annotations = setProjectAnnotations([preview]);
beforeAll(annotations.beforeAll);

// A story that shows an item or spell icon loads the database, which no server answers here, and a
// failed load rejects after the story's test has ended. The database files get an empty database;
// every other request goes through unchanged.
const realFetch = globalThis.fetch;
const emptyDatabase = UIDatabase.toJson(UIDatabase.create());
vi.stubGlobal('fetch', (input: RequestInfo | URL, init?: RequestInit) =>
	String(input).includes('/assets/database/') ? Promise.resolve(Response.json(emptyDatabase)) : realFetch(input, init),
);

type StoriesModule = { default: Meta } & Record<string, StoryObj>;

const modules = import.meta.glob<StoriesModule>('../../**/*.stories.tsx', { eager: true });

describe.each(Object.entries(modules))('%s', (_file, mod) => {
	const stories = Object.entries(composeStories(mod));
	it.each(stories)('%s mounts', async (_name, story) => {
		if (story.parameters?.smoke === false) {
			expect(story.parameters.smokeReason).toMatch(/\S/);
			return;
		}
		const canvasElement = document.body.appendChild(document.createElement('div'));
		try {
			await story.run({ canvasElement });
			expect(canvasElement.childElementCount).toBeGreaterThan(0);
		} finally {
			canvasElement.remove();
		}
	});
});
