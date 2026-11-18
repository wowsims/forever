export const hasTouch = () => 'ontouchstart' in window || navigator.maxTouchPoints > 0;

let hoverQuery: MediaQueryList | undefined;

/** Built once: the kit's tooltips ask on every render. Its `matches` follows a mouse being plugged in or out. */
export const hasHover = () => (hoverQuery ??= window.matchMedia('(any-hover: hover)')).matches;
