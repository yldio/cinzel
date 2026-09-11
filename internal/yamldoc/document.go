// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

// kind identifies which of a Value's payloads is populated.
type kind int

const (
	kindNull kind = iota
	kindScalar
	kindMap
	kindSeq
)

// Value is a single YAML value together with the fidelity facts the encoder
// needs: its kind, and any inline comment that followed it in the source.
type Value struct {
	kind    kind
	scalar  any
	doc     *Doc
	seq     []Value
	comment string
}

// Opt configures a Value at construction.
type Opt func(*Value)

// WithComment attaches an inline comment, emitted after the value on the same
// line. An empty comment is ignored.
func WithComment(comment string) Opt {
	return func(v *Value) {
		v.comment = comment
	}
}

func newValue(k kind, opts []Opt) Value {
	value := Value{kind: k}

	for _, opt := range opts {
		opt(&value)
	}

	return value
}

// Null returns a null value, encoded as a bare "key:".
func Null(opts ...Opt) Value {
	return newValue(kindNull, opts)
}

// Scalar returns a scalar value. Strings, bools and numeric types are encoded
// with their YAML type preserved; anything else is formatted with %v.
func Scalar(scalar any, opts ...Opt) Value {
	value := newValue(kindScalar, opts)
	value.scalar = scalar

	return value
}

// Map returns a mapping value wrapping doc. A nil or empty doc encodes as
// "key: {}"; a caller that wants a bare "key:" uses Null.
func Map(doc *Doc, opts ...Opt) Value {
	value := newValue(kindMap, opts)

	if doc == nil {
		doc = New()
	}

	value.doc = doc

	return value
}

// Seq returns a sequence value.
func Seq(items []Value, opts ...Opt) Value {
	value := newValue(kindSeq, opts)
	value.seq = items

	return value
}

type item struct {
	key   string
	value Value
}

// Doc is an ordered mapping. Items keep the order in which they were set,
// which for a parsed document is the order they appeared in the source.
type Doc struct {
	items []item
}

// New returns an empty Doc.
func New() *Doc {
	return &Doc{}
}

// Set adds key with the given value, or replaces the value if key is already
// present. Replacing keeps the key's original position.
//
// ponytail: linear scan. Documents here hold tens of keys at most; swap in an
// index if one ever holds thousands.
func (d *Doc) Set(key string, value Value) {
	for i := range d.items {
		if d.items[i].key == key {
			d.items[i].value = value

			return
		}
	}

	d.items = append(d.items, item{key: key, value: value})
}
