package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElasticstackAuthRoleMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticstackAuthRoleMappingCreate,
		Read:   resourceElasticstackAuthRoleMappingRead,
		Update: resourceElasticstackAuthRoleMappingUpdate,
		Delete: resourceElasticstackAuthRoleMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"rule": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"require_all": {
							Type:         schema.TypeBool,
							Optional:     true,
							ExactlyOneOf: []string{"rule.0.require_any"},
						},
						"require_any": {
							Type:         schema.TypeBool,
							Optional:     true,
							ExactlyOneOf: []string{"rule.0.require_all"},
						},
						"field": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

type esapiRoleMappingRule struct {
	Any    []*esapiRoleMappingRule `json:"any,omitempty"`
	All    []*esapiRoleMappingRule `json:"all,omitempty"`
	Except *esapiRoleMappingRule   `json:"except,omitempty"`
	Field  map[string]interface{}  `json:"field,omitempty"`
}

type esapiRoleMappingData struct {
	Name    string                `json:"-"`
	Enabled bool                  `json:"enabled,omitempty"`
	Roles   []string              `json:"roles,omitempty"`
	Rules   *esapiRoleMappingRule `json:"rules,omitempty"`
}

func parseRoleMappingData(d *schema.ResourceData) (esapiRoleMappingData, error) {
	role := esapiRoleMappingData{
		Name:    d.Get("name").(string),
		Enabled: d.Get("enabled").(bool),
		Roles:   expandStringList(d.Get("roles").([]interface{})),
		Rules: &esapiRoleMappingRule{
			Field: map[string]interface{}{},
		},
	}
	rule := d.Get("rule").([]interface{})[0].(map[string]interface{})
	fields := rule["field"].([]interface{})
	fieldRuleList := []*esapiRoleMappingRule{}

	for _, f := range fields {
		fMap := f.(map[string]interface{})
		t := fMap["type"].(string)
		v := fMap["value"].(string)
		if t != "text" {
			return role, fmt.Errorf("field rule type '%s' is not supported", t)
		}
		fieldRuleList = append(fieldRuleList, &esapiRoleMappingRule{
			Field: map[string]interface{}{
				fMap["name"].(string): v,
			},
		})
	}

	if rule["require_all"] != nil && rule["require_all"] == true {
		role.Rules = &esapiRoleMappingRule{
			All: fieldRuleList,
		}
	} else if rule["require_any"] != nil && rule["require_any"] == true {
		role.Rules = &esapiRoleMappingRule{
			Any: fieldRuleList,
		}
	} else {
		return role, fmt.Errorf("neither 'require_all' nor 'require_any' is set")
	}

	return role, nil
}

func resourceElasticstackAuthRoleMappingCreate(d *schema.ResourceData, meta interface{}) error {
	es := meta.(apiClient).es

	roleMappingData, err := parseRoleMappingData(d)
	if err != nil {
		return err
	}

	bodyJson, err := json.Marshal(roleMappingData)
	if err != nil {
		return err
	}

	req := esapi.SecurityPutRoleMappingRequest{
		Name: roleMappingData.Name,
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

	return resourceElasticstackAuthRoleMappingRead(d, meta)
}

func resourceElasticstackAuthRoleMappingRead(d *schema.ResourceData, meta interface{}) error {
	es := meta.(apiClient).es

	name := d.Get("name").(string)

	req := esapi.SecurityGetRoleMappingRequest{
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

	var roleMappingDataList map[string]esapiRoleMappingData
	err = json.NewDecoder(res.Body).Decode(&roleMappingDataList)
	if err != nil {
		return err
	}

	d.SetId(name)
	roleMappingData := roleMappingDataList[name]
	d.Set("name", name)
	d.Set("roles", collapseStringList(roleMappingData.Roles))
	d.Set("enabled", roleMappingData.Enabled)

	if roleMappingData.Rules.Field != nil {
		return fmt.Errorf("role mapping resources with top-level fields are not supported")
	}
	rules := map[string]interface{}{}
	var rulesList []*esapiRoleMappingRule
	if len(roleMappingData.Rules.All) > 0 {
		rules["require_all"] = true
		rulesList = roleMappingData.Rules.All
	} else if len(roleMappingData.Rules.Any) > 0 {
		rules["require_any"] = true
		rulesList = roleMappingData.Rules.Any
	} else {
		return fmt.Errorf("role mapping resource defined neither 'All' nor 'Any' rule array")
	}

	fieldList := []interface{}{}
	for _, r := range rulesList {
		field := map[string]interface{}{
			"type": "text",
		}
		for k := range r.Field {
			field["name"] = k
			field["value"] = r.Field[k]
			break
		}
		fieldList = append(fieldList, field)
	}
	rules["field"] = fieldList
	d.Set("rule", []interface{}{rules})

	return nil
}

func resourceElasticstackAuthRoleMappingUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticstackAuthRoleMappingCreate(d, meta)
}

func resourceElasticstackAuthRoleMappingDelete(d *schema.ResourceData, meta interface{}) error {
	es := meta.(apiClient).es

	name := d.Get("name").(string)
	req := esapi.SecurityDeleteRoleMappingRequest{
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
