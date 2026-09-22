// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package hclparser

import (
	"errors"

	"github.com/yldio/cinzel/internal/cinzelerror"
)

// Every sentinel here names something the author wrote in their own HCL: an
// index that is not a number, a reference into a value, an operator cinzel
// does not evaluate. New adds the open-an-issue line by default, so without
// the mark a typo in a file they keep asked them to file a bug about it. This
// is what 20f8a92 fixed for .cinzelrc.yaml.
//
// errUnsupportedExpressionType is the exception and is left unmarked: reaching
// it means cinzel met an expression it has no case for, which is cinzel's to
// answer for.
var (
	// errNonNumericIndex is returned when a traversal is indexed with something
	// other than a known number, which is the only thing a list index can be.
	errNonNumericIndex = cinzelerror.UserInput(errors.New("a variable can only be indexed with a number"))

	// errIndexUnsupportedType is returned when a traversal indexes a variable
	// whose type cannot be indexed by position.
	errIndexUnsupportedType = cinzelerror.UserInput(errors.New("only a list or tuple variable can be indexed by position"))

	// errNestedAttribute is returned when a traversal reaches past the variable
	// itself into an attribute of its value, which the lookup cannot follow.
	errNestedAttribute = cinzelerror.UserInput(errors.New("a variable reference cannot reach into an attribute of its value"))

	// errVariableNotFound is returned when a reference names a variable no
	// block declares.
	errVariableNotFound = cinzelerror.UserInput(errors.New("variable does not exist"))

	// errVariableNullOrUnknown is returned when an indexed variable holds no
	// value to index into.
	errVariableNullOrUnknown = cinzelerror.UserInput(errors.New("variable is null or unknown"))

	// errIndexOutOfRange is returned when an index names a position past the
	// end of the list it indexes.
	errIndexOutOfRange = cinzelerror.UserInput(errors.New("index out of range"))

	// errDivisionByZero is returned when the right side of a division
	// evaluates to zero.
	errDivisionByZero = cinzelerror.UserInput(errors.New("division by zero"))

	// errUnsupportedBinaryOperator is returned for an operator cinzel does not
	// evaluate.
	errUnsupportedBinaryOperator = cinzelerror.UserInput(errors.New("unsupported binary operator"))

	// errOperandNotNumber is returned when an arithmetic or ordering operand
	// is not a known number.
	errOperandNotNumber = cinzelerror.UserInput(errors.New("an arithmetic or comparison operator needs a number on both sides"))

	// errUnknownTemplateType is returned when an interpolation resolves to
	// something that cannot be written into a template.
	errUnknownTemplateType = cinzelerror.UserInput(errors.New("an interpolation must resolve to a string, a number or a boolean"))

	// errTypeNotAllowed is returned when a value's type is not one the
	// attribute it was written for accepts.
	errTypeNotAllowed = cinzelerror.UserInput(errors.New("the value written here is not a type this attribute accepts"))

	// errUnsupportedExpressionType is returned when cinzel meets an expression
	// it has no case for. Unmarked: this one is cinzel's.
	errUnsupportedExpressionType = errors.New("unsupported expression type")
)
