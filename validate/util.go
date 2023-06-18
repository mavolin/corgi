package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
)

func mixinCallAttrPos(mc file.MixinCall) (pos file.Position) {
	return _mixinCallAttrPos(mc.Body)
}

func _mixinCallAttrPos(s file.Scope) (pos file.Position) {
	fileutil.Walk(s, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.Block:
			return false, nil
		case file.MixinMainBlockShorthand:
			return false, nil
		case file.And:
			pos = itm.Position
			return false, fileutil.StopWalk
		case file.MixinCall:
			if itm.Mixin.File.Module == "" && itm.Mixin.File.ModulePath == "html" && itm.Name.Ident == "Attr" {
				pos = itm.Position
				return false, fileutil.StopWalk
			}

			if itm.Mixin.WritesTopLevelAttributes {
				pos = itm.Position
				return false, fileutil.StopWalk
			}

			unfilledBlocks := make([]file.LinkedMixinBlock, 0, len(itm.Mixin.Blocks))
			for _, block := range itm.Mixin.Blocks {
				if block.TopLevel && block.CanAttributes {
					unfilledBlocks = append(unfilledBlocks, block)
				}
			}

			if len(itm.Body) == 1 {
				if sh, ok := itm.Body[0].(file.MixinMainBlockShorthand); ok {
					for _, ublock := range unfilledBlocks {
						if ublock.Name == "_" {
							if blockPos := _mixinCallAttrPos(sh.Body); blockPos != file.InvalidPosition {
								pos = blockPos
								return false, fileutil.StopWalk
							}

							return false, fileutil.StopWalk
						}
					}
					return false, fileutil.StopWalk
				}
			}

		body:
			for _, itm := range itm.Body {
				block, ok := itm.(file.Block)
				if !ok {
					continue
				}

				for i, ublock := range unfilledBlocks {
					if block.Name.Ident == ublock.Name {
						if blockPos := _mixinCallAttrPos(block.Body); blockPos != file.InvalidPosition {
							pos = blockPos
							return false, fileutil.StopWalk
						}

						copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
						unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
						continue body
					}
				}
			}

			return false, nil
		default:
			return true, nil
		}
	})
	return pos
}

func getAndPlaceholder(acs []file.AttributeCollection) *file.AndPlaceholder {
	for _, ac := range acs {
		al, ok := ac.(file.AttributeList)
		if !ok {
			continue
		}

		for _, a := range al.Attributes {
			if ap, ok := a.(file.AndPlaceholder); ok {
				return &ap
			}
		}
	}

	return nil
}

func interpolationBounds(v file.InterpolationValue) (start, end file.Position) {
	switch v := v.(type) {
	case file.TextInterpolationValue:
		return v.LBracketPos, v.RBracketPos
	case file.ExpressionInterpolationValue:
		return v.LBracePos, v.RBracePos
	default:
		return file.InvalidPosition, file.InvalidPosition
	}
}

func interpolationEnd(v file.InterpolationValue) file.Position {
	switch v := v.(type) {
	case file.TextInterpolationValue:
		return v.RBracketPos
	case file.ExpressionInterpolationValue:
		return v.RBracePos
	default:
		return file.InvalidPosition
	}
}

func inThisMixinCall(f *file.File, mc file.MixinCall) corgierr.Annotation {
	return anno.Anno(f, anno.Annotation{
		Start:      mc.Position,
		Len:        len("+") + (mc.Name.Col - mc.Col) + len(mc.Name.Ident),
		Annotation: "in this mixin call",
	})
}
