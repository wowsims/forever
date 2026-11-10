import { Player as PlayerProto } from '@generated/proto/api';
import { Spec } from '@generated/proto/common';
import { type SimlDiagnostic, SimlParseRequest, SimlPrintRequest } from '@generated/proto/ui';
import type { Sim } from '@sim/sim';

// SIML (sim/siml/SPEC.md), read and written by the sim: in its web worker, or by the server it runs on.

/** A problem in a text, at its line and column counted from 1. */
export const located = (d: SimlDiagnostic) => `${d.startLine}:${d.startCol} ${d.message}`;

/**
 * The `spec` word a file writes for a spec: the field of Player's spec oneof whose message the spec is named after,
 * such as `hunter` for SpecHunter.
 */
export const specField = (spec: Spec) =>
	PlayerProto.fields.find(field => field.oneof === 'spec' && field.kind === 'message' && `Spec${field.T().typeName.split('.').pop()}` === Spec[spec])?.name;

// A result whose request the sim could not take at all, such as one sent before the name table loads, is thrown.
const answered = <T extends { error: string }>(result: T): T => {
	if (result.error) throw new Error(result.error);
	return result;
};

/** `text` read as the file `path` names, whose double extension gives its kind. A use line is refused. */
export const parseSiml = async (sim: Sim, path: string, text: string) => answered(await sim.simlParse(SimlParseRequest.create({ path, text })));

/** Settings printed as a file of `section`, or a full setup with none; text that would not read back gives its problems. */
export const printSiml = async (sim: Sim, request: Partial<SimlPrintRequest>) => answered(await sim.simlPrint(SimlPrintRequest.create(request)));
