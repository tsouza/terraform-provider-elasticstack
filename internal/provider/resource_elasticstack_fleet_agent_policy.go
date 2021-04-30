package provider

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElasticstackFleetAgentPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticstackFleetAgentCreate,
		Read:   resourceElasticstackFleetAgentRead,
		Update: resourceElasticstackFleetAgentUpdate,
		Delete: resourceElasticstackFleetAgentDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
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
							Default:     true,
							Optional:    true,
						},
						"collect_metrics": {
							Type:        schema.TypeBool,
							Description: `Enables metrics collection`,
							Default:     true,
							Optional:    true,
						},
					},
				},
			},
			"enrollment_secret": {
				Type:      schema.TypeString,
				Sensitive: true,
				Computed:  true,
			},
		},
	}
}

type KibanaFleetAgentPolicies struct {
	Items []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
}

type KibanaFleetPackagePolicy struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Namespace   string   `json:"namespace"`
	PolicyId    string   `json:"policy_id"`
	Enabled     bool     `json:"enabled"`
	OutputId    string   `json:"output_id"`
	Inputs      []string `json:"inputs"`
	Package     struct {
		Name    string `json:"name"`
		Title   string `json:"title"`
		Version string `json:"version"`
	} `json:"package"`
}

type KibanaEnrollmentKeyList struct {
	List []struct {
		Id       string `json:"id"`
		PolicyId string `json:"policy_id"`
	} `json:"list"`
}

type KibanaEnrollmentKeyDetails struct {
	Item struct {
		ApiKey string `json:"api_key"`
	}
}

// Warning! This function can create the policy but there is no support for updating it yet
func resourceElasticstackFleetAgentCreate(d *schema.ResourceData, meta interface{}) error {
	k := meta.(apiClient).k

	resp, err := k.R().
		SetHeader("Content-Type", "application/json").
		SetBody(`{"forceRecreate": false}`).
		Post("/api/fleet/agents/setup")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("error in agent setup post: %s", resp.Body())
	}

	var agentPolicies KibanaFleetAgentPolicies
	resp, err = k.R().
		SetResult(&agentPolicies).
		Get("/api/fleet/agent_policies")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("error in agent policy get: %s", resp.Body())
	}

	log.Printf("[INFO] agent policies %d, %s", resp.StatusCode(), resp.Body())
	log.Printf("[INFO] parsed agent policies %v", agentPolicies)

	var defaultId string
	for _, i := range agentPolicies.Items {
		if i.Name == "Default policy" {
			defaultId = i.Id
		}
	}

	if defaultId == "" {
		return fmt.Errorf("could not find default agent policy")
	}

	packagePolicy := KibanaFleetPackagePolicy{
		Name:      d.Get("name").(string),
		PolicyId:  defaultId,
		Namespace: "default",
		Enabled:   true,
		Inputs:    []string{},
	}
	packagePolicy.Package.Name = "endpoint"
	packagePolicy.Package.Title = "Endpoint Security"
	packagePolicy.Package.Version = "0.18.0"

	log.Printf("[INFO] package policy %v", packagePolicy)

	resp, err = k.R().
		SetHeader("Content-Type", "application/json").
		SetBody(packagePolicy).
		Post("/api/fleet/package_policies")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		if !strings.Contains(string(resp.Body()), "There is already a package with the same name on this agent policy") {
			// purposely allow this error case for now to support easily re-creating the same integration
			// this is a temporary hack
			return fmt.Errorf("error in package policy post: %s", resp.Body())
		}
	}

	log.Printf("[INFO] package policy response %d, %s", resp.StatusCode(), resp.Body())

	var enrollmentKeyList KibanaEnrollmentKeyList
	resp, err = k.R().
		SetResult(&enrollmentKeyList).
		Get("/api/fleet/enrollment-api-keys")

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("error in enrollment key list get: %s", resp.Body())
	}

	log.Printf("[INFO] enrollment key list %d, %s", resp.StatusCode(), resp.Body())
	log.Printf("[INFO] parsed enrollment key list %v", enrollmentKeyList)

	var enrollmentId string
	for _, e := range enrollmentKeyList.List {
		if e.PolicyId == defaultId {
			enrollmentId = e.Id
		}
	}

	if enrollmentId == "" {
		return fmt.Errorf("could not find enrollment key for default policy")
	}

	var enrollmentKeyDetails KibanaEnrollmentKeyDetails
	resp, err = k.R().
		SetResult(&enrollmentKeyDetails).
		Get("/api/fleet/enrollment-api-keys/" + enrollmentId)

	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return fmt.Errorf("error in enrollment key details get: %s", resp.Body())
	}
	log.Printf("[INFO] enrollment key details %d, %s", resp.StatusCode(), resp.Body())
	log.Printf("[INFO] parsed enrollment key details %v", enrollmentKeyDetails)

	d.Set("enrollment_secret", enrollmentKeyDetails.Item.ApiKey)

	return resourceElasticstackFleetAgentRead(d, meta)
}

func resourceElasticstackFleetAgentRead(d *schema.ResourceData, meta interface{}) error {
	d.SetId(d.Get("name").(string))
	return nil
}

func resourceElasticstackFleetAgentUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticstackFleetAgentCreate(d, meta)
}

func resourceElasticstackFleetAgentDelete(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("delete is not supported")
}
