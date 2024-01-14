package typeinfer

import (
	"errors"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/internal/anno"
)

// Scope infers the types of all components and state variables in s.
func Scope(f *file.File, s file.Scope) error {
	errs := make([]error, 0, len(s.Items))

	for i, itm := range s.Items {
		switch itm := itm.(type) {
		case file.Component:
			errs = append(errs, componentParams(f, &itm)...)
			s.Items[i] = itm
		case file.State:
			errs = append(errs, stateVars(f, &itm)...)
			s.Items[i] = itm
		}
	}

	return errors.Join(errs...)
}

// componentParams attempts to infer the type of m's params without an explicitly
// set type but a default expression.
//
// When it succeeds, it stores the inferred type as
// [file.componentParam.InferredType].
func componentParams(f *file.File, m *file.Component) []error {
	errs := make([]error, 0, len(m.Params))

	for i, param := range m.Params {
		if param.Type == nil && param.Default != nil {
			param.InferredType = Infer(*param.Default)
			if param.InferredType == "" {
				errs = append(errs, &fileerr.Error{
					Message: "component param: unable to infer type",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start: param.Name.Position,
						Len:   len(param.Name.Ident),
						Annotation: "this param has no explicit type,\n" +
							"and no type could be inferred from the default",
					}),
					Suggestions: []fileerr.Suggestion{
						{
							Suggestion: "give this param an explicit type",
							Example:    "`" + param.Name.Ident + " string: ...`",
						},
					},
				})
			}

			m.Params[i] = param
		}
	}

	return errs
}

func stateVars(f *file.File, state *file.State) []error {
	errs := make([]error, 0, len(state.Vars))

	for i, v := range state.Vars {
		sv, ok := v.(file.StateVar)
		if !ok || sv.Type != nil {
			continue
		}

		sv.InferredType = Infer(sv.Values[0])
		if sv.InferredType == "" {
			errs = append(errs, &fileerr.Error{
				Message: "state variable: unable to infer type",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start: sv.Names[0].Position,
					Len:   len(sv.Names[0].Ident),
					Annotation: "this state variable has no explicit type,\n" +
						"and no type could be inferred from the default",
				}),
				Suggestions: []fileerr.Suggestion{
					{
						Suggestion: "give this variable an explicit type",
						Example:    "`" + sv.Names[0].Ident + " string = ...`",
					},
				},
			})
		}
		state.Vars[i] = sv
	}

	return errs
}
