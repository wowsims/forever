// Where SIML statements end (sim/siml/SPEC.md §3.1), and where when bodies start and end (§6.9).
//
// A statement continues while a bracket is open, which the grammar sees for itself: no newline token is valid inside
// one, so the line break is whitespace. Otherwise it continues onto the next line that is not blank or comment-only
// when that line is indented deeper than the statement's first, no blank line comes between, the statement is not a
// header and the line does not start one. A when body is the statements after the when line whose first line is
// indented deeper than it, up to a header or the end of the file. Headers are found here too, so that a section ends
// at the next header without the grammar looking ahead.

#include "tree_sitter/alloc.h"
#include "tree_sitter/parser.h"

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>
#include <string.h> // memcpy and memset.

enum TokenType {
	NEWLINE,
	INDENT,
	DEDENT,
	SECTION_NAME,
	ROW_SECTION_NAME,
	GROUP_HEADER,
	ERROR_SENTINEL,
};

#define MAX_DEPTH 64
#define MAX_NAME 32

typedef struct {
	uint16_t statement_indent; // The column of the current statement's first token.
	bool in_header;            // The current statement is a header, which never continues.
	uint8_t pending_dedents;   // When bodies that end before the next statement.
	uint8_t depth;
	uint16_t when_indents[MAX_DEPTH]; // The column of each open when line, outermost first.
} Scanner;

void *tree_sitter_siml_external_scanner_create(void) {
	Scanner *s = ts_calloc(1, sizeof(Scanner));
	return s;
}

void tree_sitter_siml_external_scanner_destroy(void *payload) { ts_free(payload); }

unsigned tree_sitter_siml_external_scanner_serialize(void *payload, char *buffer) {
	Scanner *s = payload;
	unsigned size = offsetof(Scanner, when_indents) + s->depth * sizeof(uint16_t);
	memcpy(buffer, s, size);
	return size;
}

void tree_sitter_siml_external_scanner_deserialize(void *payload, const char *buffer, unsigned length) {
	Scanner *s = payload;
	memset(s, 0, sizeof(Scanner));
	if (length > 0) {
		memcpy(s, buffer, length < sizeof(Scanner) ? length : sizeof(Scanner));
	}
}

static inline void advance(TSLexer *lexer) { lexer->advance(lexer, false); }
static inline void skip(TSLexer *lexer) { lexer->advance(lexer, true); }

static inline bool is_ident_start(int32_t c) { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'; }
static inline bool is_ident_char(int32_t c) { return is_ident_start(c) || (c >= '0' && c <= '9'); }
static inline bool is_blank(int32_t c) { return c == ' ' || c == '\t' || c == '\r' || c == 0xFEFF; }

// Compares by hand: the wasm runtimes that load a grammar give its scanner only part of libc.
static bool same(const char *a, const char *b) {
	while (*a && *a == *b) {
		a++;
		b++;
	}
	return *a == *b;
}

typedef enum { NOT_HEADER, HEADER, ROW_HEADER, GROUP } HeaderKind;

// Reads `name:` or `group "Name":` at the lexer, consuming it; name ends at the word, with markEnd set after it when
// mark is true.
static HeaderKind read_header(TSLexer *lexer, bool mark) {
	if (!is_ident_start(lexer->lookahead)) {
		return NOT_HEADER;
	}
	char name[MAX_NAME + 1];
	unsigned length = 0;
	while (is_ident_char(lexer->lookahead)) {
		if (length < MAX_NAME) {
			name[length] = (char)lexer->lookahead;
		}
		length++;
		advance(lexer);
	}
	if (mark) {
		lexer->mark_end(lexer);
	}
	name[length < MAX_NAME ? length : MAX_NAME] = 0;
	while (lexer->lookahead == ' ' || lexer->lookahead == '\t') {
		advance(lexer);
	}
	if (lexer->lookahead == ':') {
		if (length <= MAX_NAME && (same(name, "prepull") || same(name, "priority"))) {
			return ROW_HEADER;
		}
		return HEADER;
	}
	if (length != 5 || !same(name, "group") || lexer->lookahead != '"') {
		return NOT_HEADER;
	}
	advance(lexer);
	for (;;) {
		if (lexer->eof(lexer) || lexer->lookahead == '\n') {
			return NOT_HEADER;
		}
		if (lexer->lookahead == '"') {
			advance(lexer);
			break;
		}
		if (lexer->lookahead == '\\') {
			advance(lexer);
			if (lexer->eof(lexer) || lexer->lookahead == '\n') {
				return NOT_HEADER;
			}
		}
		advance(lexer);
	}
	while (lexer->lookahead == ' ' || lexer->lookahead == '\t') {
		advance(lexer);
	}
	return lexer->lookahead == ':' ? GROUP : NOT_HEADER;
}

// What follows a line break: the indentation of the next line that is not blank or comment-only, whether a blank line
// comes first, whether that line starts a header, and whether the file ends first.
typedef struct {
	uint16_t indent;
	bool blank, header, eof;
} NextLine;

static NextLine next_line(TSLexer *lexer) {
	NextLine next = {0};
	for (;;) {
		uint16_t indent = 0;
		while (lexer->lookahead == ' ' || lexer->lookahead == '\t') {
			indent++;
			advance(lexer);
		}
		while (lexer->lookahead == '\r') {
			advance(lexer);
		}
		if (lexer->eof(lexer)) {
			next.eof = true;
			return next;
		}
		if (lexer->lookahead == '\n') {
			next.blank = true;
			advance(lexer);
			continue;
		}
		next.indent = indent;
		if (lexer->lookahead == '/') {
			advance(lexer);
			if (lexer->lookahead == '/') {
				while (!lexer->eof(lexer) && lexer->lookahead != '\n') {
					advance(lexer);
				}
				if (!lexer->eof(lexer)) {
					advance(lexer);
				}
				continue;
			}
			return next;
		}
		next.header = read_header(lexer, false) != NOT_HEADER;
		return next;
	}
}

// Closes the when bodies the next statement is not in: those whose line is indented as deep as it or deeper, or all of
// them before a header or the end of the file.
static void close_bodies(Scanner *s, NextLine next) {
	while (s->depth > 0 && (next.eof || next.header || s->when_indents[s->depth - 1] >= next.indent)) {
		s->depth--;
		s->pending_dedents++;
	}
	if (!next.eof) {
		s->statement_indent = next.indent;
	}
}

bool tree_sitter_siml_external_scanner_scan(void *payload, TSLexer *lexer, const bool *valid_symbols) {
	Scanner *s = payload;
	bool recovering = valid_symbols[ERROR_SENTINEL];

	if (!recovering && valid_symbols[DEDENT] && s->pending_dedents > 0) {
		s->pending_dedents--;
		lexer->result_symbol = DEDENT;
		return true;
	}

	// In error recovery every token is valid: a line break that ends the statement still ends it, so the error stays
	// on its line, but no when body opens.
	bool line_end = valid_symbols[NEWLINE] || valid_symbols[INDENT];
	bool section = valid_symbols[SECTION_NAME] || valid_symbols[ROW_SECTION_NAME] || valid_symbols[GROUP_HEADER];
	bool line_start = section && lexer->get_column(lexer) == 0;
	while (is_blank(lexer->lookahead) || (lexer->lookahead == '\n' && !line_end)) {
		if (lexer->lookahead == '\n') {
			line_start = true;
		}
		skip(lexer);
	}

	if (line_end && (lexer->lookahead == '\n' || lexer->eof(lexer))) {
		if (valid_symbols[INDENT] && !recovering) {
			// A when line: its body is what follows deeper than the line itself.
			uint16_t when_indent = s->statement_indent;
			NextLine next = {.eof = true};
			if (!lexer->eof(lexer)) {
				advance(lexer);
				lexer->mark_end(lexer);
				next = next_line(lexer);
			}
			if (s->depth < MAX_DEPTH) {
				s->when_indents[s->depth++] = when_indent;
			}
			s->in_header = false;
			close_bodies(s, next);
			lexer->result_symbol = INDENT;
			return true;
		}
		NextLine next = {.eof = true};
		if (!lexer->eof(lexer)) {
			advance(lexer);
			lexer->mark_end(lexer);
			next = next_line(lexer);
		}
		if (!next.eof && !s->in_header && !next.blank && !next.header && next.indent > s->statement_indent) {
			return false; // The statement continues on that line.
		}
		s->in_header = false;
		close_bodies(s, next);
		lexer->result_symbol = NEWLINE;
		return true;
	}

	if (!recovering && (section || valid_symbols[DEDENT])) {
		s->statement_indent = (uint16_t)lexer->get_column(lexer); // A statement starts here.
	}
	// A header is the first thing on its line.
	if (section && line_start) {
		HeaderKind kind = read_header(lexer, true);
		if (kind == NOT_HEADER) {
			return false;
		}
		enum TokenType token = kind == GROUP ? GROUP_HEADER : kind == ROW_HEADER ? ROW_SECTION_NAME : SECTION_NAME;
		if (kind == ROW_HEADER && !valid_symbols[ROW_SECTION_NAME] && valid_symbols[SECTION_NAME]) {
			token = SECTION_NAME;
		}
		if (!valid_symbols[token]) {
			return false;
		}
		s->in_header = true;
		s->depth = 0;
		s->pending_dedents = 0;
		lexer->result_symbol = token;
		return true;
	}
	return false;
}
