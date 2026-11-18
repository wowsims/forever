import { useMedia } from 'react-use';

/** From `--breakpoint-lg` up: the width CSS writes as `lg:`, and below which `max-lg:` applies. Mirrors `breakpoints.css`. */
export const useWideLayout = (): boolean => useMedia('(min-width: 1024px)');
