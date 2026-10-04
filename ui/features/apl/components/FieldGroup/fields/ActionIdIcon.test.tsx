import { OtherAction } from '@generated/proto/common';
import { ActionId } from '@sim/proto/action_id';
import { act, render } from '@testing-library/react';
import { describe, expect, it } from 'vitest';

import { ActionIdIcon } from './ActionIdIcon';

const mount = async (actionId: ActionId) => {
	const rendered = render(<ActionIdIcon actionId={actionId} useBuffAura={false} />);
	await act(async () => {});
	return rendered.container.querySelector('a') as HTMLAnchorElement;
};

describe('ActionIdIcon', () => {
	it('paints the icon of an id that already carries one, without linking anywhere', async () => {
		const anchor = await mount(ActionId.fromOtherId(OtherAction.OtherActionAttack));

		expect(anchor.style.backgroundImage).toContain('inv_sword_04.jpg');
		expect(anchor.hasAttribute('href')).toBe(false);
		expect(anchor.hasAttribute('data-whtticon')).toBe(false);
	});

	it('paints nothing for an empty id', async () => {
		const anchor = await mount(ActionId.fromEmpty());

		expect(anchor.style.backgroundImage).toBe('');
		expect(anchor.hasAttribute('href')).toBe(false);
	});
});
