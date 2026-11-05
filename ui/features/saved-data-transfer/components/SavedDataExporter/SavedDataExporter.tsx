import { Exporter } from '@features/import-export/components/Exporter/Exporter';
import { Spec } from '@generated/proto/common';
import i18n from '@i18n/config';
import type { IndividualSimHost } from '@sim/sim_host';
import type { UseSavedDataResult } from '@ui-kit/hooks/useSavedData';
import { useCallback, useState } from 'react';

import { encodeSavedData, type SavedDataType } from '../../model/transfer_text';
import { SavedDataTickList } from '../SavedDataTickList';

export interface SavedDataExporterProps<T> {
	type: SavedDataType;
	storage: Pick<UseSavedDataResult<T>, 'entries' | 'codec'>;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

const NONE_UNTICKED: ReadonlySet<string> = new Set();

export const SavedDataExporter = <T,>({ type, storage: { entries, codec }, open, onOpenChange }: SavedDataExporterProps<T>) => {
	const [unticked, setUnticked] = useState(NONE_UNTICKED);

	const getData = useCallback(
		(host: IndividualSimHost<any>) => {
			const ticked = entries.filter(({ name }) => !unticked.has(name));
			if (!ticked.length) return '';
			return encodeSavedData({
				type,
				spec: Spec[host.player.getSpec()],
				entries: Object.fromEntries(ticked.map(({ name, data }) => [name, codec.toJson(data)])),
			});
		},
		[type, entries, codec, unticked],
	);

	const setTicked = (name: string, ticked: boolean) =>
		setUnticked(current => {
			const next = new Set(current);
			if (ticked) next.delete(name);
			else next.add(name);
			return next;
		});

	return (
		<Exporter
			open={open}
			onOpenChange={next => {
				if (!next) setUnticked(NONE_UNTICKED);
				onOpenChange(next);
			}}
			title={i18n.t('saved_data_transfer.export_title', { type: i18n.t(`saved_data_transfer.types.${type}`) })}
			allowDownload
			downloadFileName={`wowsims-saved-${type}.txt`}
			downloadMimeType="text/plain"
			holdsLink
			getData={getData}>
			<SavedDataTickList
				rows={entries.map(({ name }) => ({ name, checked: !unticked.has(name) }))}
				onChange={(index, checked) => setTicked(entries[index].name, checked)}
			/>
		</Exporter>
	);
};
