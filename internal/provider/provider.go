package provider

import (
	"context"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/go-resty/resty/v2"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func newSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"elasticsearch_url": {
			Description: "Elasticsearch URL to use for API Authentication.",
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"ELASTICSEARCH_URL", "",
			),
		},
		"kibana_url": {
			Description: "Kibana URL to use for API Authentication.",
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"KIBANA_URL", "",
			),
		},
		"username": {
			Description: "Username to use for API authentication.",
			Type:        schema.TypeString,
			Required:    true,
			DefaultFunc: schema.EnvDefaultFunc(
				"ELASTICSEARCH_USER", "",
			),
		},
		"password": {
			Description: "Password to use for API authentication.",
			Type:        schema.TypeString,
			Required:    true,
			Sensitive:   true,
			DefaultFunc: schema.MultiEnvDefaultFunc(
				[]string{"ELASTICSEARCH_PASS", "ELASTICSEARCH_PASSWORD"}, "",
			),
		},
	}
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: newSchema(),
			ResourcesMap: map[string]*schema.Resource{
				"elasticstack_auth_user":         resourceElasticstackAuthUser(),
				"elasticstack_auth_role":         resourceElasticstackAuthRole(),
				"elasticstack_auth_role_mapping": resourceElasticstackAuthRoleMapping(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

type apiClient struct {
	es	*elasticsearch.Client
	k	*resty.Client
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		var diags diag.Diagnostics

		username := d.Get("username").(string)
		password := d.Get("password").(string)

		es, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{d.Get("elasticsearch_url").(string)},
			Username:  username,
			Password:  password,
		})
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Elasticsearch client",
				Detail:   err.Error(),
			})
		}

		k := resty.New().
			SetHeader("kbn-xsrf", "true").
			SetAuthScheme("Basic").
			SetBasicAuth(username, password).
			SetHostURL(d.Get("kibana_url").(string))

		_, err = k.R().Get("/")
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Kibana client",
				Detail:   err.Error(),
			})
		}

		return apiClient{es, k}, diags
	}
}
