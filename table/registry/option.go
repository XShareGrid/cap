package registry

import (
	"context"
	"sync"

	"github.com/XShareGrid/cap/i18n"
	"github.com/XShareGrid/cap/msg/errors"
	cap "github.com/XShareGrid/cap/table/proto/go"
)

type optionListStore struct {
	id string
	// ordered list
	list []*cap.OptionValue
	kv   map[int32]*cap.OptionValue
}

func (ols *optionListStore) lookup(id int32) (*cap.OptionValue, error) {
	if v, ok := ols.kv[id]; ok {
		return v, nil
	}
	return nil, errors.Wrap(ErrOptionNotFound).FillDebugArgs(id, ols.id)
}

type OptionType struct {
	ID      string
	Options []*cap.OptionValue
}

// OptionReg option registries
type OptionReg struct {
	m sync.Map
}

// Lookup value for optionID in optionTypeID
func (o *OptionReg) Lookup(optionTypeID string, optionID int32) (*cap.OptionValue, error) {
	if ols, ok := o.m.Load(optionTypeID); ok {
		return ols.(*optionListStore).lookup(optionID)
	}
	return nil, errors.Wrap(ErrOptionTypeNotFound).FillDebugArgs(optionTypeID)
}

// LookupWithI18n value for optionID in optionTypeID
func (o *OptionReg) LookupWithI18n(userLanguage string, optionTypeID string, optionID int32) (*cap.OptionValue, error) {
	if ols, ok := o.m.Load(optionTypeID); ok {
		v, err := ols.(*optionListStore).lookup(optionID)
		if v != nil {
			key := TrKeyOption(optionTypeID, int(v.Id))
			name := i18n.GetTranslation(userLanguage, key)
			if name != key {
				v.Name = name
			}
		}
		return v, err
	}
	return nil, errors.Wrap(ErrOptionTypeNotFound).FillDebugArgs(optionTypeID)
}

// LookupWithI18nCtx value for optionID in optionTypeID
func (o *OptionReg) LookupWithI18nCtx(grpcCtx context.Context, optionTypeID string, optionID int32) (*cap.OptionValue, error) {
	userLanguage := i18n.GetUserLanguageByMeta(grpcCtx)
	return o.LookupWithI18n(userLanguage, optionTypeID, optionID)
}

// GetOptions get option list by type id
func (o *OptionReg) GetOptions(optionTypeID string) ([]*cap.OptionValue, error) {
	if ols, ok := o.m.Load(optionTypeID); ok {
		return ols.(*optionListStore).list, nil
	}
	return nil, errors.Wrap(ErrOptionTypeNotFound).FillDebugArgs(optionTypeID)
}

// GetOptionsWithI18n get option list by type id
func (o *OptionReg) GetOptionsWithI18n(userLanguage string, optionTypeID string) (options []*cap.OptionValue, err error) {
	if ols, ok := o.m.Load(optionTypeID); ok {
		options = ols.(*optionListStore).list
	}
	if options != nil {
		for i, v := range options {
			key := TrKeyOption(optionTypeID, int(v.Id))
			name := i18n.GetTranslation(userLanguage, key)
			if name != key {
				options[i].Name = name
			}
		}
		return options, nil
	}

	return nil, errors.Wrap(ErrOptionTypeNotFound).FillDebugArgs(optionTypeID)
}

// GetOptionsWithI18nCtx get option list by type id
func (o *OptionReg) GetOptionsWithI18nCtx(grpcCtx context.Context, optionTypeID string) ([]*cap.OptionValue, error) {
	userLanguage := i18n.GetUserLanguageByMeta(grpcCtx)
	return o.GetOptionsWithI18n(userLanguage, optionTypeID)
}

// GetAllOptions get all options
func (o *OptionReg) GetAllOptions() []OptionType {
	var optionTypeList []OptionType
	o.m.Range(func(key, value any) bool {
		optionTypeList = append(optionTypeList, OptionType{
			ID:      value.(*optionListStore).id,
			Options: value.(*optionListStore).list,
		})
		return true
	})
	return optionTypeList
}

// Register ...
func (o *OptionReg) Register(optionTypeID string, values []*cap.OptionValue) error {
	if _, ok := o.m.Load(optionTypeID); ok {
		return errors.Wrap(ErrDupplicateNodeID).FillDebugArgs(optionTypeID)
	}
	ols := &optionListStore{
		id:   optionTypeID,
		list: values,
		kv:   map[int32]*cap.OptionValue{},
	}
	for _, v := range values {
		ols.kv[v.Id] = v
	}
	o.m.Store(optionTypeID, ols)
	return nil
}

// Store ...
func (o *OptionReg) Store() *sync.Map {
	return &o.m
}

// NewOptionReg creates option registry
func NewOptionReg() *OptionReg {
	return &OptionReg{}
}
