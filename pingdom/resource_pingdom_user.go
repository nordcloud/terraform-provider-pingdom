package pingdom

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nordcloud/go-pingdom/solarwinds"
	"log"
)

func resourceSolarwindsUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceSolarwindsUserCreate,
		Read:   resourceSolarwindsUserRead,
		Update: resourceSolarwindsUserUpdate,
		Delete: resourceSolarwindsUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"role": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},
			"products": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func userFromResource(d *schema.ResourceData) (*solarwinds.User, error) {
	user := solarwinds.User{}

	// required
	if v, ok := d.GetOk("email"); ok {
		user.Email = v.(string)
	}

	if v, ok := d.GetOk("role"); ok {
		user.Role = v.(string)
	}

	if v, ok := d.GetOk("products"); ok {
		interfaceSlice := v.(*schema.Set).List()
		user.Products = expandUserProducts(interfaceSlice)
	}

	return &user, nil
}

func expandUserProducts(l []interface{}) []solarwinds.Product {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := make([]solarwinds.Product, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		product := solarwinds.Product{}
		if name, ok := tfMap["name"].(string); ok && name != "" {
			product.Name = name
		}
		if role, ok := tfMap["role"].(string); ok && role != "" {
			product.Role = role
		}
		m = append(m, product)
	}

	return m
}

func resourceSolarwindsUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Clients).Solarwinds

	user, err := userFromResource(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] User create configuration: %#v", d.Get("email"))
	err = client.UserService.Create(*user)
	if err != nil {
		return err
	}

	d.SetId(user.Email)
	return nil
}

func resourceSolarwindsUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Clients).Solarwinds

	email := d.Id()
	user, err := client.UserService.Retrieve(email)
	if err != nil {
		return fmt.Errorf("error retrieving user with email %v", email)
	}
	if user == nil {
		d.SetId("")
		return nil
	}
	if err := d.Set("role", user.Role); err != nil {
		return err
	}

	products := schema.NewSet(
		func(product interface{}) int { return String(product.(solarwinds.Product).Name) },
		[]interface{}{},
	)
	for _, product := range user.Products {
		products.Add(product)
	}
	if err := d.Set("products", products); err != nil {
		return err
	}

	return nil
}

func resourceSolarwindsUserUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Clients).Solarwinds

	user, err := userFromResource(d)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] User update configuration: %#v", user)

	if err = client.UserService.Update(*user); err != nil {
		return fmt.Errorf("Error updating user: %s", err)
	}
	return nil
}

func resourceSolarwindsUserDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Clients).Solarwinds

	id := d.Id()
	if err := client.UserService.Delete(id); err != nil {
		return fmt.Errorf("error deleting user: %s", err)
	}

	return nil
}
