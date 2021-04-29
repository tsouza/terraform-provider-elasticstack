package provider

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func resourceElasticstackFleetAgentPolicy() *schema.Resource {
	return &schema.Resource{
		/*Create: resourceElasticstackAuthUserCreate,
		Read:   resourceElasticstackAuthUserRead,
		Update: resourceElasticstackAuthUserUpdate,
		Delete: resourceElasticstackAuthUserDelete,*/
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
			},
			"agent_monitoring": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"collect_logs": {
							Type:        schema.TypeBool,
							Description: `Enables logs collection`,
							Default: 	 true,
							Optional:    true,
						},
						"collect_metrics": {
							Type:        schema.TypeBool,
							Description: `Enables metrics collection`,
							Default: 	 true,
							Optional:    true,
						},
					},
				},
			},
		},
	}
}