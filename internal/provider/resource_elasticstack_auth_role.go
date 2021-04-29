package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElasticstackAuthRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticstackAuthRoleCreate,
		Read:   resourceElasticstackAuthRoleRead,
		Update: resourceElasticstackAuthRoleUpdate,
		Delete: resourceElasticstackAuthRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cluster_privileges": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"run_as_privileges": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"index_privilege": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"indices": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"privileges": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"granted_fields": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"denied_fields": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"allow_restricted_indices": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"kibana_privilege": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"spaces": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"privileges": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

type esapiRoleDataIndex struct {
	Names         []string `json:"names"`
	Privileges    []string `json:"privileges"`
	FieldSecurity struct {
		Grant  []string `json:"grant"`
		Except []string `json:"except"`
	} `json:"field_security,omitempty"`
	Query                  string `json:"query"`
	AllowRestrictedIndices bool   `json:"allow_restricted_indices"`
}

type esapiRoleDataApplication struct {
	Application string   `json:"application"`
	Privileges  []string `json:"privileges"`
	Resources   []string `json:"resources"`
}

type esapiRoleData struct {
	Name    string   `json:"-"`
	RunAs   []string `json:"run_as,omitempty"`
	Cluster []string `json:"cluster,omitempty"`
	Global  *struct {
		Application struct {
			Applications []string `json:"applications,omitempty"`
		} `json:"application,omitempty"`
	} `json:"global,omitempty"`
	Indices      []esapiRoleDataIndex       `json:"indices,omitempty"`
	Applications []esapiRoleDataApplication `json:"applications,omitempty"`
}

func parseRoleData(d *schema.ResourceData) (esapiRoleData, error) {
	role := esapiRoleData{
		Name:    d.Get("name").(string),
		RunAs:   expandStringList(d.Get("run_as_privileges").([]interface{})),
		Cluster: expandStringList(d.Get("cluster_privileges").([]interface{})),
	}
	indexPrivileges := d.Get("index_privilege").([]interface{})
	if len(indexPrivileges) > 0 {
		role.Indices = []esapiRoleDataIndex{}
	}
	for _, p := range indexPrivileges {
		privilege := p.(map[string]interface{})
		privilegeStruct := esapiRoleDataIndex{
			Names:                  expandStringList(privilege["indices"].([]interface{})),
			Privileges:             expandStringList(privilege["privileges"].([]interface{})),
			AllowRestrictedIndices: privilege["allow_restricted_indices"].(bool),
		}
		privilegeStruct.FieldSecurity.Grant = expandStringList(privilege["granted_fields"].([]interface{}))
		privilegeStruct.FieldSecurity.Except = expandStringList(privilege["denied_fields"].([]interface{}))
		role.Indices = append(role.Indices, privilegeStruct)
	}
	kibanaPrivileges := d.Get("kibana_privilege").([]interface{})
	if len(kibanaPrivileges) > 0 {
		role.Applications = []esapiRoleDataApplication{}
	}
	for _, p := range kibanaPrivileges {
		privilege := p.(map[string]interface{})
		privilegeStruct := esapiRoleDataApplication{
			Application: "kibana-.kibana",
			Privileges:  expandStringList(privilege["privileges"].([]interface{})),
			Resources:   []string{},
		}
		for _, v := range expandStringList(privilege["spaces"].([]interface{})) {
			if v != "*" {
				v = fmt.Sprintf("space:%s", v)
			}
			privilegeStruct.Resources = append(privilegeStruct.Resources, v)
		}
		role.Applications = append(role.Applications, privilegeStruct)
	}
	return role, nil
}

func resourceElasticstackAuthRoleCreate(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*apiClient).es

	roleData, err := parseRoleData(d)
	if err != nil {
		return err
	}

	bodyJson, err := json.Marshal(roleData)
	if err != nil {
		return err
	}

	req := esapi.SecurityPutRoleRequest{
		Name: roleData.Name,
		Body: bytes.NewReader(bodyJson),
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	return resourceElasticstackAuthRoleRead(d, meta)
}

func resourceElasticstackAuthRoleRead(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*apiClient).es

	name := d.Get("name").(string)

	req := esapi.SecurityGetRoleRequest{
		Name: []string{name},
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	var roleDataList map[string]esapiRoleData
	err = json.NewDecoder(res.Body).Decode(&roleDataList)
	if err != nil {
		return err
	}

	d.SetId(name)
	roleData := roleDataList[name]
	d.Set("name", name)
	d.Set("run_as_privileges", collapseStringList(roleData.RunAs))
	d.Set("cluster_privileges", collapseStringList(roleData.Cluster))

	indices := make([]interface{}, 0)
	for _, i := range roleData.Indices {
		indices = append(indices, map[string]interface{}{
			"indices":                  i.Names,
			"privileges":               i.Privileges,
			"granted_fields":           i.FieldSecurity.Grant,
			"denied_fields":            i.FieldSecurity.Except,
			"allow_restricted_indices": i.AllowRestrictedIndices,
		})
	}
	d.Set("index_privilege", indices)

	spacesRegex := regexp.MustCompile(`^space:(.+)$`)

	kibanaPrivileges := make([]interface{}, 0)
	for _, a := range roleData.Applications {
		spaces := []string{}
		if a.Application != "kibana-.kibana" {
			return fmt.Errorf("the application '%s' is not supported for privilege management", a.Application)
		}
		for _, r := range a.Resources {
			spacesMatches := spacesRegex.FindStringSubmatch(r)
			if spacesMatches != nil {
				spaces = append(spaces, spacesMatches[1])
				continue
			}
			if r == "*" {
				spaces = append(spaces, r)
				continue
			}
			return fmt.Errorf("the resource condition '%s' is not supported for privilege management", r)
		}
		kibanaPrivileges = append(kibanaPrivileges, map[string]interface{}{
			"spaces":     collapseStringList(spaces),
			"privileges": collapseStringList(a.Privileges),
		})
	}
	d.Set("kibana_privilege", kibanaPrivileges)

	return nil
}

func resourceElasticstackAuthRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticstackAuthRoleCreate(d, meta)
}

func resourceElasticstackAuthRoleDelete(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*apiClient).es

	name := d.Get("name").(string)
	req := esapi.SecurityDeleteRoleRequest{
		Name: name,
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	return nil
}
