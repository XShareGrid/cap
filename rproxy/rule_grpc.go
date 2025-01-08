package rproxy

import (
	"fmt"
	"sync"
)

// GRPCRule defines the rule structure
type GRPCRule struct {
	ServicePrefix, ForwardDst string
	Insecure                  bool
}

// GRPCRuleMgr ...
type GRPCRuleMgr struct {
	sync.Map
}

// GRPC global rule manager ...
var GRPC *GRPCRuleMgr

func init() {
	GRPC = &GRPCRuleMgr{}
}

// CreateOrUpdateRule creates or update rule for uri prefix
// servicePrefix is like /proto.TestService/xxxxx
// forwardDst is the full access address the request will be forward to
func (h *GRPCRuleMgr) CreateOrUpdateRule(servicePrefix, forwardDst string, insecure bool) {
	h.Store(servicePrefix, &GRPCRule{
		ServicePrefix: servicePrefix,
		ForwardDst:    forwardDst,
		Insecure:      insecure})
}

// RemoveRule removes rule for prefix
func (h *GRPCRuleMgr) RemoveRule(uriPrefix string) error {
	if _, ok := h.Load(uriPrefix); ok {
		h.Delete(ok)
		return nil
	}
	return fmt.Errorf("unknown uri prefix: " + uriPrefix)
}

// ClearRules removes all rules
func (h *GRPCRuleMgr) ClearRules() {
	h.Map = sync.Map{}
}

// ListRules lists all rules
func (h *GRPCRuleMgr) ListRules() []GRPCRule {
	rules := []GRPCRule{}
	h.Range(func(k, v interface{}) bool {
		rule := v.(*GRPCRule)
		rules = append(rules, *rule)
		return true
	})
	return rules
}

func (h *GRPCRuleMgr) getRule(uriPrefix string) (rule *GRPCRule, err error) {
	if v, ok := h.Load(uriPrefix); ok {
		return v.(*GRPCRule), nil
	}
	return nil, fmt.Errorf("unknown uri prefix: " + uriPrefix)
}

func (h *GRPCRuleMgr) rangeRule(f func(servicePrefix string, rule *GRPCRule) bool) {
	h.Range(func(k, v interface{}) bool {
		return f(k.(string), v.(*GRPCRule))
	})
}
