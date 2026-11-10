/**
 * @file The tree-sitter grammar of SIML (sim/siml/SPEC.md), for editors that colour and navigate with tree-sitter.
 * @license MIT
 */

/// <reference types="tree-sitter-cli/dsl" />
// @ts-check

// The parser in sim/siml is the language: this grammar reads the shapes an editor needs (sections, rows, entries,
// expressions, when blocks) and leaves every name to the language server. src/scanner.c decides where statements end
// (SPEC.md §3.1) and where when bodies start and end (§6.9), which hang on indentation.

const PREC = {
	or: 1,
	and: 2,
	not: 3,
	compare: 4,
	sum: 5,
	product: 6,
	member: 7,
};

const commaSep1 = rule => seq(rule, repeat(seq(',', rule)));
const commaSep = rule => optional(commaSep1(rule));

export default grammar({
	name: 'siml',

	word: $ => $.identifier,

	extras: $ => [/[ \t\r\n﻿]/, $.comment, $.note],

	externals: $ => [
		$._newline, // The end of a statement.
		$._indent, // The start of a when body, after the when line's colon.
		$._dedent, // The end of a when body.
		$._section_name, // The name of a `name:` header, other than the row sections'.
		$._row_section_name, // `prepull` or `priority` in a header.
		$._group_header, // The `group` of a `group "Name":` header.
		$._error_sentinel, // Never in the grammar: all externals valid means error recovery.
	],

	// A with list goes on past a comma only while a placeholder and = follow it (§4 note 1), which the token after the
	// comma decides.
	conflicts: $ => [[$.group_step]],

	rules: {
		source_file: $ => seq(repeat($._top_statement), repeat(choice($.section, $.row_section, $.group_section))),

		// Before the first header: the top-level lines (§3.3), and a section file's implied section.
		_top_statement: $ =>
			choice(
				$.version,
				$.spec,
				$.talents,
				$.use,
				$.var_declaration,
				$.def_declaration,
				$.when_block,
				seq(alias($._top_row, $.row), $._newline),
				seq($.entry, $._newline),
			),

		version: $ => seq('siml', field('version', $.number), $._newline),
		spec: $ => seq('spec', field('spec', $.identifier), $._newline),
		talents: $ => seq('talents', field('talents', $.string), $._newline),
		use: $ => seq('use', field('path', $.string), optional(seq('with', commaSep1($.variable_binding))), $._newline),
		variable_binding: $ => seq(field('name', $.variable), '=', field('value', $._expression)),

		var_declaration: $ => seq('var', field('name', $.variable), optional(seq('=', field('value', $._expression))), $._newline),
		def_declaration: $ =>
			seq('def', field('name', $.identifier), optional(field('parameters', $.parameters)), '=', field('value', $._expression), $._newline),
		parameters: $ => seq('(', commaSep($.identifier), ')'),

		// Headers (§3.2). Text after the colon is the section's first body line.
		section: $ => seq(field('name', alias($._section_name, $.section_name)), ':', optional($.entry), $._newline, repeat($._section_statement)),
		row_section: $ => seq(field('name', alias($._row_section_name, $.section_name)), ':', optional($.row), $._newline, repeat($._row_statement)),
		group_section: $ => seq(alias($._group_header, 'group'), field('name', $.string), ':', optional($.row), $._newline, repeat($._row_statement)),

		_section_statement: $ => choice(seq($.entry, $._newline), $.var_declaration, $.def_declaration),
		_row_statement: $ => choice(seq($.row, $._newline), $.when_block, $.var_declaration, $.def_declaration),

		// §6.9: the rows indented under the line, each read with the line's condition joined to its own.
		when_block: $ => seq(optional(field('hidden', '#')), 'when', field('condition', $._expression), ':', $._indent, repeat($._row_statement), $._dedent),

		// §6.3. A time goes on a prepull row only; the language server reports it anywhere else.
		row: $ =>
			seq(
				optional(field('hidden', '#')),
				optional(field('time', $._time)),
				field('action', $._step),
				optional(seq('if', optional(field('condition', $._expression)))),
			),
		// A row of an implied section, where a bare kind (`autocast_other_cooldowns`) or a message
		// (`healing_model(hps=1200)`) reads as an entry: a row there starts with an action keyword, `#` or a time, or has
		// an `if`.
		_top_row: $ =>
			choice(
				seq(
					optional(field('hidden', '#')),
					optional(field('time', $._time)),
					field('action', $._keyword_step),
					optional(seq('if', field('condition', $._expression))),
				),
				seq(optional(field('hidden', '#')), optional(field('time', $._time)), field('action', $._plain_step), 'if', field('condition', $._expression)),
				seq(field('hidden', '#'), optional(field('time', $._time)), field('action', $._plain_step)),
				seq(field('time', $._time), field('action', $._plain_step)),
			),
		_time: $ => choice($.number, $.negative_number, $.parenthesized_expression, $.variable, $.placeholder, $.string, $.boolean),

		// §6.4.
		_step: $ => choice($._keyword_step, $._plain_step),
		_plain_step: $ => choice($.identifier, $.call, $.raw, $.none),
		_keyword_step: $ => choice($.cast_step, $.channel_step, $.wait_step, $.move_step, $.group_step, $.schedule_step, $.sequence_step),
		// A step whose keyword has nothing after it yet reads as the step, so that a row being typed ends at its line;
		// the language server reports what is missing.
		cast_step: $ => seq('cast', optional(field('spell', $._ref)), optional(seq('on', field('target', $._unit)))),
		channel_step: $ =>
			seq(
				'channel',
				optional(field('spell', $._ref)),
				optional(seq('on', field('target', $._unit))),
				optional(seq('until', field('interrupt', $._expression))),
				optional('recast'),
			),
		wait_step: $ => seq('wait', optional('until'), optional(field('value', $._expression))),
		move_step: $ =>
			seq('move', optional(choice(seq('for', optional(field('duration', $._expression))), seq(field('range', $._expression), optional('recheck'))))),
		group_step: $ => seq('group', field('name', $.string), optional(seq('with', commaSep1($.placeholder_binding)))),
		placeholder_binding: $ => seq(field('name', $.placeholder), '=', field('value', $._expression)),
		schedule_step: $ => seq('schedule', field('schedule', $.string), field('action', $._step)),
		sequence_step: $ => seq('sequence', field('name', $.string), '[', commaSep($.action), ']'),

		// An action written where a term goes: a sequence's item, or an argument such as strict_sequence's.
		action: $ => seq(field('step', $._step), optional(seq('if', field('condition', $._expression)))),
		_action_term: $ =>
			choice(
				seq(field('step', $._keyword_step), optional(seq('if', field('condition', $._expression)))),
				seq(field('step', $._plain_step), 'if', field('condition', $._expression)),
			),

		_ref: $ => choice($.identifier, $.namespaced_name, $.none, $.raw),
		_unit: $ => choice($.identifier, $.none, $.raw),
		namespaced_name: $ => seq(field('namespace', choice('item', 'other')), '.', field('name', $.identifier)),

		// §6.5.
		_expression: $ => choice($.binary_expression, $.not_expression, $._operand),
		binary_expression: $ => {
			const table = [
				[PREC.or, 'or'],
				[PREC.and, 'and'],
				[PREC.compare, choice('==', '!=', '>=', '>', '<=', '<')],
				[PREC.sum, choice('+', '-')],
				[PREC.product, choice('*', '/')],
			];
			return choice(
				...table.map(([precedence, operator]) =>
					prec.left(
						/** @type {number} */ (precedence),
						seq(field('left', $._expression), field('operator', /** @type {RuleOrLiteral} */ (operator)), field('right', $._expression)),
					),
				),
			);
		},
		not_expression: $ => prec(PREC.not, seq('not', field('value', $._expression))),
		_operand: $ =>
			choice(
				$.parenthesized_expression,
				$.number,
				$.negative_number,
				$.string,
				$.boolean,
				$.none,
				$.variable,
				$.placeholder,
				$.raw,
				$.identifier,
				$.namespaced_name,
				$.call,
				$.member,
			),
		parenthesized_expression: $ => seq('(', $._expression, ')'),
		negative_number: $ => seq('-', $.number),

		// A kind, a def or a message with arguments. `sequence(` heads an accessor (§4 note 4), and `and(`, `or(` and
		// `not(` call the kind with no space before the parenthesis (§4 note 8).
		call: $ =>
			choice(
				seq(field('name', choice($.identifier, alias(choice('sequence', 'schedule'), $.identifier))), field('arguments', $.arguments)),
				seq(field('name', alias(choice('and', 'or', 'not'), $.identifier)), field('arguments', alias($._immediate_arguments, $.arguments))),
			),
		// An accessor (`target.aura(x).is_active`, §6.7) or a unit-prefixed kind (`self.unit_distance`, §6.8): an
		// object, a dot and a property, which takes arguments.
		member: $ =>
			prec.left(
				PREC.member,
				seq(
					field('object', choice($.identifier, $.none, $.call, $.member)),
					'.',
					field('property', $.identifier),
					optional(field('arguments', $.arguments)),
				),
			),
		arguments: $ => seq('(', commaSep($._argument), ')'),
		_immediate_arguments: $ => seq(token.immediate('('), commaSep($._argument), ')'),
		_argument: $ => choice($.named_argument, $._term),
		named_argument: $ => seq(field('name', $.identifier), '=', field('value', $._term)),
		_term: $ => choice($._expression, $.list, alias($._action_term, $.action)),
		list: $ => seq('[', commaSep($._term), ']'),

		// §5.2 and §5.3: a flag list, a field and its values, a message, a slot row, or raw JSON.
		entry: $ =>
			choice(
				seq(field('name', $.identifier), optional(seq('=', field('value', $._value))), repeat(seq(',', $.flag))),
				seq(field('name', $.identifier), $._value, repeat(choice(seq(',', $._value), $._value, $.enchant))),
				field('message', $.call),
				$.raw,
			),
		flag: $ => seq(field('name', $.identifier), optional(seq('=', field('value', $._value)))),
		enchant: $ => seq('+', field('name', $.identifier)),
		_value: $ => choice($.number, $.negative_number, $.string, $.boolean, $.none, $.identifier, $.namespaced_name, $.call, $.raw, $.list),

		// §2.
		identifier: _ => /[A-Za-z_][A-Za-z0-9_]*/,
		number: _ => /[0-9]+(\.[0-9]+)?(%|(h|ms|us|µs|ns|m|s)([0-9]+(\.[0-9]+)?(h|ms|us|µs|ns|m|s))*)?/,
		string: _ => /"([^"\\\n]|\\.)*"/,
		raw: $ => seq('`', optional(alias(token.immediate(/[^`]+/), $.json)), token.immediate('`')),
		variable: _ => /\$([A-Za-z_][A-Za-z0-9_]*|\{([^}\\\n]|\\[}\\n])*\})/,
		placeholder: _ => /%([A-Za-z_][A-Za-z0-9_]*|\{([^}\\\n]|\\[}\\n])*\})/,
		boolean: _ => choice('true', 'false'),
		none: _ => 'none',
		// `///` at the start of a line is a row's note (§10); `//` otherwise is a comment.
		note: _ => token(seq('///', /[^\n]*/)),
		comment: _ => token(seq('//', /([^/\n][^\n]*)?/)),
	},
});
