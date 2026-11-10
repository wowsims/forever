import { useEffect, useState } from 'react';

// A text box that writes on Enter or when it loses focus, not on every key, so one change is one edit.
export function CommitInput({
	value,
	onCommit,
	placeholder,
	list,
	className,
	type = 'text',
}: {
	value: string;
	onCommit: (value: string) => void;
	placeholder?: string;
	list?: string;
	className?: string;
	type?: string;
}) {
	const [draft, setDraft] = useState(value);
	useEffect(() => setDraft(value), [value]);
	const commit = () => draft !== value && onCommit(draft);
	return (
		<input
			type={type}
			className={className}
			value={draft}
			list={list}
			placeholder={placeholder}
			onChange={e => setDraft(e.target.value)}
			onBlur={commit}
			onKeyDown={e => {
				if (e.key === 'Enter') commit();
				if (e.key === 'Escape') setDraft(value);
			}}
		/>
	);
}
