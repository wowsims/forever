import type { ReactElement } from 'react';

import { Button } from './Button';

/** A literal `<button>`, or the kit `Button` without an `href`, renders its own native button; `<Button href>` renders an anchor. No render element at all defaults to Base UI's own native button. */
export const isNativeButtonTrigger = (element: ReactElement | undefined): boolean =>
	element === undefined || element.type === 'button' || (element.type === Button && (element.props as { href?: string }).href === undefined);
