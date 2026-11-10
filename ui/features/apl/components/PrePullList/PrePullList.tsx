import { AplItemActions, AplList, AplRevealTarget, AplSectionTitle } from '@features/apl/components/AplList';
import { AplListItem } from '@features/apl/components/AplListItem';
import { ListItemHeader } from '@features/apl/components/ListItemHeader';
import { ValuePicker } from '@features/apl/components/ValuePicker';
import { AplProvider } from '@features/apl/context/AplContext';
import { aplItemId, PREPULL } from '@features/apl/model/apply_move';
import { freshUuids } from '@features/apl/model/fresh_uuids';
import type { AplRevealRequest } from '@features/apl/model/reveal';
import { rotationSource } from '@features/apl/utils';
import { useListItems } from '@features/inputs/hooks/useListItems';
import { APLPrepullAction, APLValue } from '@generated/proto/apl';
import i18n from '@i18n/config';
import { usePlayer } from '@sim/context/SimHostContext';
import type { Player } from '@sim/player/player';
import { randomUUID } from '@sim/utils/misc';
import { Tree } from '@ui-kit/Tree';

import { doAtConfig } from './do_at';

const AT_PULL = () => APLValue.create({ value: { oneofKind: 'const', const: { val: '-1s' } }, uuid: { value: randomUUID() } });
const inPrepullList = (request: AplRevealRequest) => request.kind === 'prepull';

/** The actions taken before the pull, each with the time it happens at: a section of the priority pane's tree. */
export const PrePullList = () => {
	const player = usePlayer();
	const rows = useListItems(player, {
		storeSubscribe: rotationSource,
		getValue: (subject: Player<any>) => subject.aplRotation.prepullActions,
		setValue: (subject: Player<any>, next: Array<APLPrepullAction>) =>
			subject.modifyAplRotation(rotation => {
				rotation.prepullActions = next;
			}),
		newItem: () => APLPrepullAction.create({ action: {}, doAtValue: { value: { oneofKind: 'const', const: { val: '-1s' } } } }),
		copyItem: item => freshUuids(APLPrepullAction.clone(item)),
	});
	const header = i18n.t('rotation_tab.apl.prePullActions.header');
	const label = i18n.t('rotation_tab.apl.prePullActions.name');

	return (
		<AplProvider isPrepull>
			<AplList.Item
				id={PREPULL}
				label={header}
				title={<AplSectionTitle text={header} tooltip={i18n.t('rotation_tab.apl.prePullActions.tooltips.overview')} />}
				titleAs="h6"
				movable={false}
				revealOn={inPrepullList}
				testId="apl-prepull-list-root">
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
												getValidations={subject => subject.getCurrentStats().rotationStats?.prepullActions[index]?.validations || []}
											/>
										}
										actions={<AplItemActions label={label} onCopy={() => rows.copy(index)} onDelete={() => rows.remove(index)} />}>
										<Tree.Item.Body>
											<AplRevealTarget getItem={() => config.getValue(player)}>
												<AplListItem
													player={player}
													config={config}
													leading={
														<ValuePicker
															player={player}
															config={doAtConfig(
																() => config.getValue(player)?.doAtValue,
																(subject: Player<any>, newValue: APLValue | undefined) => {
																	const current = config.getValue(subject);
																	if (!current) return;
																	current.doAtValue = newValue || AT_PULL();
																	subject.touchRotation();
																},
															)}
														/>
													}
												/>
											</AplRevealTarget>
										</Tree.Item.Body>
									</AplList.Item>
								);
							})}
						</AplList.Rows>
					)}
					<Tree.AddButton
						className="mb-2"
						label={i18n.t('rotation_tab.apl.floatingActionBar.new', { itemName: label })}
						onClick={() => rows.create()}
					/>
				</Tree.Item.Children>
			</AplList.Item>
		</AplProvider>
	);
};
