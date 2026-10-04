import { useContext } from 'react';

import { FieldContext } from '../context/FieldContext';

export const useField = () => {
	const value = useContext(FieldContext);
	if (value === null) throw new Error('A Field part must be rendered inside a Field.Root.');
	return value;
};

/** The root's description ids, for a trigger Base UI does not itself carry them onto (a Menu's, unlike a Select's); `undefined` outside a Field.Root and without a description either way. */
export const useFieldDescribedBy = () => useContext(FieldContext)?.describedBy;
