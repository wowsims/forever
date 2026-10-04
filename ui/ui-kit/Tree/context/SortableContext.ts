import { createContext } from 'react';

import type { SortableController } from '../hooks/useSortable';

/** The drag controller of the topmost sortable root; the roots nested under it share it. */
export const SortableContext = createContext<SortableController | null>(null);
