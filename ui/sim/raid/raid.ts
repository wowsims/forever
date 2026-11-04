import { Raid as RaidProto } from '@generated/proto/api';
import { Debuffs, RaidBuffs } from '@generated/proto/buffs';
import { Class, UnitReference, UnitReference_Type as UnitType } from '@generated/proto/common';
import { HealingSimSettings } from '@generated/proto/ui';

import { Player } from '../player/player';
import { Sim } from '../sim';
import { batch } from '../state/batch';
import type { RaidSlice } from '../state/sim_store';
import { shallowArrayEquals } from '../state/subscriptions';
import { sum } from '../utils/math';
import { MAX_PARTY_SIZE, Party } from './party';
export const MAX_NUM_PARTIES = 5;
const MAX_TARGET_DUMMIES = 24;

// Manages all the settings for a single Raid.
export class Raid {
	// Should always hold exactly MAX_NUM_PARTIES elements.
	private parties: Array<Party>;

	// Cached return value for getActivePlayers().
	private activePlayers: Array<Player<any>>;

	readonly sim: Sim;

	constructor(sim: Sim) {
		this.sim = sim;

		this.parties = [...Array(MAX_NUM_PARTIES).keys()].map(i => {
			return new Party(this, sim, i);
		});
		this.activePlayers = [];

		const subscribe = this.sim.store.subscribe;

		// Invalidate the active-players cache synchronously on any composition /
		// party-count write (ungated: readers inside the same batch see fresh
		// data).
		subscribe(
			s => [s.raid.composition, s.raid.numActiveParties],
			() => {
				this.activePlayers = [];
			},
			{ equalityFn: shallowArrayEquals },
		);
	}

	private get raidState() {
		return this.sim.store.getState().raid;
	}

	private patchRaid(patch: Partial<RaidSlice>) {
		this.sim.store.setState(s => ({ raid: { ...s.raid, ...patch } }));
	}

	size(): number {
		return sum(this.parties.map(party => party.size()));
	}

	isEmpty(): boolean {
		return this.size() == 0;
	}

	getParties(): Array<Party> {
		// Make defensive copy.
		return this.parties.slice();
	}

	getParty(index: number): Party {
		return this.parties[index];
	}

	getPlayers(): Array<Player<any> | null> {
		return this.parties.map(party => party.getPlayers()).flat();
	}

	getPlayer(index: number): Player<any> | null {
		if (index === -1) return null;

		const party = this.parties[Math.floor(index / MAX_PARTY_SIZE)];
		return party.getPlayer(index % MAX_PARTY_SIZE);
	}

	getPlayerFromUnitReference(raidTarget: UnitReference | undefined, contextPlayer?: Player<any> | null): Player<any> | null {
		if (!raidTarget || raidTarget.type == UnitType.Unknown) {
			return null;
		} else if (raidTarget.type == UnitType.Player) {
			return this.getPlayer(raidTarget.index);
		} else if (raidTarget.type == UnitType.Self) {
			return contextPlayer || null;
		} else {
			return null;
		}
	}

	setPlayer(index: number, newPlayer: Player<any> | null) {
		const party = this.parties[Math.floor(index / MAX_PARTY_SIZE)];
		party.setPlayer(index % MAX_PARTY_SIZE, newPlayer);
	}

	getClassCount(playerClass: Class) {
		return this.getPlayers().filter(player => player != null && player.getClass() == playerClass).length;
	}

	getBuffs(): RaidBuffs {
		// Make a defensive copy
		return RaidBuffs.clone(this.raidState.buffs);
	}

	setBuffs(newBuffs: RaidBuffs) {
		if (RaidBuffs.equals(this.raidState.buffs, newBuffs)) return;

		// Make a defensive copy
		this.patchRaid({ buffs: RaidBuffs.clone(newBuffs) });
	}

	getDebuffs(): Debuffs {
		// Make a defensive copy
		return Debuffs.clone(this.raidState.debuffs);
	}

	setDebuffs(newDebuffs: Debuffs) {
		if (Debuffs.equals(this.raidState.debuffs, newDebuffs)) return;

		// Make a defensive copy
		this.patchRaid({ debuffs: Debuffs.clone(newDebuffs) });
	}

	getTanks(): Array<UnitReference> {
		// Make a defensive copy
		return this.raidState.tanks.map(tank => UnitReference.clone(tank));
	}

	setTanks(newTanks: Array<UnitReference>) {
		const tanks = this.raidState.tanks;
		if (tanks.length == newTanks.length && tanks.every((tank, i) => UnitReference.equals(tank, newTanks[i]))) return;

		// Make a defensive copy
		this.patchRaid({ tanks: newTanks.map(tank => UnitReference.clone(tank)) });
	}

	getTargetDummies(): number {
		return this.raidState.targetDummies;
	}

	setTargetDummies(newTargetDummies: number) {
		if (this.raidState.targetDummies == newTargetDummies) return;

		this.patchRaid({ targetDummies: newTargetDummies });
	}

	/** The raid slots the target dummies take, in the order the sim places them: the first empty slots of the active parties, 24 at most. */
	getTargetDummySlots(): Array<number> {
		return this.getPlayers()
			.slice(0, this.getNumActiveParties() * MAX_PARTY_SIZE)
			.flatMap((player, index) => (player ? [] : [index]))
			.slice(0, Math.max(0, Math.min(MAX_TARGET_DUMMIES, this.getTargetDummies())));
	}

	getHealingSettings(): HealingSimSettings {
		return HealingSimSettings.clone(this.raidState.healingSettings);
	}

	setHealingSettings(settings: HealingSimSettings) {
		if (HealingSimSettings.equals(this.raidState.healingSettings, settings)) return;
		this.patchRaid({ healingSettings: HealingSimSettings.clone(settings) });
	}

	getNumActiveParties(): number {
		return this.raidState.numActiveParties;
	}
	setNumActiveParties(newNumActiveParties: number) {
		if (newNumActiveParties != this.raidState.numActiveParties && newNumActiveParties > 0) {
			this.patchRaid({ numActiveParties: newNumActiveParties });
		}
	}
	getActivePlayers(): Array<Player<any>> {
		if (this.activePlayers.length == 0) {
			const activeParties = this.getParties().filter((party, i) => i < this.getNumActiveParties());
			this.activePlayers = activeParties
				.map(party => party.getPlayers())
				.flat()
				.filter(player => player != null) as Array<Player<any>>;
		}
		return this.activePlayers;
	}

	toProto(forExport?: boolean, forSimming?: boolean): RaidProto {
		return RaidProto.create({
			parties: this.parties.map(party => party.toProto(forExport, forSimming)),
			buffs: this.getBuffs(),
			debuffs: this.getDebuffs(),
			tanks: this.getTanks(),
			targetDummies: this.getTargetDummies(),
			numActiveParties: this.getNumActiveParties(),
		});
	}

	fromProto(proto: RaidProto) {
		batch(() => {
			this.setBuffs(proto.buffs || RaidBuffs.create());
			this.setDebuffs(proto.debuffs || Debuffs.create());
			this.setTanks(proto.tanks);
			this.setTargetDummies(proto.targetDummies);
			this.setNumActiveParties(proto.numActiveParties || 5);

			for (let i = 0; i < MAX_NUM_PARTIES; i++) {
				if (proto.parties[i]) {
					this.parties[i].fromProto(proto.parties[i]);
				} else {
					this.parties[i].clear();
				}
			}
		});
	}

	clear() {
		batch(() => {
			for (let i = 0; i < MAX_NUM_PARTIES; i++) {
				this.parties[i].clear();
			}
		});
	}
}
