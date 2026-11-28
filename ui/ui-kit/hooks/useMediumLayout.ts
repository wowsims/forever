import { useMedia } from 'react-use';

/** From `--breakpoint-md` up: the width CSS writes as `md:`, and below which `max-md:` applies. Mirrors `breakpoints.css`. */
export const useMediumLayout = (): boolean => useMedia('(min-width: 768px)');
