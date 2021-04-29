package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceElasticstackAuthUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceElasticstackAuthUserCreate,
		Read:   resourceElasticstackAuthUserRead,
		Update: resourceElasticstackAuthUserUpdate,
		Delete: resourceElasticstackAuthUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"full_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metadata": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"password_hash": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"roles": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

type esapiUserData struct {
	Username     string   `json:"username"`
	FullName     string   `json:"full_name"`
	Email        string   `json:"email"`
	Password     string   `json:"password,omitempty"`
	PasswordHash string   `json:"password_hash,omitempty"`
	Roles        []string `json:"roles"`
}

func parseUserResourceData(d *schema.ResourceData) (esapiUserData, error) {
	userData := esapiUserData{
		Username:     d.Get("username").(string),
		FullName:     d.Get("full_name").(string),
		Email:        d.Get("email").(string),
		Password:     d.Get("password").(string),
		PasswordHash: d.Get("password_hash").(string),
		Roles:        expandStringList(d.Get("roles").([]interface{})),
	}

	if userData.Password == "" && userData.PasswordHash == "" {
		return userData, fmt.Errorf("must specify either 'password' or 'password_hash'")
	}
	if userData.Password != "" && userData.PasswordHash != "" {
		return userData, fmt.Errorf("attributes 'password' and 'password_hash' cannot be used in the same resource")
	}

	return userData, nil
}

func resourceElasticstackAuthUserCreate(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*apiClient).es

	userData, err := parseUserResourceData(d)
	if err != nil {
		return err
	}

	bodyJson, err := json.Marshal(userData)
	if err != nil {
		return err
	}

	req := esapi.SecurityPutUserRequest{
		Username: userData.Username,
		Body:     bytes.NewReader(bodyJson),
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	return resourceElasticstackAuthUserRead(d, meta)
}

func resourceElasticstackAuthUserRead(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*apiClient).es

	username := d.Get("username").(string)

	req := esapi.SecurityGetUserRequest{
		Username: []string{username},
	}

	res, err := req.Do(context.Background(), es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("%s", res)
	}

	var userDataList map[string]esapiUserData
	err = json.NewDecoder(res.Body).Decode(&userDataList)
	if err != nil {
		return err
	}

	d.SetId(username)
	userData := userDataList[username]
	d.Set("username", userData.Username)
	d.Set("full_name", userData.FullName)
	d.Set("email", userData.Email)

	var roleList []interface{}
	for _, i := range userData.Roles {
		roleList = append(roleList, i)
	}
	d.Set("roles", roleList)

	return nil
}

func resourceElasticstackAuthUserUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceElasticstackAuthUserCreate(d, meta)
}

func resourceElasticstackAuthUserDelete(d *schema.ResourceData, meta interface{}) error {
	es := meta.(*apiClient).es

	username := d.Get("username").(string)
	req := esapi.SecurityDeleteUserRequest{
		Username: username,
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
