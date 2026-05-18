package cmdhandler

import (
	"errors"
	"fmt"
	"strings"

	"github.com/GizClaw/gizclaw-go/pkg/trie"
)

type CmdHandler struct {
	trie *trie.Trie[Handler]
}

func New() *CmdHandler {
	return &CmdHandler{trie: trie.New[Handler]()}
}

func (m *CmdHandler) Handle(path string, handler Handler) error {
	if handler == nil {
		return errors.New("cmdhandler: nil handler")
	}
	path = strings.Trim(path, "/")
	return m.trie.Set(path, func(ptr *Handler, existed bool) error {
		if existed {
			return fmt.Errorf("cmdhandler: handler already registered for %s", path)
		}
		*ptr = handler
		return nil
	})
}

func (m *CmdHandler) Lookup(path string) (Handler, bool) {
	path = strings.Trim(path, "/")
	ptr, ok := m.trie.Get(path)
	if !ok || ptr == nil || *ptr == nil {
		return nil, false
	}
	return *ptr, true
}
