package file

// Walk walks the passed Scope in depth-first order, calling visit with a
// pointer to each ScopeItem it encounters.
// If visit reports true, Walk will then call Walk on the ScopeItem's Scope, if
// it has one.
//
// If Walk encounters an If, IfBlock, or Switch and visit returns true, Walk
// will Walk all branches (thens) of the conditional.
func Walk(s Scope, visit func(*ScopeItem) bool) {
	_ = WalkError(s, func(itm *ScopeItem) (bool, error) {
		return visit(itm), nil
	})
}

// WalkError is the same as Walk, but adds an error return value to visit.
// If visit returns an error, WalkError will stop traversing the Scope and
// return that error.
func WalkError(s Scope, visit func(*ScopeItem) (bool, error)) error {
	for i, itm := range s {
		switch typedItm := itm.(type) {
		case Block:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Body, visit); err != nil {
					return err
				}
			}
		case Comment:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case Element:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Body, visit); err != nil {
					return err
				}
			}
		case Include:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				corgiIncl, ok := typedItm.Include.(CorgiInclude)
				if ok {
					if err = WalkError(corgiIncl.File.Scope, visit); err != nil {
						return err
					}
				}
			}
		case Code:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case If:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Then, visit); err != nil {
					return err
				}

				for _, ei := range typedItm.ElseIfs {
					if err = WalkError(ei.Then, visit); err != nil {
						return err
					}
				}

				if typedItm.Else != nil {
					if err = WalkError(typedItm.Else.Then, visit); err != nil {
						return err
					}
				}
			}
		case IfBlock:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Then, visit); err != nil {
					return err
				}

				for _, ei := range typedItm.ElseIfs {
					if err = WalkError(ei.Then, visit); err != nil {
						return err
					}
				}

				if typedItm.Else != nil {
					if err = WalkError(typedItm.Else.Then, visit); err != nil {
						return err
					}
				}
			}
		case Switch:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				for _, c := range typedItm.Cases {
					if err = WalkError(c.Then, visit); err != nil {
						return err
					}
				}

				if typedItm.Default != nil {
					if err = WalkError(typedItm.Default.Then, visit); err != nil {
						return err
					}
				}
			}
		case For:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Body, visit); err != nil {
					return err
				}
			}
		case And:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case Text:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case ExpressionInterpolation:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case ElementInterpolation:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case TextInterpolation:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		case Mixin:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Body, visit); err != nil {
					return err
				}
			}
		case MixinCall:
			doVisit, err := visit(&s[i])
			if err != nil {
				return err
			}

			if doVisit {
				if err = WalkError(typedItm.Body, visit); err != nil {
					return err
				}
			}
		case Filter:
			if _, err := visit(&s[i]); err != nil {
				return err
			}
		}
	}

	return nil
}
