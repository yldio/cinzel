package parsers

import (
	"testing"
)

func TestParseYaml(t *testing.T) {
	t.Run("converts a struct to yaml", func(t *testing.T) {
		// 		parse := NewYamlParser()

		// 		workflow := actions.Workflow{
		// 			Name:    "Deploy",
		// 			RunName: "Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}",
		// 			Events: []actions.Event{{
		// 				On: "push",
		// 			}},
		// 			Permissions: []actions.Permission{
		// 				{Perm: "Actions", Value: "read"},
		// 				{Perm: "PullRequests", Value: "write"},
		// 			},
		// 			Envs: []actions.Env{
		// 				{Name: "ENVIRONMENT", Value: "dev"},
		// 			},
		// 			Defaults: actions.Defaults{
		// 				Run: actions.Run{
		// 					Shell:            "bash",
		// 					WorkingDirectory: "./scripts",
		// 				},
		// 			},
		// 			Concurrency: actions.Concurrency{
		// 				Group:            "group-1",
		// 				CancelInProgress: true,
		// 			},
		// 		}

		// 		yaml, err := parse.Do()
		// 		if err != nil {
		// 			t.Errorf(err.Error())
		// 		}

		// 		got, err := parse.ParseYaml(yaml)
		// 		if err != nil {
		// 			t.Errorf(err.Error())
		// 		}

		// 		expected := []byte(`name: Deploy
		// run-name: Deploy to ${{ inputs.deploy_target }} by @${{ github.actor }}
		// "on": push
		// permissions:
		//   actions: read
		//   pull-requests: write
		// env:
		//   ENVIRONMENT: dev
		// defaults:
		//   run:
		//     shell: bash
		//     working-directory: ./scripts
		// concurrency:
		//   group: group-1
		//   cancel-in-progress: true
		// `,
		// 		)

		// 		if !bytes.Equal(got, expected) {
		// 			t.Errorf(err.Error())
		// 		}
	})
}
