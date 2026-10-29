// Package lsp is the language server behind ant lsp (spec/lsp.md): it speaks the Language Server
// Protocol over a stream, analyzing each open document as the root of its own program.
package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"runtime/debug"
	"slices"
	"strings"
)

// Server is one language server, serving one client.
type Server struct {
	conn *conn
	log  *log.Logger
	// utf16 says positions count UTF-16 code units; false means bytes (utf-8).
	utf16 bool
	docs  map[string]*document // open documents by URI
	// gen counts changes; an analysis is cached for the generation it was made in.
	gen      int
	cache    map[string]*analysis
	cacheGen int
	roots    []string // the workspace folders' paths
	prelude  *document
	caps     ClientCapabilities
	files    map[string]*workspaceFile // the workspace's parsed files, by path

	initialized bool
	shutdown    bool
	exitCode    int
}

// Serve runs a server over r and w until the client sends exit or the stream ends, logging to
// logw, and returns the exit status (spec/lsp.md section 1).
func Serve(ctx context.Context, r io.Reader, w io.Writer, logw io.Writer) int {
	s := &Server{
		conn:  &conn{r: bufio.NewReaderSize(r, 1<<20), w: w},
		log:   log.New(logw, "ant lsp: ", 0),
		utf16: true,
		docs:  map[string]*document{},
		cache: map[string]*analysis{},
	}
	return s.run(ctx)
}

func (s *Server) run(ctx context.Context) int {
	for {
		if ctx.Err() != nil {
			return 1
		}
		m, err := s.conn.read()
		if err != nil {
			if err == errClosed {
				return 1 // only exit after shutdown ends with 0 (spec/lsp.md section 1)
			}
			var re *rpcError
			if e, ok := err.(*rpcError); ok {
				re = e
			}
			if re != nil {
				s.conn.fail(null, re.Code, re.Message)
				continue
			}
			s.log.Printf("read: %v", err)
			return 1
		}
		if m.Method == "" {
			continue // a response to a request of ours; none is outstanding
		}
		if done := s.handle(m); done {
			return s.exitCode
		}
	}
}

// handle dispatches one request or notification and reports whether the client said exit.
func (s *Server) handle(m *message) (exit bool) {
	defer func() {
		if r := recover(); r != nil {
			s.log.Printf("%s: panic: %v\n%s", m.Method, r, debug.Stack())
			if m.request() {
				s.conn.fail(m.ID, codeInternalError, fmt.Sprintf("internal error: %v", r))
			}
		}
	}()
	switch m.Method {
	case "initialize":
		var params InitializeParams
		s.params(m, &params)
		s.conn.reply(m.ID, s.initialize(&params))
		return false
	case "initialized":
		s.initialized = true
		return false
	case "shutdown":
		s.shutdown = true
		s.conn.reply(m.ID, nil)
		return false
	case "exit":
		if !s.shutdown {
			s.exitCode = 1
		}
		return true
	case "$/cancelRequest", "$/setTrace", "workspace/didChangeConfiguration":
		return false
	}
	if !s.initialized && m.Method != "initialize" {
		if m.request() {
			s.conn.fail(m.ID, codeServerNotInit, "the server is not initialized")
		}
		return false
	}
	result, err := s.dispatch(m)
	if m.request() {
		if err != nil {
			s.conn.fail(m.ID, err.Code, err.Message)
		} else {
			s.conn.reply(m.ID, result)
		}
	} else if err != nil {
		s.log.Printf("%s: %s", m.Method, err.Message)
	}
	return false
}

// params decodes a message's params into v, tolerating absent ones.
func (s *Server) params(m *message, v any) *rpcError {
	if len(m.Params) == 0 || string(m.Params) == "null" {
		return nil
	}
	if err := json.Unmarshal(m.Params, v); err != nil {
		return &rpcError{Code: codeInvalidParams, Message: err.Error()}
	}
	return nil
}

// dispatch answers the requests and notifications of the protocol.
func (s *Server) dispatch(m *message) (any, *rpcError) {
	switch m.Method {
	case "textDocument/didOpen":
		var p DidOpenTextDocumentParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		s.docs[p.TextDocument.URI] = newDocument(p.TextDocument.URI, p.TextDocument.Version, p.TextDocument.Text)
		s.changed()
		s.publishAll()
	case "textDocument/didChange":
		var p DidChangeTextDocumentParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		d, ok := s.docs[p.TextDocument.URI]
		if !ok {
			return nil, nil
		}
		d.version = p.TextDocument.Version
		d.apply(p.ContentChanges, s.utf16)
		s.changed()
		s.publishAll()
	case "textDocument/didClose":
		var p DidCloseTextDocumentParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		if _, ok := s.docs[p.TextDocument.URI]; ok {
			delete(s.docs, p.TextDocument.URI)
			s.conn.notify("textDocument/publishDiagnostics", PublishDiagnosticsParams{URI: p.TextDocument.URI, Diagnostics: []Diagnostic{}})
			s.changed()
			s.publishAll()
		}
	case "textDocument/didSave":
		// The text is the document's; nothing changes.
	case "workspace/didChangeWatchedFiles":
		s.changed()
		s.publishAll()
	case "workspace/didChangeWorkspaceFolders":
		var p struct {
			Event struct {
				Added   []WorkspaceFolder `json:"added"`
				Removed []WorkspaceFolder `json:"removed"`
			} `json:"event"`
		}
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		for _, f := range p.Event.Removed {
			s.roots = slices.DeleteFunc(s.roots, func(r string) bool { return r == uriPath(f.URI) })
		}
		for _, f := range p.Event.Added {
			s.roots = append(s.roots, uriPath(f.URI))
		}
	case "textDocument/completion":
		var p CompletionParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.completion(&p), nil
	case "textDocument/hover":
		var p TextDocumentPositionParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.hover(&p), nil
	case "textDocument/signatureHelp":
		var p SignatureHelpParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.signatureHelp(&p), nil
	case "textDocument/definition":
		var p TextDocumentPositionParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.definition(&p), nil
	case "textDocument/references":
		var p ReferenceParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.references(&p), nil
	case "textDocument/documentHighlight":
		var p TextDocumentPositionParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.highlight(&p), nil
	case "textDocument/prepareRename":
		var p TextDocumentPositionParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.prepareRename(&p)
	case "textDocument/rename":
		var p RenameParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.rename(&p)
	case "textDocument/documentSymbol":
		var p DocumentSymbolParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.documentSymbols(&p), nil
	case "workspace/symbol":
		var p WorkspaceSymbolParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.workspaceSymbols(&p), nil
	case "textDocument/formatting":
		var p DocumentFormattingParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.formatting(&p), nil
	case "textDocument/foldingRange":
		var p FoldingRangeParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.foldingRanges(&p), nil
	case "textDocument/selectionRange":
		var p SelectionRangeParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.selectionRanges(&p), nil
	case "textDocument/semanticTokens/full":
		var p SemanticTokensParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.semanticTokens(&p), nil
	case "textDocument/inlayHint":
		var p InlayHintParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.inlayHints(&p), nil
	case "textDocument/codeAction":
		var p CodeActionParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		return s.codeActions(&p), nil
	case "ant/source":
		var p SourceParams
		if err := s.params(m, &p); err != nil {
			return nil, err
		}
		if p.URI == preludeURI {
			return s.preludeDoc().text, nil
		}
		return nil, &rpcError{Code: codeInvalidParams, Message: "no such document: " + p.URI}
	default:
		if m.request() {
			return nil, &rpcError{Code: codeMethodNotFound, Message: "unknown method " + m.Method}
		}
	}
	return nil, nil
}

// changed notes that a document changed, so that every analysis is made again.
func (s *Server) changed() { s.gen++ }

// doc returns the open document at uri when it is a file the server analyzes, the prelude's
// own document for its URI, or nil.
func (s *Server) doc(uri string) *document {
	if uri == preludeURI {
		return s.preludeDoc()
	}
	d, ok := s.docs[uri]
	if !ok || d.path == "" {
		return nil
	}
	return d
}

// byPath maps each open file's path to its document.
func (s *Server) byPath() map[string]*document {
	out := make(map[string]*document, len(s.docs))
	for _, d := range s.docs {
		if d.path != "" {
			out[d.path] = d
		}
	}
	return out
}

// initialize answers the client's first request: the server's capabilities, and the position
// encoding, utf-8 when the client offers it (spec/lsp.md section 2).
func (s *Server) initialize(p *InitializeParams) InitializeResult {
	s.caps = p.Capabilities
	for _, f := range p.WorkspaceFolders {
		s.roots = append(s.roots, uriPath(f.URI))
	}
	if len(s.roots) == 0 && p.RootURI != "" {
		s.roots = append(s.roots, uriPath(p.RootURI))
	}
	encoding := "utf-16"
	if slices.Contains(p.Capabilities.General.PositionEncodings, "utf-8") {
		encoding, s.utf16 = "utf-8", false
	}
	return InitializeResult{
		ServerInfo: ServerInfo{Name: "ant lsp"},
		Capabilities: ServerCapabilities{
			PositionEncoding:           encoding,
			TextDocumentSync:           TextDocumentSyncOptions{OpenClose: true, Change: 2},
			CompletionProvider:         &CompletionOptions{TriggerCharacters: []string{".", "|", "<", "/", "\"", "{", "("}},
			HoverProvider:              true,
			SignatureHelpProvider:      &SignatureHelpOptions{TriggerCharacters: []string{"(", ","}, RetriggerCharacters: []string{","}},
			DefinitionProvider:         true,
			ReferencesProvider:         true,
			DocumentHighlightProvider:  true,
			DocumentSymbolProvider:     true,
			WorkspaceSymbolProvider:    true,
			CodeActionProvider:         &CodeActionOptions{CodeActionKinds: []string{"quickfix"}},
			DocumentFormattingProvider: true,
			RenameProvider:             &RenameOptions{PrepareProvider: true},
			FoldingRangeProvider:       true,
			SelectionRangeProvider:     true,
			SemanticTokensProvider:     &SemanticTokensOptions{Legend: semanticLegend, Full: true},
			InlayHintProvider:          true,
		},
	}
}

// markdown wraps text as markdown content.
func markdown(text string) *MarkupContent {
	return &MarkupContent{Kind: "markdown", Value: text}
}

// code wraps Ant source as a fenced block.
func code(src string) string {
	return "```ant\n" + strings.TrimRight(src, "\n") + "\n```"
}
