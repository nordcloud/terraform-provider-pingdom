module github.com/nordcloud/terraform-provider-pingdom

go 1.15

require (
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/nordcloud/go-pingdom v1.3.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.3
)

replace github.com/nordcloud/go-pingdom => /Users/chszchen/GolandProjects/go-pingdom
