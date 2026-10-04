import { composeStories, type Meta, setProjectAnnotations, type StoryObj } from '@storybook/react-vite';
import { beforeAll, describe, expect, it } from 'vitest';

import * as preview from '../../../.storybook/preview';

const annotations = setProjectAnnotations([preview]);
beforeAll(annotations.beforeAll);

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
