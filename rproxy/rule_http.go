package rproxy

import (
	"fmt"
	"sync"
)

// HTTPRule defines the rule structure
type HTTPRule struct {
	URIPrefix, ForwardDst string
}

// HTTPRuleMgr ...
type HTTPRuleMgr struct {
	sync.Map
}

// HTTP global rule manager ...
var HTTP *HTTPRuleMgr

func init() {
	HTTP = &HTTPRuleMgr{}
}

// CreateOrUpdateRule creates or update rule for uri prefix
// uriPrefix is like /path
// forwardDst is the full access address the request will be forward to
func (h *HTTPRuleMgr) CreateOrUpdateRule(uriPrefix, forwardDst string) {
	h.Store(uriPrefix, forwardDst)
}

// RemoveRule removes rule for prefix
func (h *HTTPRuleMgr) RemoveRule(uriPrefix string) error {
	if _, ok := h.Load(uriPrefix); ok {
		h.Delete(ok)
		return nil
	}
	return fmt.Errorf("unknown uri prefix: " + uriPrefix)
}

// ClearRules removes all rules
func (h *HTTPRuleMgr) ClearRules() {
	h.Map = sync.Map{}
}

// ListRules lists all rules
func (h *HTTPRuleMgr) ListRules() []HTTPRule {
	rules := []HTTPRule{}
	h.Range(func(k, v interface{}) bool {
		rule := HTTPRule{URIPrefix: k.(string), ForwardDst: v.(string)}
		rules = append(rules, rule)
		return true
	})
	return rules
}

func (h *HTTPRuleMgr) getRule(uriPrefix string) (forwardDst string, err error) {
	if v, ok := h.Load(uriPrefix); ok {
		return v.(string), nil
	}
	return forwardDst, fmt.Errorf("unknown uri prefix: " + uriPrefix)
}

func (h *HTTPRuleMgr) rangeRule(f func(uriPrefix, forwardDst string) bool) {
	h.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(string))
	})
}
