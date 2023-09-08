package parsers

import (
	"log"
	"os"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var ImportScript = function.New(&function.Spec{
	Description: "Imports an external script and copies it's contents as a multiline text.",
	Params: []function.Parameter{
		{
			Name:             "str",
			Type:             cty.String,
			AllowDynamicType: true,
		},
	},
	Type: function.StaticReturnType(cty.String),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		// TODO: needs to be passed as relative or it looses it's purpose
		script := args[0].AsString()

		content, err := os.ReadFile(script)
		if err != nil {
			log.Fatal(err)
		}

		return cty.StringVal(string(content)), nil
	},
})
