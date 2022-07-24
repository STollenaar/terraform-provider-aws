package sts_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// TestAccProfiles defined data resource for the terraform plugin
func TestAccProfiles(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProfiles_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckProfiles("data.aws_list_profiles.profiles"),
				),
			},
		},
	})
}

const testAccProfiles_basic = `
data "aws_list_profiles" "profiles" {}
`
