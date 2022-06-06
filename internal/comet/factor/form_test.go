package factor

import (
	"fmt"
	"github.com/quanxiang-cloud/entrepot/pkg/client"
	"testing"
)

func TestName(t *testing.T) {

	fmt.Println("111")

	resp := client.TreeResp{
		ID:             "11",
		DepartmentName: "11",
		Child: []client.TreeResp{
			{
				ID:             "22",
				DepartmentName: "22",
				Child: []client.TreeResp{
					{
						ID:             "221",
						DepartmentName: "221",
					},
					{
						ID:             "222",
						DepartmentName: "222",
					},
				},
			},
			{
				ID:             "33",
				DepartmentName: "33",
			},
		},
	}
	resp1 := make(map[string]string)
	root := fmt.Sprintf("/%s", resp.DepartmentName)
	resp1[root] = resp.ID
	for _, value := range resp.Child {
		depTraverse(root, "", value, resp1)
	}
	fmt.Println(resp1)

}
