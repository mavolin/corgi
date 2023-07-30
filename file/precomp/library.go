package precomp

import (
	"fmt"

	cfile "github.com/mavolin/corgi/file"
)

// https://github.com/tinylib/msgp
//go:generate msgp -unexported

type library struct {
	Files []file

	Dependencies []libDependency
	GlobalCode   []code
	Mixins       []mixin
}

func newLibrary(l *cfile.Library) (*library, error) {
	if l == nil {
		return nil, nil
	}

	files := make([]file, len(l.Files))
	for i, f := range l.Files {
		f := f
		files[i] = *newFile(f)
	}
	dependencies := make([]libDependency, len(l.Dependencies))
	for i, d := range l.Dependencies {
		d := d
		dependencies[i] = *newLibDependency(&d)
	}
	globalCode := make([]code, len(l.GlobalCode))
	for i, c := range l.GlobalCode {
		c := c
		globalCode[i] = *newCode(&c)
	}
	mixins := make([]mixin, len(l.Mixins))
	for i, m := range l.Mixins {
		m := m
		var err error
		mixins[i], err = deref(newMixin(l.Files, &m))
		if err != nil {
			return nil, err
		}
	}
	return &library{
		Files:        files,
		Dependencies: dependencies,
		GlobalCode:   globalCode,
		Mixins:       mixins,
	}, nil
}

func (l *library) toFile() *cfile.Library {
	if l == nil {
		return nil
	}

	files := make([]*cfile.File, len(l.Files))
	for i, f := range l.Files {
		files[i] = f.toFile()
	}
	dependencies := make([]cfile.LibDependency, len(l.Dependencies))
	for i, d := range l.Dependencies {
		dependencies[i] = *d.toFile()
	}
	globalCode := make([]cfile.PrecompiledCode, len(l.GlobalCode))
	for i, c := range l.GlobalCode {
		globalCode[i] = *c.toFile()
	}
	mixins := make([]cfile.PrecompiledMixin, len(l.Mixins))
	for i, imp := range l.Mixins {
		mixins[i] = *imp.toFile(files)
	}
	return &cfile.Library{
		Precompiled:  true,
		Files:        files,
		Dependencies: dependencies,
		GlobalCode:   globalCode,
		Mixins:       mixins,
	}
}

type file struct {
	Name       string
	Module     string
	ModulePath string

	Imports []_import
}

func newFile(f *cfile.File) *file {
	if f == nil {
		return nil
	}

	imports := make([]_import, len(f.Imports))
	for i, imp := range f.Imports {
		imp := imp
		imports[i] = *newImport(&imp)
	}
	return &file{
		Name:       f.Name,
		Module:     f.Module,
		ModulePath: f.PathInModule,
		Imports:    imports,
	}
}

func (f *file) toFile() *cfile.File {
	if f == nil {
		return nil
	}

	imports := make([]cfile.Import, len(f.Imports))
	for i, imp := range f.Imports {
		imports[i] = *imp.toFile()
	}
	return &cfile.File{
		Name:         f.Name,
		Module:       f.Module,
		PathInModule: f.ModulePath,
		Imports:      imports,
	}
}

type _import struct {
	Imports  []importSpec
	Position position
}

func newImport(i *cfile.Import) *_import {
	if i == nil {
		return nil
	}

	imports := make([]importSpec, len(i.Imports))
	for i, imp := range i.Imports {
		imp := imp
		imports[i] = *newImportSpec(&imp)
	}
	return &_import{Imports: imports, Position: *newPosition(&i.Position)}
}

func (i *_import) toFile() *cfile.Import {
	if i == nil {
		return nil
	}

	imports := make([]cfile.ImportSpec, len(i.Imports))
	for i, imp := range i.Imports {
		imports[i] = *imp.toFile()
	}
	return &cfile.Import{Imports: imports, Position: *i.Position.toFile()}
}

type importSpec struct {
	Alias *goIdent
	Path  corgiString

	Position position
}

func newImportSpec(is *cfile.ImportSpec) *importSpec {
	if is == nil {
		return nil
	}
	return &importSpec{Alias: newGoIdent(is.Alias), Path: *newCorgiString(&is.Path), Position: *newPosition(&is.Position)}
}

func (is *importSpec) toFile() *cfile.ImportSpec {
	if is == nil {
		return nil
	}
	return &cfile.ImportSpec{Alias: is.Alias.toFile(), Path: *is.Path.toFile(), Position: *is.Position.toFile()}
}

type libDependency struct {
	Module     string
	ModulePath string

	Mixins []mixinDependency
}

func newLibDependency(d *cfile.LibDependency) *libDependency {
	if d == nil {
		return nil
	}

	mixins := make([]mixinDependency, len(d.Mixins))
	for i, m := range d.Mixins {
		m := m
		mixins[i] = *newMixinDependency(&m)
	}
	return &libDependency{Module: d.Module, ModulePath: d.ModulePath, Mixins: mixins}
}

func (d *libDependency) toFile() *cfile.LibDependency {
	if d == nil {
		return nil
	}

	mixins := make([]cfile.MixinDependency, len(d.Mixins))
	for i, m := range d.Mixins {
		mixins[i] = *m.toFile()
	}
	return &cfile.LibDependency{Module: d.Module, ModulePath: d.ModulePath, Mixins: mixins}
}

type mixinDependency struct {
	Name       string
	RequiredBy []string
}

func newMixinDependency(d *cfile.MixinDependency) *mixinDependency {
	if d == nil {
		return nil
	}
	return &mixinDependency{Name: d.Name, RequiredBy: d.RequiredBy}
}

func (d *mixinDependency) toFile() *cfile.MixinDependency {
	if d == nil {
		return nil
	}
	return &cfile.MixinDependency{Name: d.Name, RequiredBy: d.RequiredBy}
}

type code struct {
	MachineComments []string
	Lines           []string
}

func newCode(c *cfile.PrecompiledCode) *code {
	if c == nil {
		return nil
	}
	return &code{MachineComments: c.MachineComments, Lines: c.Lines}
}

func (c *code) toFile() *cfile.PrecompiledCode {
	if c == nil {
		return nil
	}
	return &cfile.PrecompiledCode{MachineComments: c.MachineComments, Lines: c.Lines}
}

type mixin struct {
	FileIndex int

	MachineComments []string

	Name corgiIdent

	LParenPos *position
	Params    []mixinParam
	RParenPos *position

	Position position

	Precompiled []byte

	WritesBody               bool
	WritesElements           bool
	WritesTopLevelAttributes bool
	TopLevelAndPlaceholder   bool
	Blocks                   []mixinBlock
	HasAndPlaceholders       bool
}

func newMixin(fs []*cfile.File, m *cfile.PrecompiledMixin) (*mixin, error) {
	if m == nil {
		return nil, nil
	}

	var fileIndex int
	for i, f := range fs {
		if f.Name == m.File.Name {
			fileIndex = i
			break
		}
	}

	params := make([]mixinParam, len(m.Mixin.Params))
	for i, param := range m.Mixin.Params {
		param := param
		var err error
		params[i], err = deref(newMixinParam(&param))
		if err != nil {
			return nil, err
		}
	}
	blocks := make([]mixinBlock, len(m.Mixin.Blocks))
	for i, block := range m.Mixin.Blocks {
		block := block
		blocks[i] = *newMixinBlock(&block)
	}

	return &mixin{
		FileIndex:                fileIndex,
		MachineComments:          m.MachineComments,
		Name:                     *newCorgiIdent(&m.Mixin.Name),
		LParenPos:                newPosition(m.Mixin.LParenPos),
		Params:                   params,
		RParenPos:                newPosition(m.Mixin.RParenPos),
		Position:                 *newPosition(&m.Mixin.Position),
		Precompiled:              m.Precompiled,
		WritesBody:               m.Mixin.WritesBody,
		WritesElements:           m.Mixin.WritesElements,
		WritesTopLevelAttributes: m.Mixin.WritesTopLevelAttributes,
		TopLevelAndPlaceholder:   m.Mixin.TopLevelAndPlaceholder,
		Blocks:                   blocks,
		HasAndPlaceholders:       m.Mixin.HasAndPlaceholders,
	}, nil
}

func (m *mixin) toFile(fs []*cfile.File) *cfile.PrecompiledMixin {
	if m == nil {
		return nil
	}

	params := make([]cfile.MixinParam, len(m.Params))
	for i, eItm := range m.Params {
		params[i] = *eItm.toFile()
	}
	blocks := make([]cfile.MixinBlockInfo, len(m.Blocks))
	for i, eItm := range m.Blocks {
		blocks[i] = *eItm.toFile()
	}
	return &cfile.PrecompiledMixin{
		File:            fs[m.FileIndex],
		MachineComments: m.MachineComments,
		Mixin: cfile.Mixin{
			Name:      *m.Name.toFile(),
			LParenPos: m.LParenPos.toFile(),
			Params:    params,
			RParenPos: m.RParenPos.toFile(),
			MixinInfo: &cfile.MixinInfo{
				WritesBody:               m.WritesBody,
				WritesElements:           m.WritesElements,
				WritesTopLevelAttributes: m.WritesTopLevelAttributes,
				TopLevelAndPlaceholder:   m.TopLevelAndPlaceholder,
				Blocks:                   blocks,
				HasAndPlaceholders:       m.HasAndPlaceholders,
			},
			Position: *m.Position.toFile(),
		},
		Precompiled: m.Precompiled,
	}
}

type mixinParam struct {
	Name corgiIdent
	Type *goType

	AssignPos *position
	Default   *expression

	Position position
}

func newMixinParam(param *cfile.MixinParam) (*mixinParam, error) {
	if param == nil {
		return nil, nil
	}

	defaultExpr, err := newExpression(param.Default)
	if err != nil {
		return nil, err
	}
	return &mixinParam{
		Name:      *newCorgiIdent(&param.Name),
		Type:      newGoType(param.Type),
		AssignPos: newPosition(param.AssignPos),
		Default:   defaultExpr,
		Position:  *newPosition(&param.Position),
	}, nil
}

func (param *mixinParam) toFile() *cfile.MixinParam {
	if param == nil {
		return nil
	}
	return &cfile.MixinParam{
		Name:      *param.Name.toFile(),
		Type:      param.Type.toFile(),
		AssignPos: param.AssignPos.toFile(),
		Default:   param.Default.toFile(),
		Position:  *param.Position.toFile(),
	}
}

type mixinBlock struct {
	Name                            string
	TopLevel                        bool
	CanAttributes                   bool
	DefaultWritesBody               bool
	DefaultWritesElements           bool
	DefaultWritesTopLevelAttributes bool
	DefaultTopLevelAndPlaceholder   bool
}

func newMixinBlock(mb *cfile.MixinBlockInfo) *mixinBlock {
	if mb == nil {
		return nil
	}
	return &mixinBlock{
		Name:                            mb.Name,
		TopLevel:                        mb.TopLevel,
		CanAttributes:                   mb.CanAttributes,
		DefaultWritesBody:               mb.DefaultWritesBody,
		DefaultWritesElements:           mb.DefaultWritesElements,
		DefaultWritesTopLevelAttributes: mb.DefaultWritesTopLevelAttributes,
		DefaultTopLevelAndPlaceholder:   mb.DefaultTopLevelAndPlaceholder,
	}
}

func (mb *mixinBlock) toFile() *cfile.MixinBlockInfo {
	if mb == nil {
		return nil
	}
	return &cfile.MixinBlockInfo{
		Name:                            mb.Name,
		TopLevel:                        mb.TopLevel,
		CanAttributes:                   mb.CanAttributes,
		DefaultWritesBody:               mb.DefaultWritesBody,
		DefaultWritesElements:           mb.DefaultWritesElements,
		DefaultWritesTopLevelAttributes: mb.DefaultWritesTopLevelAttributes,
		DefaultTopLevelAndPlaceholder:   mb.DefaultTopLevelAndPlaceholder,
	}
}

type expression struct {
	Expressions []expressionItem
}

func newExpression(expr *cfile.Expression) (*expression, error) {
	if expr == nil {
		return nil, nil
	}

	exprs := make([]expressionItem, len(expr.Expressions))
	for i, eItm := range expr.Expressions {
		var err error
		exprs[i], err = deref(newExpressionItem(eItm))
		if err != nil {
			return nil, err
		}
	}
	return &expression{Expressions: exprs}, nil
}

func (expr *expression) toFile() *cfile.Expression {
	if expr == nil {
		return nil
	}

	exprs := make([]cfile.ExpressionItem, len(expr.Expressions))
	for i, eItm := range expr.Expressions {
		exprs[i] = eItm.toFile()
	}
	return &cfile.Expression{Expressions: exprs}
}

type expressionItem struct {
	GoExpression string `msg:",omitempty"`

	Quote    byte                   `msg:",omitempty"`
	Contents []stringExpressionItem `msg:",omitempty"`

	Condition expression `msg:",omitempty"`
	IfTrue    expression `msg:",omitempty"`
	IfFalse   expression `msg:",omitempty"`
	RParenPos position   `msg:",omitempty"`

	Position position
}

func newExpressionItem(itm cfile.ExpressionItem) (*expressionItem, error) {
	if itm == nil {
		return nil, nil
	}

	switch itm := itm.(type) {
	case cfile.GoExpression:
		return &expressionItem{GoExpression: itm.Expression, Position: *newPosition(&itm.Position)}, nil
	case cfile.StringExpression:
		contents := make([]stringExpressionItem, len(itm.Contents))
		for i, seItm := range itm.Contents {
			var err error
			contents[i], err = deref(newStringExpressionItem(seItm))
			if err != nil {
				return nil, err
			}
		}

		return &expressionItem{
			Quote:    itm.Quote,
			Contents: contents,
			Position: *newPosition(&itm.Position),
		}, nil
	case cfile.TernaryExpression:
		cond, err := newExpression(&itm.Condition)
		if err != nil {
			return nil, err
		}
		ifTrue, err := newExpression(&itm.IfTrue)
		if err != nil {
			return nil, err
		}
		ifFalse, err := newExpression(&itm.IfFalse)
		if err != nil {
			return nil, err
		}
		return &expressionItem{
			Condition: *cond,
			IfTrue:    *ifTrue,
			IfFalse:   *ifFalse,
			RParenPos: *newPosition(&itm.RParenPos),
			Position:  *newPosition(&itm.Position),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported file.ExpressionItem: %T", itm)
	}
}

func (itm *expressionItem) toFile() cfile.ExpressionItem {
	if itm == nil {
		return nil
	}

	switch {
	case itm.Quote != 0:
		contents := make([]cfile.StringExpressionItem, len(itm.Contents))
		for i, seItm := range itm.Contents {
			contents[i] = seItm.toFile()
		}

		return cfile.StringExpression{
			Quote:    itm.Quote,
			Contents: contents,
			Position: *itm.Position.toFile(),
		}
	case len(itm.Condition.Expressions) > 0:
		return cfile.TernaryExpression{
			Condition: *itm.Condition.toFile(),
			IfTrue:    *itm.IfTrue.toFile(),
			IfFalse:   *itm.IfFalse.toFile(),
			RParenPos: *itm.RParenPos.toFile(),
			Position:  *itm.Position.toFile(),
		}
	default:
		return cfile.GoExpression{Expression: itm.GoExpression, Position: *itm.Position.toFile()}
	}
}

type stringExpressionItem struct {
	Text string `msg:",omitempty"`

	FormatDirective string     `msg:",omitempty"`
	Expression      expression `msg:",omitempty"`

	Position position
}

func newStringExpressionItem(itm cfile.StringExpressionItem) (*stringExpressionItem, error) {
	if itm == nil {
		return nil, nil
	}

	switch itm := itm.(type) {
	case cfile.StringExpressionText:
		return &stringExpressionItem{Text: itm.Text, Position: *newPosition(&itm.Position)}, nil
	case cfile.StringExpressionInterpolation:
		expr, err := newExpression(&itm.Expression)
		if err != nil {
			return nil, err
		}
		return &stringExpressionItem{
			FormatDirective: itm.FormatDirective,
			Expression:      *expr,
			Position:        *newPosition(&itm.Position),
		}, nil
	default:
		return nil, fmt.Errorf("unknown file.StringExpressionItem: %T", itm)
	}
}

func (itm *stringExpressionItem) toFile() cfile.StringExpressionItem {
	if itm == nil {
		return nil
	}

	if len(itm.Expression.Expressions) > 0 {
		return cfile.StringExpressionInterpolation{
			FormatDirective: itm.FormatDirective,
			Expression:      *itm.Expression.toFile(),
			Position:        *itm.Position.toFile(),
		}
	}

	return cfile.StringExpressionText{Text: itm.Text, Position: *itm.Position.toFile()}
}

type goIdent struct {
	Ident    string
	Position position
}

func newGoIdent(i *cfile.GoIdent) *goIdent {
	if i == nil {
		return nil
	}
	return &goIdent{Ident: i.Ident, Position: *newPosition(&i.Position)}
}

func (i *goIdent) toFile() *cfile.GoIdent {
	if i == nil {
		return nil
	}
	return &cfile.GoIdent{Ident: i.Ident, Position: *i.Position.toFile()}
}

type goType struct {
	Type     string
	Position position
}

func newGoType(t *cfile.GoType) *goType {
	if t == nil {
		return nil
	}
	return &goType{Type: t.Type, Position: *newPosition(&t.Position)}
}

func (t *goType) toFile() *cfile.GoType {
	if t == nil {
		return nil
	}
	return &cfile.GoType{Type: t.Type, Position: *t.Position.toFile()}
}

type corgiIdent struct {
	Ident    string
	Position position
}

func newCorgiIdent(i *cfile.Ident) *corgiIdent {
	if i == nil {
		return nil
	}
	return &corgiIdent{Ident: i.Ident, Position: *newPosition(&i.Position)}
}

func (i *corgiIdent) toFile() *cfile.Ident {
	if i == nil {
		return nil
	}
	return &cfile.Ident{Ident: i.Ident, Position: *i.Position.toFile()}
}

type corgiString struct {
	Quote    byte
	Contents string

	Position position
}

func newCorgiString(s *cfile.String) *corgiString {
	if s == nil {
		return nil
	}
	return &corgiString{Quote: s.Quote, Contents: s.Contents, Position: *newPosition(&s.Position)}
}

func (s *corgiString) toFile() *cfile.String {
	if s == nil {
		return nil
	}
	return &cfile.String{Quote: s.Quote, Contents: s.Contents, Position: *s.Position.toFile()}
}

//msgp:tuple position
type position struct {
	Line int
	Col  int
}

func newPosition(pos *cfile.Position) *position {
	if pos == nil {
		return nil
	}
	return &position{Line: pos.Line, Col: pos.Col}
}

func (pos *position) toFile() *cfile.Position {
	if pos == nil {
		return nil
	}
	return &cfile.Position{Line: pos.Line, Col: pos.Col}
}

func deref[T any](t *T, err error) (T, error) {
	if err != nil {
		var z T
		return z, err
	}

	return *t, nil
}
