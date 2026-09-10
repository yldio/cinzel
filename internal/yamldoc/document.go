// Copyright 2026 YLD Limited
// SPDX-License-Identifier: Apache-2.0

package yamldoc

// Kind identifies which of a Value's payloads is populated.
type Kind int

const (
	// KindNull is a null value. It encodes as a bare "key:".
	KindNull Kind = iota
	// KindScalar is a string, bool or numeric value.
	KindScalar
	// KindMap is a nested mapping. An empty map encodes as "key: {}".
	KindMap
	// KindSeq is a sequence.
	KindSeq
)

// Value is a single YAML value together with the fidelity facts the encoder
// needs: its kind, and any inline comment that followed it in the source.
type Value struct {
	kind    Kind
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

func newValue(kind Kind, opts []Opt) Value {
	value := Value{kind: kind}

	for _, opt := range opts {
		opt(&value)
	}

	return value
}

// Null returns a null value, encoded as a bare "key:".
func Null(opts ...Opt) Value {
	return newValue(KindNull, opts)
}

// Scalar returns a scalar value. Strings, bools and numeric types are encoded
// with their YAML type preserved; anything else is formatted with %v.
func Scalar(scalar any, opts ...Opt) Value {
	value := newValue(KindScalar, opts)
	value.scalar = scalar

	return value
}

// Map returns a mapping value wrapping doc. A nil or empty doc encodes as
// "key: {}"; a caller that wants a bare "key:" uses Null.
func Map(doc *Doc, opts ...Opt) Value {
	value := newValue(KindMap, opts)

	if doc == nil {
		doc = New()
	}

	value.doc = doc

	return value
}

// Seq returns a sequence value.
func Seq(items []Value, opts ...Opt) Value {
	value := newValue(KindSeq, opts)
	value.seq = items

	return value
}

// Item is one key/value pair in a Doc, in the order it was set.
type Item struct {
	Key   string
	Value Value
}

// Doc is an ordered mapping. Items keep the order in which they were set,
// which for a parsed document is the order they appeared in the source.
type Doc struct {
	items []Item
	index map[string]int
}

// New returns an empty Doc.
func New() *Doc {
	return &Doc{index: map[string]int{}}
}

// Set adds key with the given value, or replaces the value if key is already
// present. Replacing keeps the key's original position.
func (d *Doc) Set(key string, value Value) {
	if d.index == nil {
		d.index = map[string]int{}
	}

	if pos, ok := d.index[key]; ok {
		d.items[pos].Value = value

		return
	}

	d.index[key] = len(d.items)
	d.items = append(d.items, Item{Key: key, Value: value})
}
