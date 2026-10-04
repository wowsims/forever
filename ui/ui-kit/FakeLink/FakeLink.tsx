import type { MouseEvent } from 'react';

export interface FakeLinkProps {
	onActivate: (event: MouseEvent<HTMLSpanElement>) => void;
	testId?: string;
}

/**
 * A click layer over its positioned parent: placed last, it takes the clicks on everything the parent
 * holds, while a sibling lifted with `relative z-1` keeps its own. It is hidden
 * from assistive tech and out of the tab order, so the parent's real control stays the one keyboard
 * and screen-reader path.
 */
export const FakeLink = ({ onActivate, testId = 'fake-link' }: FakeLinkProps) => (
	<span aria-hidden="true" className="absolute inset-0 cursor-pointer" data-testid={testId} onClick={onActivate} />
);
