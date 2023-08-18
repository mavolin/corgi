package link

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/internal/stack"
)

type queuedMixin struct {
	f *file.File
	m *file.Mixin
}

func (l *Linker) analyzeMixins(fs ...*file.File) *errList {
	var errs errList

	var ms list.List[queuedMixin]

	for _, f := range fs {
		fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
			_, ok := (*ctx.Item).(file.Mixin)
			if !ok {
				return true, nil
			}

			// see mixin_call.go as to why this is necessary
			m := ptrOfSliceElem[file.ScopeItem, file.Mixin](ctx.Scope, ctx.Index)

			if abort := l.analyzeMixin(f, m); abort {
				ms.PushBack(queuedMixin{f, m})
			}

			// mixins cannot contain mixins
			return false, nil
		})
	}

	for ms.Len() > 0 {
		startLen := ms.Len()

		for mE := ms.Front(); mE != nil; mE = mE.Next() {
			m := mE.V()
			if abort := l.analyzeMixin(m.f, m.m); !abort {
				ms.Remove(mE)
			}
		}

		if startLen != ms.Len() {
			continue
		}

		for mE := ms.Front(); mE != nil; mE = mE.Next() {
			m := mE.V()
			errs.PushBack(&corgierr.Error{
				Message: "linker: failed to analyze mixin",
				ErrorAnnotation: anno.Anno(m.f, anno.Annotation{
					Start:      m.m.Name.Position,
					Len:        len(m.m.Name.Ident),
					Annotation: "the linker failed to analyze this mixin, most likely because of a, possibly indirect, recursion",
				}),
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: "Corgi does not allow mixins to recursively call themselves, even indirectly,\n" +
							"i.e. mixin `a` calls mixin `b` which calls mixin `a`.\n" +
							"However, this and the other mixins that this error is reported on, perform direct or" +
							"indirect recursion on each other.\n" +
							"This error can be resolved by removing this recursion.",
					},
				},
			})
		}
		return &errs
	}

	return &errs
}

var (
	htmlAttrMixinInfo = file.MixinInfo{
		WritesTopLevelAttributes: true,
	}
	htmlElementMixinInfo = file.MixinInfo{
		WritesBody:     true,
		WritesElements: true,
		Blocks: []file.MixinBlockInfo{
			{
				Name:          "_",
				CanAttributes: true,
			},
		},
		HasAndPlaceholders: true,
	}
)

func (l *Linker) analyzeMixin(f *file.File, m *file.Mixin) bool {
	m.MixinInfo = new(file.MixinInfo)
	var abort bool

	if fileutil.IsAttrMixin(file.LinkedMixin{File: f, Mixin: m}) {
		*m.MixinInfo = htmlAttrMixinInfo
		return false
	} else if fileutil.IsElementMixin(file.LinkedMixin{File: f, Mixin: m}) {
		*m.MixinInfo = htmlElementMixinInfo
		return false
	}

	blockInfos := make(map[string]file.MixinBlockInfo)

	var canAttrs stack.Stack[bool]
	canAttrs.Push(true)

	fileutil.Walk(m.Body, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		if len(parents)+1 > canAttrs.Len() {
			switch (*parents[len(parents)-1].Item).(type) {
			case file.Element:
				canAttrs.Push(true)
			case file.DivShorthand:
				canAttrs.Push(true)
			default:
				canAttrs.Push(canAttrs.Peek())
			}
		} else if len(parents)+1 < canAttrs.Len() {
			for i := 0; i < canAttrs.Len()-(len(parents)+1); i++ {
				canAttrs.Pop()
			}
		}

		switch itm := (*ctx.Item).(type) {
		case file.Element:
			m.WritesElements = true
			m.WritesBody = true

			if !m.HasAndPlaceholders && hasAndPlaceholder(itm.Attributes) {
				m.HasAndPlaceholders = true
			}

			canAttrs.Swap(false)
		case file.ArrowBlock:
			m.WritesBody = true
			canAttrs.Swap(false)
		case file.InlineText:
			m.WritesBody = true
			canAttrs.Swap(false)
		case file.HTMLComment:
			m.WritesBody = true
			m.WritesElements = true
			canAttrs.Swap(false)
		case file.And:
			topLvl := isTopLevel(parents)
			if topLvl {
				m.WritesTopLevelAttributes = true
			}

			if (!m.HasAndPlaceholders || !m.TopLevelAndPlaceholder) && hasAndPlaceholder(itm.Attributes) {
				m.HasAndPlaceholders = true
				if topLvl {
					m.TopLevelAndPlaceholder = true
				}
			}
		case file.MixinCall:
			topLvl := isTopLevel(parents)
			anal, abort2 := analyzeMixinCall(m.MixinInfo, itm, topLvl, canAttrs.Peek(), blockInfos)
			if abort2 {
				abort = true
				return false, fileutil.StopWalk
			}

			if anal.writesBody {
				m.WritesBody = true
			}
			if anal.writesElements {
				m.WritesElements = true
			}
			if topLvl && anal.writesTopLevelAttrs {
				m.WritesTopLevelAttributes = true
			}
			if anal.usesAndPlaceholders {
				m.HasAndPlaceholders = true
			}
			if topLvl && anal.usesTopLvlAndPlaceholder {
				m.TopLevelAndPlaceholder = true
			}
			return false, nil
		case file.Block:
			_, abort2 := analyzeBlock(m.MixinInfo, itm, isTopLevel(parents), canAttrs.Peek(), blockInfos)
			if abort2 {
				abort = true
				return false, fileutil.StopWalk
			}

			canAttrs.Swap(false)

			return false, nil
		}

		return true, nil
	})

	m.Blocks = make([]file.MixinBlockInfo, 0, len(blockInfos))
	for _, info := range blockInfos {
		m.Blocks = append(m.Blocks, info)
	}

	if abort {
		m.MixinInfo = nil
	}

	return abort
}

func isTopLevel(parents []fileutil.WalkContext) bool {
	for _, parent := range parents {
		switch (*parent.Item).(type) {
		case file.If:
		case file.IfBlock:
		case file.Switch:
		case file.For:
		default:
			return false
		}
	}

	return true
}

func hasAndPlaceholder(acolls []file.AttributeCollection) bool {
	for _, acoll := range acolls {
		al, ok := acoll.(file.AttributeList)
		if !ok {
			continue
		}

		for _, attr := range al.Attributes {
			if _, ok := attr.(file.AndPlaceholder); ok {
				return true
			}
		}
	}

	return false
}

func analyzeBlock(mi *file.MixinInfo, b file.Block, topLvl, canAttrs bool, blockInfos map[string]file.MixinBlockInfo) (file.MixinBlockInfo, bool) {
	bi, ok := blockInfos[b.Name.Ident]
	bi.Name = b.Name.Ident
	if !ok {
		bi.CanAttributes = canAttrs
	} else if !canAttrs {
		bi.CanAttributes = false
	}
	if topLvl {
		bi.TopLevel = true
	}

	defer func() { blockInfos[b.Name.Ident] = bi }()

	var canAttrsStack stack.Stack[bool]
	canAttrsStack.Push(canAttrs)

	var abort bool
	fileutil.Walk(b.Body, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		if len(parents)+1 > canAttrsStack.Len() {
			switch (*parents[len(parents)-1].Item).(type) {
			case file.Element:
				canAttrsStack.Push(true)
			case file.DivShorthand:
				canAttrsStack.Push(true)
			default:
				canAttrsStack.Push(canAttrsStack.Peek())
			}
		} else if len(parents)+1 < canAttrsStack.Len() {
			for i := 0; i < canAttrsStack.Len()-(len(parents)+1); i++ {
				canAttrsStack.Pop()
			}
		}

		switch itm := (*ctx.Item).(type) {
		case file.Element:
			bi.DefaultWritesBody = true
			bi.DefaultWritesElements = true
			canAttrsStack.Swap(false)
		case file.ArrowBlock:
			bi.DefaultWritesBody = true
			canAttrsStack.Swap(false)
		case file.InlineText:
			bi.DefaultWritesBody = true
			canAttrsStack.Swap(false)
		case file.HTMLComment:
			bi.DefaultWritesBody = true
			bi.DefaultWritesElements = true
			canAttrsStack.Swap(false)
		case file.And:
			topLvl := isTopLevel(parents)
			if topLvl {
				bi.DefaultWritesTopLevelAttributes = true

				if !bi.DefaultTopLevelAndPlaceholder && hasAndPlaceholder(itm.Attributes) {
					bi.DefaultTopLevelAndPlaceholder = true
				}
			}
		case file.MixinCall:
			anal, abort2 := analyzeMixinCall(mi, itm, topLvl && isTopLevel(parents), canAttrsStack.Peek(), blockInfos)
			if abort2 {
				abort = true
				return false, fileutil.StopWalk
			}

			if anal.writesBody {
				bi.DefaultWritesBody = true
			}
			if anal.writesElements {
				bi.DefaultWritesElements = true
			}
			if anal.writesTopLevelAttrs {
				bi.DefaultWritesTopLevelAttributes = true
			}
			if anal.usesAndPlaceholders {
				mi.HasAndPlaceholders = true
			}
			if anal.usesTopLvlAndPlaceholder {
				bi.DefaultTopLevelAndPlaceholder = true
			}
			return false, nil
		case file.Block:
			subBi, subAbort := analyzeBlock(mi, itm, topLvl && isTopLevel(parents), canAttrsStack.Peek(), blockInfos)
			if subAbort {
				abort = true
				return false, fileutil.StopWalk
			}

			if subBi.DefaultWritesBody {
				bi.DefaultWritesBody = true
			}

			canAttrsStack.Swap(false)
			return false, nil
		}

		return true, nil
	})

	return bi, abort
}

type mixinCallAnalysis struct {
	writesBody               bool
	writesElements           bool
	writesTopLevelAttrs      bool
	usesAndPlaceholders      bool
	usesTopLvlAndPlaceholder bool
}

func analyzeMixinCall(mi *file.MixinInfo, mc file.MixinCall, topLvl, canAttrs bool, blockInfos map[string]file.MixinBlockInfo) (anal mixinCallAnalysis, abort bool) {
	if mc.Mixin == nil || mc.Mixin.Mixin.MixinInfo == nil {
		return mixinCallAnalysis{}, true
	}

	if mc.Mixin.Mixin.WritesBody {
		anal.writesBody = true
	}
	if mc.Mixin.Mixin.WritesElements {
		anal.writesElements = true
	}
	if mc.Mixin.Mixin.WritesTopLevelAttributes {
		anal.writesTopLevelAttrs = true
	}
	if mc.Mixin.Mixin.HasAndPlaceholders {
		anal.usesAndPlaceholders = true
	}

	blocks := make(map[string]*mixinCallBlockAnalysis, len(mc.Mixin.Mixin.Blocks))
	for _, block := range mc.Mixin.Mixin.Blocks {
		blocks[block.Name] = &mixinCallBlockAnalysis{
			topLvl:              block.TopLevel,
			writesBody:          block.DefaultWritesBody,
			writesElements:      block.DefaultWritesElements,
			writesTopLevelAttrs: block.DefaultWritesTopLevelAttributes,
			// usesAndPlaceholders: , <- false friend, this refers to whether the outer mixin's
			// and placeholder is placed inside this block, not whether this block's
			// default has an and placeholder
			isDefault: true,
		}
	}

	if len(mc.Body) == 1 {
		switch itm := mc.Body[0].(type) {
		case file.BlockExpansion:
			var blockInfo *file.MixinBlockInfo
			for _, block := range mc.Mixin.Mixin.Blocks {
				if block.Name == "_" {
					block := block
					blockInfo = &block
					break
				}
			}
			if blockInfo == nil {
				// validate will catch this
				break
			}

			abort = analyzeMixinCallBlock(mi, blocks["_"], file.Block{
				Type: file.BlockTypeBlock,
				Name: file.Ident{Ident: "_"},
				Body: file.Scope{itm.Item},
			}, *blockInfo, topLvl, canAttrs, blockInfos)
			if abort {
				return anal, true
			}
			goto analyzed
		case file.MixinMainBlockShorthand:
			var blockInfo *file.MixinBlockInfo
			for _, block := range mc.Mixin.Mixin.Blocks {
				if block.Name == "_" {
					block := block
					blockInfo = &block
					break
				}
			}
			if blockInfo == nil {
				// validate will catch this
				break
			}

			abort = analyzeMixinCallBlock(mi, blocks["_"], file.Block{
				Type: file.BlockTypeBlock,
				Name: file.Ident{Ident: "_"},
				Body: itm.Body,
			}, *blockInfo, topLvl, canAttrs, blockInfos)
			if abort {
				return anal, true
			}
			goto analyzed
		}
	}

	fileutil.Walk(mc.Body, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.And:
			if mc.Mixin.Mixin.TopLevelAndPlaceholder {
				anal.writesTopLevelAttrs = true
			}

			if hasAndPlaceholder(itm.Attributes) {
				anal.usesAndPlaceholders = true

				if mc.Mixin.Mixin.TopLevelAndPlaceholder {
					anal.usesTopLvlAndPlaceholder = true
				}
			}
		case file.Block:
			var blockInfo *file.MixinBlockInfo
			for _, block := range mc.Mixin.Mixin.Blocks {
				if block.Name == itm.Name.Ident {
					block := block
					blockInfo = &block
					break
				}
			}
			if blockInfo == nil {
				// validate will catch this
				return false, fileutil.StopWalk
			}

			abort = analyzeMixinCallBlock(mi, blocks[itm.Name.Ident], itm, *blockInfo, topLvl, canAttrs, blockInfos)
			if abort {
				return false, fileutil.StopWalk
			}
			return false, nil
		case file.MixinCall:
			subAnal, abort2 := analyzeMixinCall(mi, itm, topLvl && isTopLevel(parents), canAttrs, blockInfos)
			if abort2 {
				abort = true
				return false, fileutil.StopWalk
			}

			if subAnal.writesTopLevelAttrs {
				anal.writesTopLevelAttrs = true
			}
			if subAnal.usesAndPlaceholders {
				anal.usesAndPlaceholders = true
			}
			if subAnal.usesTopLvlAndPlaceholder {
				anal.usesTopLvlAndPlaceholder = true
			}

			return false, nil
		}
		return true, nil
	})

analyzed:
	for _, block := range blocks {
		if block.writesBody {
			anal.writesBody = true
		}
		if block.writesElements {
			anal.writesElements = true
		}
		if block.topLvl {
			if block.writesTopLevelAttrs {
				anal.writesTopLevelAttrs = true
			}
			if block.usesTopLvlAndPlaceholder {
				anal.usesTopLvlAndPlaceholder = true
			}
		}
		if block.usesTopLvlAndPlaceholder {
			anal.usesAndPlaceholders = true
		}
	}

	return anal, abort
}

type mixinCallBlockAnalysis struct {
	topLvl                   bool
	writesBody               bool
	writesElements           bool
	writesTopLevelAttrs      bool
	usesAndPlaceholders      bool
	usesTopLvlAndPlaceholder bool

	isDefault bool
}

func analyzeMixinCallBlock(
	mi *file.MixinInfo, anal *mixinCallBlockAnalysis, b file.Block, bInfo file.MixinBlockInfo, mixinTopLvl, mixinCanAttrs bool, blockInfos map[string]file.MixinBlockInfo,
) (abort bool) {
	if anal.isDefault {
		*anal = mixinCallBlockAnalysis{topLvl: anal.topLvl}
	}

	var canAttrsStack stack.Stack[bool]
	if bInfo.TopLevel {
		canAttrsStack.Push(bInfo.CanAttributes && mixinCanAttrs)
	} else {
		canAttrsStack.Push(bInfo.CanAttributes)
	}

	fileutil.Walk(b.Body, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		if len(parents)+1 > canAttrsStack.Len() {
			switch (*parents[len(parents)-1].Item).(type) {
			case file.Element:
				canAttrsStack.Push(true)
			case file.DivShorthand:
				canAttrsStack.Push(true)
			default:
				canAttrsStack.Push(canAttrsStack.Peek())
			}
		} else if len(parents)+1 < canAttrsStack.Len() {
			for i := 0; i < canAttrsStack.Len()-(len(parents)+1); i++ {
				canAttrsStack.Pop()
			}
		}

		switch itm := (*ctx.Item).(type) {
		case file.Element:
			anal.writesBody = true
			anal.writesElements = true
			canAttrsStack.Swap(false)
		case file.ArrowBlock:
			anal.writesBody = true
			canAttrsStack.Swap(false)
		case file.InlineText:
			anal.writesBody = true
			canAttrsStack.Swap(false)
		case file.HTMLComment:
			anal.writesBody = true
			anal.writesElements = true
			canAttrsStack.Swap(false)
		case file.And:
			topLevel := isTopLevel(parents)
			if topLevel {
				anal.writesTopLevelAttrs = true
			}
			if hasAndPlaceholder(itm.Attributes) {
				anal.usesAndPlaceholders = true

				if topLevel {
					anal.usesTopLvlAndPlaceholder = true
				}
			}
		case file.MixinCall:
			callAnal, abort2 := analyzeMixinCall(mi, itm, mixinTopLvl && bInfo.TopLevel && isTopLevel(parents), canAttrsStack.Peek(), blockInfos)
			if abort2 {
				abort = true
				return false, fileutil.StopWalk
			}
			if callAnal.writesBody {
				anal.writesBody = true
			}
			if callAnal.writesElements {
				anal.writesElements = true
			}
			if callAnal.writesTopLevelAttrs {
				anal.writesTopLevelAttrs = true
			}
			if callAnal.usesAndPlaceholders {
				anal.usesAndPlaceholders = true
			}
			if callAnal.usesTopLvlAndPlaceholder {
				anal.usesTopLvlAndPlaceholder = true
			}

			return false, nil
		case file.Block:
			blockAnal, abort2 := analyzeBlock(mi, itm, mixinTopLvl && bInfo.TopLevel && isTopLevel(parents), canAttrsStack.Peek(), blockInfos)
			if abort2 {
				abort = true
				return false, fileutil.StopWalk
			}

			if blockAnal.DefaultWritesBody {
				anal.writesBody = true
			}
			if blockAnal.DefaultWritesElements {
				anal.writesElements = true
			}

			canAttrsStack.Swap(false)
			return false, nil
		}

		return true, nil
	})

	return abort
}
