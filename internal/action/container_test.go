// Copyright (c) 2024 YLD Limited
// SPDX-License-Identifier: MIT

package action

import (
	"testing"
)

func TestContainer(t *testing.T) {
	// type Test struct {
	// 	name   string
	// 	have   *ContainerConfig
	// 	expect Container
	// }

	// image := hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("node:18"),
	// 		},
	// 	},
	// }

	// ports := hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.NumberIntVal(80),
	// 		},
	// 	},
	// }

	// volumes := hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("my_docker_volume:/volume_mount"),
	// 		},
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("/data/my_data"),
	// 		},
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("/source/directory:/destination/directory"),
	// 		},
	// 	},
	// }

	// options := hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("--cpus 1"),
	// 		},
	// 	},
	// }

	// username := hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("${{ github.actor }}"),
	// 		},
	// 	},
	// }

	// password := hclsyntax.TemplateExpr{
	// 	Parts: []hclsyntax.Expression{
	// 		&hclsyntax.LiteralValueExpr{
	// 			Val: cty.StringVal("${{ secrets.github_token }}"),
	// 		},
	// 	},
	// }

	// var have1 = ContainerConfig{
	// 	Image: &image,
	// 	Credentials: CredentialsConfig{
	// 		Username: &username,
	// 		Password: &password,
	// 	},
	// 	Env: EnvConfig{
	// 		Variable: []VariableConfig{
	// 			{
	// 				Name: "NODE_ENV",
	// 				Value: &hclsyntax.TemplateExpr{
	// 					Parts: []hclsyntax.Expression{
	// 						&hclsyntax.LiteralValueExpr{
	// 							Val: cty.StringVal("development"),
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	Ports:   &ports,
	// 	Volumes: &volumes,
	// 	Options: &options,
	// }
	// var expect1 = Container{
	// 	Image: "node:18",
	// 	Credentials: Credentials{
	// 		Username: "${{ github.actor }}",
	// 		Password: "${{ secrets.github_token }}",
	// 	},
	// 	Env: Env{
	// 		"NODE_ENV": "development",
	// 	},
	// 	Ports: []int32{80},
	// 	Volumes: []string{
	// 		"my_docker_volume:/volume_mount",
	// 		"/data/my_data",
	// 		"/source/directory:/destination/directory",
	// 	},
	// 	Options: "--cpus 1",
	// }

	// var tests = []Test{
	// 	{"with defined container", &have1, expect1},
	// }

	// for _, tt := range tests {
	// 	t.Run(tt.name, func(t *testing.T) {
	// 		got, err := tt.have.Parse()
	// 		if err != nil {
	// 			t.Fatal(err.Error())
	// 		}

	// 		if !reflect.DeepEqual(got, tt.expect) {
	// 			t.Fatal(tt.name)
	// 		}
	// 	})
	// }
}
