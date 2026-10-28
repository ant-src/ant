package lsp

import "encoding/json"

// The types of the Language Server Protocol, version 3.17, that the server uses: the subset it
// reads and writes, with the field names the protocol gives them.

// Position is a line and a character in a document, both from 0. The character counts UTF-16
// code units, or bytes when the client agreed to utf-8 (spec/lsp.md section 2).
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Range is a span of a document, from Start inclusive to End exclusive.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Location is a range in a document.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// TextEdit replaces a range by text.
type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

// WorkspaceEdit lists the edits to make in each document.
type WorkspaceEdit struct {
	Changes map[string][]TextEdit `json:"changes"`
}

type TextDocumentIdentifier struct {
	URI string `json:"uri"`
}

type VersionedTextDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type TextDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Position               `json:"position"`
}

// ---- lifecycle

type InitializeParams struct {
	ProcessID        *int               `json:"processId"`
	RootURI          string             `json:"rootUri"`
	Capabilities     ClientCapabilities `json:"capabilities"`
	WorkspaceFolders []WorkspaceFolder  `json:"workspaceFolders"`
	ClientInfo       *struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"clientInfo"`
}

type WorkspaceFolder struct {
	URI  string `json:"uri"`
	Name string `json:"name"`
}

// ClientCapabilities is the part of the client's capabilities the server reads.
type ClientCapabilities struct {
	General struct {
		PositionEncodings []string `json:"positionEncodings"`
	} `json:"general"`
	TextDocument struct {
		Completion struct {
			CompletionItem struct {
				SnippetSupport      bool     `json:"snippetSupport"`
				DocumentationFormat []string `json:"documentationFormat"`
				LabelDetailsSupport bool     `json:"labelDetailsSupport"`
			} `json:"completionItem"`
		} `json:"completion"`
		Hover struct {
			ContentFormat []string `json:"contentFormat"`
		} `json:"hover"`
		SignatureHelp struct {
			SignatureInformation struct {
				DocumentationFormat  []string `json:"documentationFormat"`
				ParameterInformation struct {
					LabelOffsetSupport bool `json:"labelOffsetSupport"`
				} `json:"parameterInformation"`
			} `json:"signatureInformation"`
		} `json:"signatureHelp"`
		DocumentSymbol struct {
			HierarchicalDocumentSymbolSupport bool `json:"hierarchicalDocumentSymbolSupport"`
		} `json:"documentSymbol"`
		Rename struct {
			PrepareSupport bool `json:"prepareSupport"`
		} `json:"rename"`
	} `json:"textDocument"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ServerCapabilities struct {
	PositionEncoding           string                  `json:"positionEncoding,omitempty"`
	TextDocumentSync           TextDocumentSyncOptions `json:"textDocumentSync"`
	CompletionProvider         *CompletionOptions      `json:"completionProvider,omitempty"`
	HoverProvider              bool                    `json:"hoverProvider"`
	SignatureHelpProvider      *SignatureHelpOptions   `json:"signatureHelpProvider,omitempty"`
	DefinitionProvider         bool                    `json:"definitionProvider"`
	ReferencesProvider         bool                    `json:"referencesProvider"`
	DocumentHighlightProvider  bool                    `json:"documentHighlightProvider"`
	DocumentSymbolProvider     bool                    `json:"documentSymbolProvider"`
	WorkspaceSymbolProvider    bool                    `json:"workspaceSymbolProvider"`
	CodeActionProvider         *CodeActionOptions      `json:"codeActionProvider,omitempty"`
	DocumentFormattingProvider bool                    `json:"documentFormattingProvider"`
	RenameProvider             *RenameOptions          `json:"renameProvider,omitempty"`
	FoldingRangeProvider       bool                    `json:"foldingRangeProvider"`
	SelectionRangeProvider     bool                    `json:"selectionRangeProvider"`
	SemanticTokensProvider     *SemanticTokensOptions  `json:"semanticTokensProvider,omitempty"`
	InlayHintProvider          bool                    `json:"inlayHintProvider"`
}

type TextDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	Change    int  `json:"change"` // 1 full, 2 incremental
}

type CompletionOptions struct {
	TriggerCharacters []string `json:"triggerCharacters"`
	ResolveProvider   bool     `json:"resolveProvider"`
}

type SignatureHelpOptions struct {
	TriggerCharacters   []string `json:"triggerCharacters"`
	RetriggerCharacters []string `json:"retriggerCharacters"`
}

type CodeActionOptions struct {
	CodeActionKinds []string `json:"codeActionKinds"`
}

type RenameOptions struct {
	PrepareProvider bool `json:"prepareProvider"`
}

type SemanticTokensOptions struct {
	Legend SemanticTokensLegend `json:"legend"`
	Full   bool                 `json:"full"`
}

type SemanticTokensLegend struct {
	TokenTypes     []string `json:"tokenTypes"`
	TokenModifiers []string `json:"tokenModifiers"`
}

// ---- documents

type DidOpenTextDocumentParams struct {
	TextDocument TextDocumentItem `json:"textDocument"`
}

type DidChangeTextDocumentParams struct {
	TextDocument   VersionedTextDocumentIdentifier  `json:"textDocument"`
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

// TextDocumentContentChangeEvent replaces Range by Text, or the whole document when Range is
// nil.
type TextDocumentContentChangeEvent struct {
	Range *Range `json:"range"`
	Text  string `json:"text"`
}

type DidCloseTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidSaveTextDocumentParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DidChangeWatchedFilesParams struct {
	Changes []FileEvent `json:"changes"`
}

type FileEvent struct {
	URI  string `json:"uri"`
	Type int    `json:"type"` // 1 created, 2 changed, 3 deleted
}

// ---- diagnostics

type PublishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Version     *int         `json:"version,omitempty"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type Diagnostic struct {
	Range              Range                          `json:"range"`
	Severity           int                            `json:"severity"` // 1 error, 2 warning, 3 information, 4 hint
	Source             string                         `json:"source"`
	Message            string                         `json:"message"`
	RelatedInformation []DiagnosticRelatedInformation `json:"relatedInformation,omitempty"`
}

type DiagnosticRelatedInformation struct {
	Location Location `json:"location"`
	Message  string   `json:"message"`
}

// ---- completion

type CompletionParams struct {
	TextDocumentPositionParams
	Context *CompletionContext `json:"context"`
}

type CompletionContext struct {
	TriggerKind      int    `json:"triggerKind"` // 1 invoked, 2 trigger character, 3 incomplete
	TriggerCharacter string `json:"triggerCharacter"`
}

type CompletionList struct {
	IsIncomplete bool             `json:"isIncomplete"`
	Items        []CompletionItem `json:"items"`
}

type CompletionItem struct {
	Label            string                  `json:"label"`
	LabelDetails     *CompletionLabelDetails `json:"labelDetails,omitempty"`
	Kind             int                     `json:"kind,omitempty"`
	Detail           string                  `json:"detail,omitempty"`
	Documentation    *MarkupContent          `json:"documentation,omitempty"`
	SortText         string                  `json:"sortText,omitempty"`
	FilterText       string                  `json:"filterText,omitempty"`
	InsertText       string                  `json:"insertText,omitempty"`
	InsertTextFormat int                     `json:"insertTextFormat,omitempty"` // 1 plain, 2 snippet
	TextEdit         *TextEdit               `json:"textEdit,omitempty"`
	CommitCharacters []string                `json:"commitCharacters,omitempty"`
}

type CompletionLabelDetails struct {
	Detail      string `json:"detail,omitempty"`
	Description string `json:"description,omitempty"`
}

// The completion item kinds the server uses.
const (
	ItemText          = 1
	ItemMethod        = 2
	ItemFunction      = 3
	ItemConstructor   = 4
	ItemField         = 5
	ItemVariable      = 6
	ItemClass         = 7
	ItemInterface     = 8
	ItemModule        = 9
	ItemProperty      = 10
	ItemUnit          = 11
	ItemValue         = 12
	ItemEnum          = 13
	ItemKeyword       = 14
	ItemSnippet       = 15
	ItemFile          = 17
	ItemReference     = 18
	ItemFolder        = 19
	ItemConstant      = 21
	ItemStruct        = 22
	ItemTypeParameter = 25
)

type MarkupContent struct {
	Kind  string `json:"kind"` // "plaintext" or "markdown"
	Value string `json:"value"`
}

// ---- hover and signature help

type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type SignatureHelpParams struct {
	TextDocumentPositionParams
	Context *SignatureHelpContext `json:"context"`
}

type SignatureHelpContext struct {
	TriggerKind         int            `json:"triggerKind"`
	TriggerCharacter    string         `json:"triggerCharacter"`
	IsRetrigger         bool           `json:"isRetrigger"`
	ActiveSignatureHelp *SignatureHelp `json:"activeSignatureHelp"`
}

type SignatureHelp struct {
	Signatures      []SignatureInformation `json:"signatures"`
	ActiveSignature int                    `json:"activeSignature"`
	ActiveParameter int                    `json:"activeParameter"`
}

type SignatureInformation struct {
	Label           string                 `json:"label"`
	Documentation   *MarkupContent         `json:"documentation,omitempty"`
	Parameters      []ParameterInformation `json:"parameters"`
	ActiveParameter *int                   `json:"activeParameter,omitempty"`
}

// ParameterInformation names a parameter by the offsets of its label in the signature's label.
type ParameterInformation struct {
	Label         [2]int         `json:"label"`
	Documentation *MarkupContent `json:"documentation,omitempty"`
}

// ---- navigation

type ReferenceParams struct {
	TextDocumentPositionParams
	Context struct {
		IncludeDeclaration bool `json:"includeDeclaration"`
	} `json:"context"`
}

type DocumentHighlight struct {
	Range Range `json:"range"`
	Kind  int   `json:"kind,omitempty"` // 1 text, 2 read, 3 write
}

type RenameParams struct {
	TextDocumentPositionParams
	NewName string `json:"newName"`
}

type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// The symbol kinds the server uses.
const (
	SymbolFile          = 1
	SymbolModule        = 2
	SymbolNamespace     = 3
	SymbolPackage       = 4
	SymbolClass         = 5
	SymbolMethod        = 6
	SymbolProperty      = 7
	SymbolField         = 8
	SymbolConstructor   = 9
	SymbolEnum          = 10
	SymbolInterface     = 11
	SymbolFunction      = 12
	SymbolVariable      = 13
	SymbolConstant      = 14
	SymbolString        = 15
	SymbolStruct        = 23
	SymbolTypeParameter = 26
)

type WorkspaceSymbolParams struct {
	Query string `json:"query"`
}

type SymbolInformation struct {
	Name          string   `json:"name"`
	Kind          int      `json:"kind"`
	Location      Location `json:"location"`
	ContainerName string   `json:"containerName,omitempty"`
}

// ---- formatting, folding, selection

type DocumentFormattingParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Options      struct {
		TabSize      int  `json:"tabSize"`
		InsertSpaces bool `json:"insertSpaces"`
	} `json:"options"`
}

type FoldingRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type FoldingRange struct {
	StartLine      int    `json:"startLine"`
	StartCharacter *int   `json:"startCharacter,omitempty"`
	EndLine        int    `json:"endLine"`
	EndCharacter   *int   `json:"endCharacter,omitempty"`
	Kind           string `json:"kind,omitempty"` // "comment", "imports", "region"
}

type SelectionRangeParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Positions    []Position             `json:"positions"`
}

type SelectionRange struct {
	Range  Range           `json:"range"`
	Parent *SelectionRange `json:"parent,omitempty"`
}

// ---- semantic tokens, inlay hints, code actions

type SemanticTokensParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

type SemanticTokens struct {
	Data []uint32 `json:"data"`
}

type InlayHintParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
}

type InlayHint struct {
	Position     Position `json:"position"`
	Label        string   `json:"label"`
	Kind         int      `json:"kind,omitempty"` // 1 type, 2 parameter
	PaddingLeft  bool     `json:"paddingLeft,omitempty"`
	PaddingRight bool     `json:"paddingRight,omitempty"`
}

type CodeActionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Range        Range                  `json:"range"`
	Context      struct {
		Diagnostics []Diagnostic `json:"diagnostics"`
		Only        []string     `json:"only"`
	} `json:"context"`
}

type CodeAction struct {
	Title       string         `json:"title"`
	Kind        string         `json:"kind,omitempty"`
	Diagnostics []Diagnostic   `json:"diagnostics,omitempty"`
	IsPreferred bool           `json:"isPreferred,omitempty"`
	Edit        *WorkspaceEdit `json:"edit,omitempty"`
}

// ---- messages to the client

type ShowMessageParams struct {
	Type    int    `json:"type"` // 1 error, 2 warning, 3 info, 4 log
	Message string `json:"message"`
}

type LogMessageParams struct {
	Type    int    `json:"type"`
	Message string `json:"message"`
}

// SourceParams is the parameter of the request ant/source, which the client makes for the text
// of a document the server owns, such as the prelude (spec/lsp.md section 4, Definition).
type SourceParams struct {
	URI string `json:"uri"`
}

// raw returns a JSON null as a RawMessage, for a result that is null.
var null = json.RawMessage("null")
