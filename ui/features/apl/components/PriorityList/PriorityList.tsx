import { AplItemActions, AplList, AplRevealTarget, AplSectionTitle } from '@features/apl/components/AplList';
import { AplListItem } from '@features/apl/components/AplListItem';
import { AplListToolbar } from '@features/apl/components/AplListToolbar';
import { AplNameDialog } from '@features/apl/components/AplNameDialog';
import { AplPasteDialog } from '@features/apl/components/AplPasteDialog';
import { ListItemHeader } from '@features/apl/components/ListItemHeader';
import { uuidValidations } from '@features/apl/components/ListItemHeader/utils';
import { copyAplClip, useAplPaste } from '@features/apl/hooks/useAplClipboard';
import { useVariableExtraction } from '@features/apl/hooks/useVariableExtraction';
import { aplItemId, PRIORITY } from '@features/apl/model/apply_move';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import type { AplRevealRequest } from '@features/apl/model/reveal';
import { rotationSource } from '@features/apl/utils';
import { useListItems } from '@features/inputs/hooks/useListItems';
import { APLListItem } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { Tree } from '@ui-kit/Tree';

const newRow = () => APLListItem.create({ action: {} });
const inPriorityList = (request: AplRevealRequest) => request.kind === 'priority';

/** The rotation's priority list: a section of the priority pane's tree, one item per row. */
export const PriorityList = () => {
	const player = usePlayer();
	const rows = useListItems(player, {
		storeSubscribe: rotationSource,
		getValue: (subject: Player<any>) => subject.aplRotation.priorityList,
		setValue: (subject: Player<any>, next: Array<APLListItem>) =>
			subject.modifyAplRotation(rotation => {
				rotation.priorityList = next;
			}),
		newItem: newRow,
		copyItem: item => freshUuids(APLListItem.clone(item)),
	});
	const extraction = useVariableExtraction(
		player,
		index => player.aplRotation.priorityList[index]?.action?.condition,
		(index, reference) => {
			player.aplRotation.priorityList[index].action!.condition = reference;
		},
	);
	const pasting = useAplPaste('action', rows);
	const header = i18n.t('rotation_tab.apl.priorityList.header');
	const label = i18n.t('rotation_tab.apl.priorityList.name');

	return (
		<AplList.Item
			id={PRIORITY}
			label={header}
			title={<AplSectionTitle text={header} tooltip={i18n.t('rotation_tab.apl.priorityList.tooltips.overview')} />}
			titleAs="h6"
			movable={false}
			revealOn={inPriorityList}
			testId="apl-priority-list-picker-root">
			<Tree.Item.Children>
				{rows.items.length > 0 && (
					<AplList.Rows>
						{rows.items.map((item, index) => {
							const config = rows.itemConfig(item, index);
							const id = aplItemId(item, index);
							return (
								<AplList.Item
									key={id}
									id={id}
									label={label}
									muted={item.hide}
									aside={
										<ListItemHeader
											player={player}
											getItem={config.getValue}
											getValidations={subject => [
												...(subject.getCurrentStats().rotationStats?.priorityList[index]?.validations || []),
												...uuidValidations(subject, config.getValue(subject)?.action?.condition?.uuid?.value),
											]}
										/>
									}
									actions={
										<AplItemActions
											label={label}
											onCopy={() => rows.copy(index)}
											onCopyToClipboard={() => copyAplClip({ kind: 'action', item: config.getValue(player) })}
											onExtract={() => extraction.extract(index)}
											canExtract={() => extraction.canExtract(index)}
											onDelete={() => rows.remove(index)}
										/>
									}>
									<Tree.Item.Body>
										<AplRevealTarget getItem={() => config.getValue(player)}>
											<AplListItem player={player} config={config} />
										</AplRevealTarget>
									</Tree.Item.Body>
								</AplList.Item>
							);
						})}
					</AplList.Rows>
				)}
				<AplListToolbar className="self-stretch" itemName={label} onCreate={() => rows.create()} onPaste={pasting.paste} />
				<AplNameDialog {...extraction.dialog} />
				<AplPasteDialog {...pasting.textDialog} />
			</Tree.Item.Children>
		</AplList.Item>
	);
};
